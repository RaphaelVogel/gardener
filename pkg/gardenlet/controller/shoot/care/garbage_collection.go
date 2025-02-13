// Copyright 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package care

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/operation"
	"github.com/gardener/gardener/pkg/operation/shoot"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
)

// GarbageCollection contains required information for shoot and seed garbage collection.
type GarbageCollection struct {
	initializeShootClients ShootClientInit
	shoot                  *shoot.Shoot
	seedClient             client.Client
	log                    logr.Logger
}

// NewGarbageCollection creates a new garbage collection instance.
func NewGarbageCollection(op *operation.Operation, shootClientInit ShootClientInit) *GarbageCollection {
	return &GarbageCollection{
		shoot:                  op.Shoot,
		initializeShootClients: shootClientInit,
		seedClient:             op.SeedClientSet.Client(),
		log:                    op.Logger,
	}
}

// Collect cleans the Seed and the Shoot cluster from no longer required
// objects. It receives a botanist object <botanist> which stores the Shoot object.
func (g *GarbageCollection) Collect(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := g.performGarbageCollectionSeed(ctx); err != nil {
			g.log.Error(err, "Error during seed garbage collection")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		shootClient, apiServerRunning, err := g.initializeShootClients()
		if err != nil || !apiServerRunning {
			if err != nil {
				g.log.Error(err, "Could not initialize Shoot client for garbage collection")
			}
			return
		}
		if err := g.performGarbageCollectionShoot(ctx, shootClient.Client()); err != nil {
			g.log.Error(err, "Error during shoot garbage collection")
		}
	}()

	wg.Wait()
	g.log.V(1).Info("Successfully performed full garbage collection")
}

// PerformGarbageCollectionSeed performs garbage collection in the Shoot namespace in the Seed cluster
func (g *GarbageCollection) performGarbageCollectionSeed(ctx context.Context) error {
	return g.deleteStalePods(ctx, g.seedClient, g.shoot.SeedNamespace)
}

// PerformGarbageCollectionShoot performs garbage collection in the kube-system namespace in the Shoot
// cluster, i.e., it deletes evicted pods (mitigation for https://github.com/kubernetes/kubernetes/issues/55051).
func (g *GarbageCollection) performGarbageCollectionShoot(ctx context.Context, shootClient client.Client) error {
	if err := g.deleteOrphanedNodeLeases(ctx, shootClient); err != nil {
		return fmt.Errorf("failed deleting orphaned node lease objects: %w", err)
	}

	namespace := metav1.NamespaceSystem
	if g.shoot.GetInfo().DeletionTimestamp != nil {
		namespace = metav1.NamespaceAll
	}

	return g.deleteStalePods(ctx, shootClient, namespace)
}

// See https://github.com/gardener/gardener/issues/8749 and https://github.com/kubernetes/kubernetes/issues/109777.
// kubelet sometimes created Lease objects without owner reference. When the respective node gets deleted eventually,
// the Lease object remains in the system and no Kubernetes controller will ever clean it up. Hence, this function takes
// over this task.
// TODO: Remove this function when support for Kubernetes 1.28 is dropped.
func (g *GarbageCollection) deleteOrphanedNodeLeases(ctx context.Context, c client.Client) error {
	leaseList := &coordinationv1.LeaseList{}
	if err := c.List(ctx, leaseList, client.InNamespace(corev1.NamespaceNodeLease)); err != nil {
		return err
	}

	var orphanedLeases []client.Object

	for _, l := range leaseList.Items {
		if len(l.OwnerReferences) > 0 {
			continue
		}
		lease := l.DeepCopy()

		if err := c.Get(ctx, client.ObjectKey{Name: lease.Name}, &metav1.PartialObjectMetadata{TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Node"}}); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("failed getting node %s when checking for potential orphaned Lease %s: %w", lease.Name, client.ObjectKeyFromObject(lease), err)
			}

			g.log.Info("Detected orphaned Lease object, cleaning it up", "nodeName", lease.Name, "lease", client.ObjectKeyFromObject(lease))
			orphanedLeases = append(orphanedLeases, lease)
		}
	}

	return kubernetesutils.DeleteObjects(ctx, c, orphanedLeases...)
}

// GardenerDeletionGracePeriod is the default grace period for Gardener's force deletion methods.
const GardenerDeletionGracePeriod = 5 * time.Minute

func (g *GarbageCollection) deleteStalePods(ctx context.Context, c client.Client, namespace string) error {
	podList := &corev1.PodList{}
	if err := c.List(ctx, podList, client.InNamespace(namespace)); err != nil {
		return err
	}

	var result error

	for _, pod := range podList.Items {
		log := g.log.WithValues("pod", client.ObjectKeyFromObject(&pod))

		if strings.Contains(pod.Status.Reason, "Evicted") || strings.HasPrefix(pod.Status.Reason, "OutOf") {
			log.V(1).Info("Deleting pod", "reason", pod.Status.Reason)
			if err := c.Delete(ctx, &pod, kubernetes.DefaultDeleteOptions...); client.IgnoreNotFound(err) != nil {
				result = multierror.Append(result, err)
			}
			continue
		}

		if shouldObjectBeRemoved(&pod, GardenerDeletionGracePeriod) {
			g.log.V(1).Info("Deleting stuck terminating pod")
			if err := c.Delete(ctx, &pod, kubernetes.ForceDeleteOptions...); client.IgnoreNotFound(err) != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return result
}

// shouldObjectBeRemoved determines whether the given object should be gone now.
// This is calculated by first checking the deletion timestamp of an object: If the deletion timestamp
// is unset, the object should not be removed - i.e. this returns false.
// Otherwise, it is checked whether the deletionTimestamp is before the current time minus the
// grace period.
func shouldObjectBeRemoved(obj metav1.Object, gracePeriod time.Duration) bool {
	deletionTimestamp := obj.GetDeletionTimestamp()
	if deletionTimestamp == nil {
		return false
	}

	return deletionTimestamp.Time.Before(time.Now().Add(-gracePeriod))
}
