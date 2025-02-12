// Copyright 2023 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package controllerutils_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/gardener/pkg/api/indexer"
	"github.com/gardener/gardener/pkg/apis/core"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/gardener/gardener/pkg/controllerutils"
)

var _ = Describe("Associations", func() {
	var (
		ctx        context.Context
		fakeClient client.Client

		namespace = "some-namespace"

		quota                  *gardencorev1beta1.Quota
		shoot                  *gardencorev1beta1.Shoot
		backupbucket           *gardencorev1beta1.BackupBucket
		secretBinding          *gardencorev1beta1.SecretBinding
		controllerinstallation *gardencorev1beta1.ControllerInstallation
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeClient = fakeclient.NewClientBuilder().
			WithScheme(kubernetes.GardenScheme).
			WithIndex(&gardencorev1beta1.BackupBucket{}, core.BackupBucketSeedName, indexer.BackupBucketSeedNameIndexerFunc).
			WithIndex(&gardencorev1beta1.ControllerInstallation{}, core.SeedRefName, indexer.ControllerInstallationSeedRefNameIndexerFunc).
			Build()

		shoot = &gardencorev1beta1.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "shoot",
				Namespace: namespace,
			},
		}

		secretBinding = &gardencorev1beta1.SecretBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secretbinding",
				Namespace: namespace,
			},
		}
	})

	DescribeTable("#DetermineShootsAssociatedTo",
		func(obj client.Object, mutateFunc func(shoot *gardencorev1beta1.Shoot, obj client.Object), errorMatcher gomegatypes.GomegaMatcher) {
			mutateFunc(shoot, obj)
			Expect(fakeClient.Create(ctx, shoot)).To(Succeed())

			shoots, err := DetermineShootsAssociatedTo(ctx, fakeClient, obj)
			Expect(err).To(errorMatcher)

			if err == nil {
				Expect(shoots).To(HaveLen(1))
				Expect(shoots).To(ConsistOf(shoot.Namespace + "/" + shoot.Name))
			} else {
				Expect(shoots).To(BeEmpty())
			}
		},

		Entry("should return shoots associated to cloudprofile",
			&gardencorev1beta1.CloudProfile{ObjectMeta: metav1.ObjectMeta{Name: "cloudprofile"}}, func(s *gardencorev1beta1.Shoot, obj client.Object) {
				s.Spec.CloudProfileName = obj.GetName()
			}, BeNil()),
		Entry("should return shoots associated to seed",
			&gardencorev1beta1.Seed{ObjectMeta: metav1.ObjectMeta{Name: "seed"}}, func(s *gardencorev1beta1.Shoot, obj client.Object) {
				s.Spec.SeedName = pointer.String(obj.GetName())
			}, BeNil()),
		Entry("should return shoots associated to secretbinding",
			&gardencorev1beta1.SecretBinding{ObjectMeta: metav1.ObjectMeta{Name: "secretbinding", Namespace: namespace}}, func(s *gardencorev1beta1.Shoot, obj client.Object) {
				s.Spec.SecretBindingName = pointer.String(obj.GetName())
			}, BeNil()),
		Entry("should return shoots associated to exposureclass",
			&gardencorev1beta1.ExposureClass{ObjectMeta: metav1.ObjectMeta{Name: "exposureclass"}}, func(s *gardencorev1beta1.Shoot, obj client.Object) {
				s.Spec.ExposureClassName = pointer.String(obj.GetName())
			}, BeNil()),
		Entry("should return error if the object is of not supported type",
			&gardencorev1beta1.BackupBucket{ObjectMeta: metav1.ObjectMeta{Name: "backupbucket"}}, func(s *gardencorev1beta1.Shoot, obj client.Object) {}, HaveOccurred()),
	)

	Describe("#DetermineSecretBindingAssociations", func() {
		It("should return secretBinding associated to quota", func() {
			quota = &gardencorev1beta1.Quota{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "quota",
					Namespace: namespace,
				},
			}

			secretBinding.Quotas = []corev1.ObjectReference{{Name: quota.Name, Namespace: quota.Namespace}}
			Expect(fakeClient.Create(ctx, secretBinding)).To(Succeed())

			secretBindings, err := DetermineSecretBindingAssociations(ctx, fakeClient, quota)
			Expect(err).ToNot(HaveOccurred())
			Expect(secretBindings).To(HaveLen(1))
			Expect(secretBindings).To(ConsistOf(secretBinding.Namespace + "/" + secretBinding.Name))
		})
	})

	Describe("#DetermineBackupBucketAssociations", func() {
		It("should return backupbucket associated to seed", func() {
			backupbucket = &gardencorev1beta1.BackupBucket{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backupbucket",
				},
				Spec: gardencorev1beta1.BackupBucketSpec{
					SeedName: pointer.String("test"),
				},
			}

			Expect(fakeClient.Create(ctx, backupbucket)).To(Succeed())

			backupbuckets, err := DetermineBackupBucketAssociations(ctx, fakeClient, "test")
			Expect(err).ToNot(HaveOccurred())
			Expect(backupbuckets).To(HaveLen(1))
			Expect(backupbuckets).To(ConsistOf(backupbucket.Name))
		})
	})

	Describe("#DetermineControllerInstallationAssociations", func() {
		It("should return controllerinstallation associated to seed", func() {
			controllerinstallation = &gardencorev1beta1.ControllerInstallation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "controllerinstallation",
				},
				Spec: gardencorev1beta1.ControllerInstallationSpec{
					SeedRef: corev1.ObjectReference{Name: "test"},
				},
			}

			Expect(fakeClient.Create(ctx, controllerinstallation)).To(Succeed())

			controllerinstallations, err := DetermineControllerInstallationAssociations(ctx, fakeClient, "test")
			Expect(err).ToNot(HaveOccurred())
			Expect(controllerinstallations).To(HaveLen(1))
			Expect(controllerinstallations).To(ConsistOf(controllerinstallation.Name))
		})
	})
})
