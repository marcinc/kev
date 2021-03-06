/**
 * Copyright 2020 Appvia Ltd <info@appvia.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kev_test

import (
	"github.com/appvia/kev/pkg/kev"
	"github.com/appvia/kev/pkg/kev/config"
	"github.com/appvia/kev/pkg/kev/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Reconcile", func() {
	var (
		hook          *test.Hook
		loggedMsgs    string
		workingDir    string
		source        *kev.ComposeProject
		overrideFiles []string
		override      *kev.ComposeProject
		manifest      *kev.Manifest
		env           *kev.Environment
		mErr          error
	)

	JustBeforeEach(func() {
		if len(overrideFiles) > 0 {
			override, _ = kev.NewComposeProject(overrideFiles)
		}
		hook = testutil.NewLogger(logrus.DebugLevel)
		manifest, mErr = kev.Reconcile(workingDir)
		if mErr == nil {
			env, _ = manifest.GetEnvironment("dev")
		}
		loggedMsgs = testutil.GetLoggedMsgs(hook)
	})

	JustAfterEach(func() {
		hook.Reset()
	})

	Describe("Reconcile changes from overrides", func() {
		When("the override version has been updated", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-override-rollback"
				source, _ = kev.NewComposeProject([]string{workingDir + "/docker-compose.yaml"})
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			It("confirms the version pre reconciliation", func() {
				Expect(override.GetVersion()).NotTo(Equal(source.GetVersion()))
			})

			It("should roll back the change", func() {
				Expect(env.GetVersion()).To(Equal(source.GetVersion()))
			})
		})

		Context("for services and volumes", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-override-keep"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}

			})

			When("the override service label overrides have been updated", func() {
				It("confirms the values pre reconciliation", func() {
					s, _ := override.GetService("db")
					Expect(s.Labels["kev.workload.cpu"]).To(Equal("0.5"))
					Expect(s.Labels["kev.workload.max-cpu"]).To(Equal("0.75"))
					Expect(s.Labels["kev.workload.memory"]).To(Equal("50Mi"))
					Expect(s.Labels["kev.workload.replicas"]).To(Equal("5"))
					Expect(s.Labels["kev.workload.service-account-name"]).To(Equal("overridden-service-account-name"))
				})

				It("keeps overridden override values", func() {
					s, _ := env.GetService("db")
					Expect(s.Labels["kev.workload.cpu"]).To(Equal("0.5"))
					Expect(s.Labels["kev.workload.max-cpu"]).To(Equal("0.75"))
					Expect(s.Labels["kev.workload.memory"]).To(Equal("50Mi"))
					Expect(s.Labels["kev.workload.replicas"]).To(Equal("5"))
					Expect(s.Labels["kev.workload.service-account-name"]).To(Equal("overridden-service-account-name"))
				})
			})

			When("the override volume label overrides have been updated", func() {
				It("confirms the values pre reconciliation", func() {
					v, _ := override.Volumes["db_data"]
					Expect(v.Labels["kev.volume.size"]).To(Equal("200Mi"))
				})

				It("keeps overridden override values", func() {
					v, _ := env.GetVolume("db_data")
					Expect(v.Labels["kev.volume.size"]).To(Equal("200Mi"))
				})
			})
		})
	})

	Describe("Reconciling changes from sources", func() {

		Context("when the compose version has been updated", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-version"
				source, _ = kev.NewComposeProject([]string{workingDir + "/docker-compose.yaml"})
			})

			It("should update all environments with the new version", func() {
				Expect(env.GetVersion()).To(Equal(source.GetVersion()))
			})

			It("should log the change summary using the debug level", func() {
				Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
			})

			It("should create a change summary", func() {
				Expect(loggedMsgs).To(ContainSubstring(env.Name))
				Expect(loggedMsgs).To(ContainSubstring(env.Name))
				Expect(loggedMsgs).To(ContainSubstring("3.7"))
				Expect(loggedMsgs).To(ContainSubstring(env.GetVersion()))
			})

			It("should not error", func() {
				Expect(mErr).NotTo(HaveOccurred())
			})
		})

		Context("when a compose service has been removed", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-service-removal"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			It("confirms the number of services pre reconciliation", func() {
				Expect(override.Services).To(HaveLen(2))
				Expect(override.ServiceNames()).To(ContainElements("db", "wordpress"))
			})

			It("should remove the service from all environments", func() {
				Expect(env.GetServices()).To(HaveLen(1))
				Expect(env.GetServices()[0].Name).To(Equal("db"))
			})

			It("should log the change summary using the debug level", func() {
				Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
			})

			It("should create a change summary", func() {
				Expect(loggedMsgs).To(ContainSubstring(env.Name))
				Expect(loggedMsgs).To(ContainSubstring("wordpress"))
				Expect(loggedMsgs).To(ContainSubstring("deleted"))
			})

			It("should not error", func() {
				Expect(mErr).NotTo(HaveOccurred())
			})
		})

		Context("when the compose service is edited", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-service-edit"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			Context("and it changes port mode to host", func() {
				It("confirms the edit pre reconciliation", func() {
					s, _ := override.GetService("wordpress")
					Expect(s.Labels["kev.service.type"]).To(Equal("LoadBalancer"))
				})

				It("should not update the label in all environments", func() {
					s, _ := env.GetService("wordpress")
					Expect(s.Labels["kev.service.type"]).To(Equal("LoadBalancer"))
				})

				It("should log the change summary using the debug level", func() {
					Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
				})

				It("should create a change summary", func() {
					Expect(loggedMsgs).To(ContainSubstring(env.Name))
					Expect(loggedMsgs).To(ContainSubstring("nothing to update"))
				})

				It("should not error", func() {
					Expect(mErr).NotTo(HaveOccurred())
				})
			})
		})

		Context("when a new compose service has been added", func() {
			Context("and the service has no deploy or healthcheck config", func() {
				BeforeEach(func() {
					workingDir = "testdata/reconcile-service-basic"
					overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
				})

				It("confirms the number of services pre reconciliation", func() {
					Expect(override.Services).To(HaveLen(1))
					Expect(override.ServiceNames()).To(ContainElements("db"))
				})

				It("should add the new service labels to all environments", func() {
					Expect(env.GetServices()).To(HaveLen(2))
					Expect(env.GetServices()[0].Name).To(Equal("db"))
					Expect(env.GetServices()[1].Name).To(Equal("wordpress"))
				})

				It("should configure the added service labels with defaults", func() {
					expected := newDefaultServiceLabels("wordpress")
					Expect(env.GetServices()[1].GetLabels()).To(Equal(expected))
				})

				It("should not include any env vars", func() {
					Expect(env.GetServices()[1].Environment).To(HaveLen(0))
				})

				It("should log the change summary using the debug level", func() {
					Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
				})

				It("should create a change summary", func() {
					Expect(loggedMsgs).To(ContainSubstring(env.Name))
					Expect(loggedMsgs).To(ContainSubstring("added"))
					Expect(loggedMsgs).To(ContainSubstring("wordpress"))
				})

				It("should not error", func() {
					Expect(mErr).NotTo(HaveOccurred())
				})
			})

			Context("and the service has deploy config", func() {
				BeforeEach(func() {
					workingDir = "testdata/reconcile-service-deploy"
				})

				It("should configure the added service labels from deploy config", func() {
					expected := newDefaultServiceLabels("wordpress")
					expected[config.LabelWorkloadReplicas] = "3"
					Expect(env.GetServices()[1].GetLabels()).To(Equal(expected))
				})
			})

			Context("and the service has healthcheck config", func() {
				BeforeEach(func() {
					workingDir = "testdata/reconcile-service-healthcheck"
				})

				It("should configure the added service labels from healthcheck config", func() {
					expected := newDefaultServiceLabels("wordpress")
					expected[config.LabelWorkloadLivenessProbeCommand] = "[\"CMD\", \"curl\", \"localhost:80/healthy\"]"
					Expect(env.GetServices()[1].GetLabels()).To(Equal(expected))
				})
			})
		})

		Context("when a new compose volume has been added", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-volume-add"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			It("confirms the number of volumes pre reconciliation", func() {
				Expect(override.Volumes).To(HaveLen(0))
			})

			It("should add the new volume labels to all environments", func() {
				Expect(env.GetVolumes()).To(HaveLen(1))

				v, _ := env.GetVolume("db_data")
				Expect(v.Labels).To(HaveLen(1))
				Expect(v.Labels[config.LabelVolumeSize]).To(Equal("100Mi"))
			})

			It("should log the change summary using the debug level", func() {
				Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
			})

			It("should create a change summary", func() {
				Expect(loggedMsgs).To(ContainSubstring(env.Name))
				Expect(loggedMsgs).To(ContainSubstring("added"))
				Expect(loggedMsgs).To(ContainSubstring("db_data"))
			})

			It("should not error", func() {
				Expect(mErr).NotTo(HaveOccurred())
			})
		})

		Context("when a compose volume has been removed", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-volume-removal"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			It("confirms the number of volumes pre reconciliation", func() {
				Expect(override.Volumes).To(HaveLen(1))
			})

			It("should remove the volume from all environments", func() {
				Expect(env.GetVolumes()).To(HaveLen(0))
			})

			It("should log the change summary using the debug level", func() {
				Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
			})

			It("should create a change summary", func() {
				Expect(loggedMsgs).To(ContainSubstring(env.Name))
				Expect(loggedMsgs).To(ContainSubstring("deleted"))
				Expect(loggedMsgs).To(ContainSubstring("db_data"))
			})

			It("should not error", func() {
				Expect(mErr).NotTo(HaveOccurred())
			})
		})

		Context("when a compose volume has been edited", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-volume-edit"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			It("confirms the volume name pre reconciliation", func() {
				Expect(override.VolumeNames()).To(HaveLen(1))
				Expect(override.VolumeNames()).To(ContainElements("db_data"))
			})

			It("should update the volume in all environments", func() {
				Expect(env.VolumeNames()).To(HaveLen(1))
				Expect(env.VolumeNames()).To(ContainElements("mysql_data"))
			})

			It("should log the change summary using the debug level", func() {
				Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
			})

			It("should create a change summary", func() {
				Expect(loggedMsgs).To(ContainSubstring(env.Name))
				Expect(loggedMsgs).To(ContainSubstring("deleted"))
				Expect(loggedMsgs).To(ContainSubstring("db_data"))
				Expect(loggedMsgs).To(ContainSubstring("added"))
				Expect(loggedMsgs).To(ContainSubstring("mysql_data"))
			})

			It("should not error", func() {
				Expect(mErr).NotTo(HaveOccurred())
			})
		})

		Context("when compose env vars have been removed", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-env-var-removal"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			It("confirms the env vars pre reconciliation", func() {
				s, _ := override.GetService("wordpress")
				Expect(s.Environment).To(HaveLen(2))
				Expect(s.Environment).To(HaveKey("WORDPRESS_CACHE_USER"))
				Expect(s.Environment).To(HaveKey("WORDPRESS_CACHE_PASSWORD"))
			})

			It("should remove the env vars from all environments", func() {
				vars, _ := env.GetEnvVarsForService("wordpress")
				Expect(vars).To(HaveLen(0))
			})

			It("confirms environment labels post reconciliation", func() {
				s, _ := env.GetService("wordpress")
				Expect(s.GetLabels()).To(HaveLen(16))
			})

			It("should log the change summary using the debug level", func() {
				Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
			})

			It("should create a change summary", func() {
				Expect(loggedMsgs).To(ContainSubstring(env.Name))
				Expect(loggedMsgs).To(ContainSubstring("deleted"))
				Expect(loggedMsgs).To(ContainSubstring("WORDPRESS_CACHE_USER"))
				Expect(loggedMsgs).To(ContainSubstring("WORDPRESS_CACHE_PASSWORD"))
			})

			It("should not error", func() {
				Expect(mErr).NotTo(HaveOccurred())
			})
		})

		Context("when compose env var is overridden in an environment", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-env-var-override"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			It("confirms the overridden env var pre reconciliation", func() {
				s, _ := override.GetService("db")
				Expect(s.Environment).To(HaveLen(1))
				Expect(s.Environment).To(HaveKey("TO_OVERRIDE"))
			})

			It("should keep the overridden env var in all environments", func() {
				vars, _ := env.GetEnvVarsForService("db")
				Expect(vars).To(HaveLen(1))
				Expect(vars).To(HaveKey("TO_OVERRIDE"))
			})

			It("should log the change summary using the debug level", func() {
				Expect(testutil.GetLoggedLevel(hook)).To(Equal("debug"))
			})

			It("should create a change summary", func() {
				Expect(loggedMsgs).To(ContainSubstring(env.Name))
				Expect(loggedMsgs).To(ContainSubstring("nothing to update"))
			})

			It("should not error", func() {
				Expect(mErr).NotTo(HaveOccurred())
			})
		})

		Context("when compose or override env var is not assigned a value", func() {
			BeforeEach(func() {
				workingDir = "testdata/reconcile-env-var-unassigned"
				overrideFiles = []string{workingDir + "/docker-compose.kev.dev.yaml"}
			})

			It("should not error", func() {
				Expect(mErr).NotTo(HaveOccurred())
			})
		})
	})
})

func newDefaultServiceLabels(name string) map[string]string {
	return map[string]string{
		config.LabelWorkloadLivenessProbeCommand: "[\"CMD\", \"echo\", \"Define healthcheck command for service " + name + "\"]",
		config.LabelWorkloadReplicas:             "1",
	}
}
