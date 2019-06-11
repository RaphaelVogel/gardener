// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package networkpolicies

import (
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	// OpenStackMetadataServiceHost points to openstack-specific Metadata service.
	OpenStackMetadataServiceHost = &Host{
		Description: "Metadata service",
		HostName:    "169.254.169.254",
		Port:        80,
	}

	// OpenStackCloudControllerManagerNotSecured points to OpenStack specific cloud-controller-manager running on HTTP port.
	OpenStackCloudControllerManagerNotSecured = &SourcePod{
		Ports: NewSinglePort(10253),
		Pod: NewPod("cloud-controller-manager-http", labels.Set{
			"app":                     "kubernetes",
			"garden.sapcloud.io/role": "controlplane",
			"role":                    "cloud-controller-manager",
		}, "< 1.13"),
		ExpectedPolicies: sets.NewString(
			"allow-from-prometheus",
			"allow-to-dns",
			"allow-to-private-networks",
			"allow-to-public-networks",
			"allow-to-shoot-apiserver",
			"deny-all",
		),
	}

	// OpenStackCloudControllerManagerSecured points to OpenStack specific cloud-controller-manager running on HTTPS port.
	OpenStackCloudControllerManagerSecured = &SourcePod{
		Ports: NewSinglePort(10258),
		Pod: NewPod("cloud-controller-manager-https", labels.Set{
			"app":                     "kubernetes",
			"garden.sapcloud.io/role": "controlplane",
			"role":                    "cloud-controller-manager",
		}, ">= 1.13"),
		ExpectedPolicies: sets.NewString(
			"allow-from-prometheus",
			"allow-to-dns",
			"allow-to-private-networks",
			"allow-to-public-networks",
			"allow-to-shoot-apiserver",
			"deny-all",
		),
	}
)

// OpenStackNetworkPolicy holds openstack-specific network policy settings.
// +gen-netpoltests=true
// +gen-packagename=openstack
type OpenStackNetworkPolicy struct {
}

// ToSources returns list of all openstack-specific sources and targets.
func (a *OpenStackNetworkPolicy) ToSources() []Rule {

	return []Rule{
		a.newSource(KubeAPIServerInfo).AllowPod(EtcdMainInfo, EtcdEventsInfo).AllowHost(SeedKubeAPIServer, ExternalHost).Build(),
		a.newSource(EtcdMainInfo).AllowHost(ExternalHost).Build(),
		a.newSource(EtcdEventsInfo).AllowHost(ExternalHost).Build(),
		a.newSource(OpenStackCloudControllerManagerNotSecured).AllowPod(KubeAPIServerInfo).AllowHost(ExternalHost).Build(),
		a.newSource(OpenStackCloudControllerManagerSecured).AllowPod(KubeAPIServerInfo).AllowHost(ExternalHost).Build(),
		a.newSource(DependencyWatchdog).AllowHost(SeedKubeAPIServer).Build(),
		a.newSource(ElasticSearchInfo).Build(),
		a.newSource(GrafanaInfo).AllowPod(PrometheusInfo).Build(),
		a.newSource(KibanaInfo).AllowTargetPod(ElasticSearchInfo.FromPort("http")).Build(),
		a.newSource(AddonManagerInfo).AllowPod(KubeAPIServerInfo).Build(),
		a.newSource(KubeControllerManagerInfoNotSecured).AllowPod(KubeAPIServerInfo).AllowHost(OpenStackMetadataServiceHost, ExternalHost).Build(),
		a.newSource(KubeControllerManagerInfoSecured).AllowPod(KubeAPIServerInfo).AllowHost(OpenStackMetadataServiceHost, ExternalHost).Build(),
		a.newSource(KubeSchedulerInfoNotSecured).AllowPod(KubeAPIServerInfo).Build(),
		a.newSource(KubeSchedulerInfoSecured).AllowPod(KubeAPIServerInfo).Build(),
		a.newSource(KubeStateMetricsShootInfo).AllowPod(KubeAPIServerInfo).Build(),
		a.newSource(KubeStateMetricsSeedInfo).AllowHost(SeedKubeAPIServer, ExternalHost).Build(),
		a.newSource(MachineControllerManagerInfo).AllowPod(KubeAPIServerInfo).AllowHost(SeedKubeAPIServer, ExternalHost).Build(),
		a.newSource(PrometheusInfo).AllowPod(
			OpenStackCloudControllerManagerNotSecured,
			OpenStackCloudControllerManagerSecured,
			EtcdEventsInfo,
			EtcdMainInfo,
			KubeAPIServerInfo,
			KubeControllerManagerInfoNotSecured,
			KubeControllerManagerInfoSecured,
			KubeSchedulerInfoNotSecured,
			KubeSchedulerInfoSecured,
			KubeStateMetricsSeedInfo,
			KubeStateMetricsShootInfo,
			MachineControllerManagerInfo,
		).AllowTargetPod(ElasticSearchInfo.FromPort("metrics")).AllowHost(SeedKubeAPIServer, ExternalHost, GardenPrometheus).Build(),
	}
}

// EgressFromOtherNamespaces returns list of all openstack-specific sources and targets.
func (a *OpenStackNetworkPolicy) EgressFromOtherNamespaces(sourcePod *SourcePod) Rule {
	return NewSource(sourcePod).DenyPod(a.allPods()...).AllowPod(KubeAPIServerInfo).Build()
}

func (a *OpenStackNetworkPolicy) newSource(sourcePod *SourcePod) *RuleBuilder {
	return NewSource(sourcePod).DenyPod(a.allPods()...).DenyHost(OpenStackMetadataServiceHost, ExternalHost, GardenPrometheus)
}

func (a *OpenStackNetworkPolicy) allPods() []*SourcePod {
	return []*SourcePod{
		AddonManagerInfo,
		DependencyWatchdog,
		OpenStackCloudControllerManagerNotSecured,
		OpenStackCloudControllerManagerSecured,
		ElasticSearchInfo,
		EtcdEventsInfo,
		EtcdMainInfo,
		GrafanaInfo,
		KibanaInfo,
		KubeAPIServerInfo,
		KubeControllerManagerInfoNotSecured,
		KubeControllerManagerInfoSecured,
		KubeSchedulerInfoNotSecured,
		KubeSchedulerInfoSecured,
		KubeStateMetricsSeedInfo,
		KubeStateMetricsShootInfo,
		MachineControllerManagerInfo,
		PrometheusInfo,
	}
}

// Provider returns OpenStack cloud provider.
func (a *OpenStackNetworkPolicy) Provider() v1beta1.CloudProvider {
	return v1beta1.CloudProviderOpenStack
}
