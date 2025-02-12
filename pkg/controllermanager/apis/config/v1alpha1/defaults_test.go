// Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package v1alpha1_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/pointer"

	. "github.com/gardener/gardener/pkg/controllermanager/apis/config/v1alpha1"
	"github.com/gardener/gardener/pkg/logger"
)

var _ = Describe("Defaults", func() {
	var obj *ControllerManagerConfiguration

	BeforeEach(func() {
		obj = &ControllerManagerConfiguration{}
	})

	Describe("ControllerManagerConfiguration defaulting", func() {
		It("should default ControllerManagerConfiguration correctly", func() {
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.LogLevel).To(Equal(logger.InfoLevel))
			Expect(obj.LogFormat).To(Equal(logger.FormatJSON))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				LogLevel:  "warning",
				LogFormat: "md",
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.LogLevel).To(Equal("warning"))
			Expect(obj.LogFormat).To(Equal("md"))
		})
	})

	Describe("ClientConnectionConfiguration defaulting", func() {
		It("should default ClientConnectionConfiguration correctly", func() {
			expected := &componentbaseconfigv1alpha1.ClientConnectionConfiguration{
				QPS:   50.0,
				Burst: 100,
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(&obj.GardenClientConnection).To(Equal(expected))
		})

		It("should not default ContentType and AcceptContentTypes", func() {
			SetObjectDefaults_ControllerManagerConfiguration(obj)
			// ContentType fields will be defaulted by client constructors / controller-runtime based on whether a
			// given APIGroup supports protobuf or not. defaults must not touch these, otherwise the intelligent
			// logic will be overwritten
			Expect(obj.GardenClientConnection.ContentType).To(BeEmpty())
			Expect(obj.GardenClientConnection.AcceptContentTypes).To(BeEmpty())
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				GardenClientConnection: componentbaseconfigv1alpha1.ClientConnectionConfiguration{
					QPS:   60.0,
					Burst: 120,
				},
			}
			expected := obj.GardenClientConnection.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(&obj.GardenClientConnection).To(Equal(expected))
		})
	})

	Describe("LeaderElectionConfiguration defaulting", func() {
		It("should default LeaderElectionConfiguration correctly", func() {
			expected := &componentbaseconfigv1alpha1.LeaderElectionConfiguration{
				LeaderElect:       pointer.Bool(true),
				ResourceLock:      "leases",
				RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
				RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
				LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
				ResourceNamespace: "garden",
				ResourceName:      "gardener-controller-manager-leader-election",
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.LeaderElection).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				LeaderElection: &componentbaseconfigv1alpha1.LeaderElectionConfiguration{
					LeaderElect:       pointer.Bool(true),
					ResourceLock:      "foo",
					RetryPeriod:       metav1.Duration{Duration: 40 * time.Second},
					RenewDeadline:     metav1.Duration{Duration: 41 * time.Second},
					LeaseDuration:     metav1.Duration{Duration: 42 * time.Second},
					ResourceNamespace: "other-garden-ns",
					ResourceName:      "lock-object",
				},
			}
			expected := obj.LeaderElection.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.LeaderElection).To(Equal(expected))
		})
	})

	Describe("ShootRetryControllerConfiguration defaulting", func() {
		It("should default ShootRetryControllerConfiguration correctly", func() {
			expected := &ShootRetryControllerConfiguration{
				ConcurrentSyncs:   pointer.Int(DefaultControllerConcurrentSyncs),
				RetryPeriod:       &metav1.Duration{Duration: 10 * time.Minute},
				RetryJitterPeriod: &metav1.Duration{Duration: 5 * time.Minute},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootRetry).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ShootRetry: &ShootRetryControllerConfiguration{
						ConcurrentSyncs:   pointer.Int(10),
						RetryPeriod:       &metav1.Duration{Duration: 12 * time.Minute},
						RetryJitterPeriod: &metav1.Duration{Duration: 8 * time.Minute},
					},
				},
			}
			expected := obj.Controllers.ShootRetry.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootRetry).To(Equal(expected))
		})
	})

	Describe("SeedControllerConfiguration defaulting", func() {
		It("should default SeedControllerConfiguration correctly", func() {
			expected := &SeedControllerConfiguration{
				ConcurrentSyncs:    pointer.Int(DefaultControllerConcurrentSyncs),
				SyncPeriod:         &metav1.Duration{Duration: 10 * time.Second},
				MonitorPeriod:      &metav1.Duration{Duration: 40 * time.Second},
				ShootMonitorPeriod: &metav1.Duration{Duration: 5 * 40 * time.Second},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Seed).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					Seed: &SeedControllerConfiguration{
						ConcurrentSyncs:    pointer.Int(10),
						SyncPeriod:         &metav1.Duration{Duration: 12 * time.Second},
						MonitorPeriod:      &metav1.Duration{Duration: 42 * time.Second},
						ShootMonitorPeriod: &metav1.Duration{Duration: 6 * 42 * time.Second},
					},
				},
			}
			expected := obj.Controllers.Seed.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Seed).To(Equal(expected))
		})
	})

	Describe("ProjectControllerConfiguration defaulting", func() {
		It("should default ProjectControllerConfiguration correctly", func() {
			expected := &ProjectControllerConfiguration{
				ConcurrentSyncs:         pointer.Int(DefaultControllerConcurrentSyncs),
				MinimumLifetimeDays:     pointer.Int(30),
				StaleGracePeriodDays:    pointer.Int(14),
				StaleExpirationTimeDays: pointer.Int(90),
				StaleSyncPeriod: &metav1.Duration{
					Duration: 12 * time.Hour,
				},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Project).To(Equal(expected))
		})

		It("should default ProjectControllerConfiguration unset QuotaConfiguration correctly", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					Project: &ProjectControllerConfiguration{
						Quotas: []QuotaConfiguration{
							{},
							{ProjectSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}}},
							{},
						},
					},
				},
			}
			expected := &ProjectControllerConfiguration{
				Quotas: []QuotaConfiguration{
					{ProjectSelector: &metav1.LabelSelector{}},
					{ProjectSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}}},
					{ProjectSelector: &metav1.LabelSelector{}},
				},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Project.Quotas).To(Equal(expected.Quotas))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					Project: &ProjectControllerConfiguration{
						ConcurrentSyncs:         pointer.Int(20),
						MinimumLifetimeDays:     pointer.Int(40),
						StaleGracePeriodDays:    pointer.Int(24),
						StaleExpirationTimeDays: pointer.Int(100),
						StaleSyncPeriod: &metav1.Duration{
							Duration: 12 * time.Hour,
						},
					},
				},
			}
			expected := obj.Controllers.Project.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Project).To(Equal(expected))
		})
	})

	Describe("ServerConfiguration defaulting", func() {
		It("should default ServerConfiguration correctly", func() {
			expected := &ServerConfiguration{
				HealthProbes: &Server{
					Port: 2718,
				},
				Metrics: &Server{
					Port: 2719,
				},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(&obj.Server).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Server: ServerConfiguration{
					HealthProbes: &Server{
						Port: 3000,
					},
					Metrics: &Server{
						Port: 4000,
					},
				},
			}
			expected := obj.Server.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(&obj.Server).To(Equal(expected))
		})
	})

	Describe("BastionControllerConfiguration defaulting", func() {
		It("should default BastionControllerConfiguration correctly", func() {
			expected := &BastionControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
				MaxLifetime:     &metav1.Duration{Duration: 24 * time.Hour},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Bastion).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					Bastion: &BastionControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
						MaxLifetime:     &metav1.Duration{Duration: 48 * time.Hour},
					},
				},
			}
			expected := obj.Controllers.Bastion.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Bastion).To(Equal(expected))
		})
	})

	Describe("CertificateSigningRequestControllerConfiguration defaulting", func() {
		It("should default CertificateSigningRequestControllerConfiguration correctly", func() {
			expected := &CertificateSigningRequestControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.CertificateSigningRequest).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					CertificateSigningRequest: &CertificateSigningRequestControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.CertificateSigningRequest.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.CertificateSigningRequest).To(Equal(expected))
		})
	})

	Describe("CloudProfileControllerConfiguration defaulting", func() {
		It("should default CloudProfileControllerConfiguration correctly", func() {
			expected := &CloudProfileControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.CloudProfile).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					CloudProfile: &CloudProfileControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.CloudProfile.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.CloudProfile).To(Equal(expected))
		})
	})

	Describe("ControllerDeploymentControllerConfiguration defaulting", func() {
		It("should default ControllerDeploymentControllerConfiguration correctly", func() {
			expected := &ControllerDeploymentControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ControllerDeployment).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ControllerDeployment: &ControllerDeploymentControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.ControllerDeployment.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ControllerDeployment).To(Equal(expected))
		})
	})

	Describe("ControllerRegistrationControllerConfiguration defaulting", func() {
		It("should default ControllerRegistrationControllerConfiguration correctly", func() {
			expected := &ControllerRegistrationControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ControllerRegistration).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ControllerRegistration: &ControllerRegistrationControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.ControllerRegistration.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ControllerRegistration).To(Equal(expected))
		})
	})

	Describe("ExposureClassControllerConfiguration defaulting", func() {
		It("should default ExposureClassControllerConfiguration correctly", func() {
			expected := &ExposureClassControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ExposureClass).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ExposureClass: &ExposureClassControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.ExposureClass.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ExposureClass).To(Equal(expected))
		})
	})

	Describe("QuotaControllerConfiguration defaulting", func() {
		It("should default QuotaControllerConfiguration correctly", func() {
			expected := &QuotaControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Quota).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					Quota: &QuotaControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.Quota.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Quota).To(Equal(expected))
		})
	})

	Describe("SecretBindingControllerConfiguration defaulting", func() {
		It("should default SecretBindingControllerConfiguration correctly", func() {
			expected := &SecretBindingControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.SecretBinding).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					SecretBinding: &SecretBindingControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.SecretBinding.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.SecretBinding).To(Equal(expected))
		})
	})

	Describe("SeedExtensionsCheckControllerConfiguration defaulting", func() {
		It("should default SeedExtensionsCheckControllerConfiguration correctly", func() {
			expected := &SeedExtensionsCheckControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
				SyncPeriod:      &metav1.Duration{Duration: 30 * time.Second},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.SeedExtensionsCheck).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					SeedExtensionsCheck: &SeedExtensionsCheckControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
						SyncPeriod:      &metav1.Duration{Duration: 60 * time.Second},
					},
				},
			}
			expected := obj.Controllers.SeedExtensionsCheck.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.SeedExtensionsCheck).To(Equal(expected))
		})
	})

	Describe("SeedBackupBucketsCheckControllerConfiguration defaulting", func() {
		It("should default SeedBackupBucketsCheckControllerConfiguration correctly", func() {
			expected := &SeedBackupBucketsCheckControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
				SyncPeriod:      &metav1.Duration{Duration: 30 * time.Second},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.SeedBackupBucketsCheck).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					SeedBackupBucketsCheck: &SeedBackupBucketsCheckControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
						SyncPeriod:      &metav1.Duration{Duration: 60 * time.Second},
					},
				},
			}
			expected := obj.Controllers.SeedBackupBucketsCheck.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.SeedBackupBucketsCheck).To(Equal(expected))
		})
	})

	Describe("ShootHibernationControllerConfiguration defaulting", func() {
		It("should default ShootHibernationControllerConfiguration correctly", func() {
			expected := &ShootHibernationControllerConfiguration{
				ConcurrentSyncs:         pointer.Int(DefaultControllerConcurrentSyncs),
				TriggerDeadlineDuration: &metav1.Duration{Duration: 2 * time.Hour},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(&obj.Controllers.ShootHibernation).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ShootHibernation: ShootHibernationControllerConfiguration{
						ConcurrentSyncs:         pointer.Int(10),
						TriggerDeadlineDuration: &metav1.Duration{Duration: 3 * time.Hour},
					},
				},
			}
			expected := obj.Controllers.ShootHibernation.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(&obj.Controllers.ShootHibernation).To(Equal(expected))
		})
	})

	Describe("ShootMaintenanceControllerConfiguration defaulting", func() {
		It("should default ShootMaintenanceControllerConfiguration correctly", func() {
			expected := &ShootMaintenanceControllerConfiguration{
				ConcurrentSyncs:                  pointer.Int(DefaultControllerConcurrentSyncs),
				EnableShootControlPlaneRestarter: pointer.Bool(true),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(&obj.Controllers.ShootMaintenance).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ShootMaintenance: ShootMaintenanceControllerConfiguration{
						ConcurrentSyncs:                  pointer.Int(10),
						EnableShootControlPlaneRestarter: pointer.Bool(false),
					},
				},
			}
			expected := obj.Controllers.ShootMaintenance.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(&obj.Controllers.ShootMaintenance).To(Equal(expected))
		})
	})

	Describe("ShootQuotaControllerConfiguration defaulting", func() {
		It("should default ShootQuotaControllerConfiguration correctly", func() {
			expected := &ShootQuotaControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
				SyncPeriod: &metav1.Duration{
					Duration: 60 * time.Minute,
				},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootQuota).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ShootQuota: &ShootQuotaControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
						SyncPeriod: &metav1.Duration{
							Duration: 120 * time.Minute,
						},
					},
				},
			}
			expected := obj.Controllers.ShootQuota.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootQuota).To(Equal(expected))
		})
	})

	Describe("ShootReferenceControllerConfiguration defaulting", func() {
		It("should default ShootReferenceControllerConfiguration correctly", func() {
			expected := &ShootReferenceControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootReference).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ShootReference: &ShootReferenceControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.ShootReference.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootReference).To(Equal(expected))
		})
	})

	Describe("ShootConditionsControllerConfiguration defaulting", func() {
		It("should default ShootConditionsControllerConfiguration correctly", func() {
			expected := &ShootConditionsControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootConditions).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ShootConditions: &ShootConditionsControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.ShootConditions.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootConditions).To(Equal(expected))
		})
	})

	Describe("EventControllerConfiguration defaulting", func() {
		It("should default EventControllerConfiguration correctly if set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					Event: &EventControllerConfiguration{},
				},
			}
			expected := &EventControllerConfiguration{
				ConcurrentSyncs:   pointer.Int(DefaultControllerConcurrentSyncs),
				TTLNonShootEvents: &metav1.Duration{Duration: 1 * time.Hour},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Event).To(Equal(expected))
		})

		It("should not default EventControllerConfiguration if not set", func() {
			var expected *EventControllerConfiguration
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Event).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					Event: &EventControllerConfiguration{
						ConcurrentSyncs:   pointer.Int(10),
						TTLNonShootEvents: &metav1.Duration{Duration: 2 * time.Hour},
					},
				},
			}
			expected := obj.Controllers.Event.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.Event).To(Equal(expected))
		})
	})

	Describe("ShootStatusLabelControllerConfiguration defaulting", func() {
		It("should default ShootStatusLabelControllerConfiguration correctly", func() {
			expected := &ShootStatusLabelControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootStatusLabel).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ShootStatusLabel: &ShootStatusLabelControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
					},
				},
			}
			expected := obj.Controllers.ShootStatusLabel.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ShootStatusLabel).To(Equal(expected))
		})
	})

	Describe("ManagedSeedSetControllerConfiguration defaulting", func() {
		It("should default ManagedSeedSetControllerConfiguration correctly if nil", func() {
			expected := &ManagedSeedSetControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
				MaxShootRetries: pointer.Int(3),
				SyncPeriod: metav1.Duration{
					Duration: 30 * time.Minute,
				},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ManagedSeedSet).To(Equal(expected))
		})

		It("should default ManagedSeedSetControllerConfiguration correctly if not nil", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ManagedSeedSet: &ManagedSeedSetControllerConfiguration{
						SyncPeriod: metav1.Duration{
							Duration: 20 * time.Minute,
						},
					},
				},
			}
			expected := &ManagedSeedSetControllerConfiguration{
				ConcurrentSyncs: pointer.Int(DefaultControllerConcurrentSyncs),
				MaxShootRetries: pointer.Int(3),
				SyncPeriod: metav1.Duration{
					Duration: 20 * time.Minute,
				},
			}
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ManagedSeedSet).To(Equal(expected))
		})

		It("should not default fields that are set", func() {
			obj = &ControllerManagerConfiguration{
				Controllers: ControllerManagerControllerConfiguration{
					ManagedSeedSet: &ManagedSeedSetControllerConfiguration{
						ConcurrentSyncs: pointer.Int(10),
						MaxShootRetries: pointer.Int(5),
						SyncPeriod: metav1.Duration{
							Duration: 10 * time.Minute,
						},
					},
				},
			}
			expected := obj.Controllers.ManagedSeedSet.DeepCopy()
			SetObjectDefaults_ControllerManagerConfiguration(obj)

			Expect(obj.Controllers.ManagedSeedSet).To(Equal(expected))
		})
	})

})

var _ = Describe("Constants", func() {
	It("should have the same values as the corresponding constants in the logger package", func() {
		Expect(LogLevelDebug).To(Equal(logger.DebugLevel))
		Expect(LogLevelInfo).To(Equal(logger.InfoLevel))
		Expect(LogLevelError).To(Equal(logger.ErrorLevel))
		Expect(LogFormatJSON).To(Equal(logger.FormatJSON))
		Expect(LogFormatText).To(Equal(logger.FormatText))
	})
})
