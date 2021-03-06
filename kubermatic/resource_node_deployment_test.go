package kubermatic

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/kubermatic/go-kubermatic/client/project"
	"github.com/kubermatic/go-kubermatic/models"
)

const testKubeletVersion16 = "1.16.10"
const testKubeletVersion17 = "1.17.6"

func TestAccKubermaticNodeDeployment_Openstack_Basic(t *testing.T) {
	var ndepl models.NodeDeployment
	testName := randomTestName()

	username := os.Getenv(testEnvOpenstackUsername)
	password := os.Getenv(testEnvOpenstackPassword)
	tenant := os.Getenv(testEnvOpenstackTenant)
	nodeDC := os.Getenv(testEnvOpenstackNodeDC)
	image := os.Getenv(testEnvOpenstackImage)
	image2 := os.Getenv(testEnvOpenstackImage2)
	flavor := os.Getenv(testEnvOpenstackFlavor)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckForOpenstack(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubermaticNodeDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckKubermaticNodeDeploymentBasic(testName, nodeDC, username, password, tenant, testClusterVersion17, testKubeletVersion16, image, flavor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubermaticNodeDeploymentExists("kubermatic_node_deployment.acctest_nd", &ndepl),
					testAccCheckKubermaticNodeDeploymentFields(&ndepl, flavor, image, testKubeletVersion16, 1, 0, false, false),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "name", testName),
					resource.TestCheckResourceAttrPtr("kubermatic_node_deployment.acctest_nd", "name", &ndepl.Name),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.replicas", "1"),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.openstack.0.flavor", flavor),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.openstack.0.image", image),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.operating_system.0.ubuntu.#", "1"),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.versions.0.kubelet", testKubeletVersion16),
				),
			},
			{
				Config: testAccCheckKubermaticNodeDeploymentBasic2(testName, nodeDC, username, password, tenant, testClusterVersion17, testKubeletVersion17, image2, flavor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testResourceInstanceState("kubermatic_node_deployment.acctest_nd", func(is *terraform.InstanceState) error {
						_, _, _, id, err := kubermaticNodeDeploymentParseID(is.ID)
						if err != nil {
							return err
						}
						if id != ndepl.ID {
							return fmt.Errorf("node deployment not updated. Want ID=%v, got %v", ndepl.ID, id)
						}
						return nil
					}),
					testAccCheckKubermaticNodeDeploymentExists("kubermatic_node_deployment.acctest_nd", &ndepl),
					testAccCheckKubermaticNodeDeploymentFields(&ndepl, flavor, image2, testKubeletVersion17, 2, 123, true, true),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "name", testName),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.replicas", "2"),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.openstack.0.flavor", flavor),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.openstack.0.image", image2),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.openstack.0.use_floating_ip", "true"),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.openstack.0.disk_size", "123"),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.operating_system.0.ubuntu.0.dist_upgrade_on_boot", "true"),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.versions.0.kubelet", testKubeletVersion17),
				),
			},
		},
	})
}

func testAccCheckKubermaticNodeDeploymentBasic(testName, nodeDC, username, password, tenant, clusterVersion, kubeletVersion, image, flavor string) string {
	return fmt.Sprintf(`
	resource "kubermatic_project" "acctest_project" {
		name = "%s"
	}

	resource "kubermatic_cluster" "acctest_cluster" {
		name = "%s"
		dc_name = "%s"
		project_id = kubermatic_project.acctest_project.id
		spec {
			version = "%s"
			cloud {
				openstack {
					tenant = "%s"
					username = "%s"
					password = "%s"
					floating_ip_pool = "ext-net"
				}
			}
		}
	}

	resource "kubermatic_node_deployment" "acctest_nd" {
		cluster_id = kubermatic_cluster.acctest_cluster.id
		name = "%s"
		spec {
			replicas = 1
			template {
				cloud {
					openstack {
						flavor = "%s"
						image = "%s"
					}
				}
				operating_system {
					ubuntu {}
				}
				versions {
					kubelet = "%s"
				}
			}
		}
	}`, testName, testName, nodeDC, clusterVersion, tenant, username, password, testName, flavor, image, kubeletVersion)
}

func testAccCheckKubermaticNodeDeploymentBasic2(testName, nodeDC, username, password, tenant, clusterVersion, kubeletVersion, image, flavor string) string {
	return fmt.Sprintf(`
	resource "kubermatic_project" "acctest_project" {
		name = "%s"
		labels = {
			"project-label" = "val"
		}
	}

	resource "kubermatic_cluster" "acctest_cluster" {
		name = "%s"
		dc_name = "%s"
		project_id = kubermatic_project.acctest_project.id
		labels = {
			"cluster-label" = "val"
		}
		spec {
			version = "%s"
			cloud {
				openstack {
					tenant = "%s"
					username = "%s"
					password = "%s"
					floating_ip_pool = "ext-net"
				}
			}
		}
	}

	resource "kubermatic_node_deployment" "acctest_nd" {
		cluster_id = kubermatic_cluster.acctest_cluster.id
		name = "%s"
		spec {
			replicas = 2
			template {
				labels = {
					"foo" = "bar"
				}
				cloud {
					openstack {
						flavor = "%s"
						image = "%s"
						use_floating_ip = true
						disk_size = 123
					}
				}
				operating_system {
					ubuntu {
						dist_upgrade_on_boot = true
					}
				}
				versions {
					kubelet = "%s"
				}
			}
		}
	}`, testName, testName, nodeDC, clusterVersion, tenant, username, password, testName, flavor, image, kubeletVersion)
}

func testAccCheckKubermaticNodeDeploymentDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckKubermaticNodeDeploymentFields(rec *models.NodeDeployment, flavor, image, kubeletVersion string, replicas, diskSize int, floatingIP, distUpgrade bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if rec == nil {
			return fmt.Errorf("No Record")
		}

		if rec.Spec == nil || rec.Spec.Template == nil || rec.Spec.Template.Cloud == nil || rec.Spec.Template.Cloud.Openstack == nil {
			return fmt.Errorf("No Openstack cloud spec present")
		}

		openstack := rec.Spec.Template.Cloud.Openstack

		if openstack.Flavor == nil {
			return fmt.Errorf("No Flavor spec present")
		}
		if *openstack.Flavor != flavor {
			return fmt.Errorf("Flavor=%s, want %s", *openstack.Flavor, flavor)
		}

		if openstack.Image == nil {
			return fmt.Errorf("No Image spec present")
		}

		if *openstack.Image != image {
			return fmt.Errorf("Image=%s, want %s", *openstack.Image, image)
		}

		if openstack.RootDiskSizeGB != int64(diskSize) {
			return fmt.Errorf("want RootDiskSizeGB=%d, got %d", openstack.RootDiskSizeGB, diskSize)
		}

		if openstack.UseFloatingIP != floatingIP {
			return fmt.Errorf("want UseFloatingIP=%v, got %v", openstack.UseFloatingIP, floatingIP)
		}

		opSys := rec.Spec.Template.OperatingSystem
		if opSys == nil {
			return fmt.Errorf("No OperatingSystem spec present")
		}

		ubuntu := opSys.Ubuntu
		if ubuntu == nil {
			return fmt.Errorf("No Ubuntu spec present")
		}

		if ubuntu.DistUpgradeOnBoot != distUpgrade {
			return fmt.Errorf("want Ubuntu.DistUpgradeOnBoot=%v, got %v", ubuntu.DistUpgradeOnBoot, distUpgrade)
		}

		versions := rec.Spec.Template.Versions
		if versions == nil {
			return fmt.Errorf("No Versions")
		}

		if versions.Kubelet != kubeletVersion {
			return fmt.Errorf("Versions.Kubelet=%s, want %s", versions.Kubelet, kubeletVersion)
		}

		if rec.Spec.Replicas == nil || *rec.Spec.Replicas != int32(replicas) {
			return fmt.Errorf("Replicas=%d, want %d", rec.Spec.Replicas, replicas)
		}

		return nil
	}
}

func TestAccKubermaticNodeDeployment_Azure_Basic(t *testing.T) {
	var nodedepl models.NodeDeployment
	testName := randomTestName()

	clientID := os.Getenv(testEnvAzureClientID)
	clientSecret := os.Getenv(testEnvAzureClientSecret)
	tenantID := os.Getenv(testEnvAzureTenantID)
	subsID := os.Getenv(testEnvAzureSubscriptionID)
	nodeDC := os.Getenv(testEnvAzureNodeDC)
	nodeSize := os.Getenv(testEnvAzureNodeSize)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckForAzure(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubermaticNodeDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckKubermaticNodeDeploymentAzureBasic(testName, clientID, clientSecret, tenantID, subsID, nodeDC, nodeSize),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubermaticNodeDeploymentExists("kubermatic_node_deployment.acctest_nd", &nodedepl),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.azure.0.size", nodeSize),
				),
			},
		},
	})
}

func testAccCheckKubermaticNodeDeploymentAzureBasic(n, clientID, clientSecret, tenantID, subscID, nodeDC, nodeSize string) string {
	return fmt.Sprintf(`
	resource "kubermatic_project" "acctest_project" {
		name = "%s"
	}

	resource "kubermatic_cluster" "acctest_cluster" {
		name = "%s"
		dc_name = "%s"
		project_id = kubermatic_project.acctest_project.id

		spec {
			version = "1.17.6"
			cloud {
				azure {
					client_id = "%s"
					client_secret = "%s"
					tenant_id = "%s"
					subscription_id = "%s"
				}
			}
		}
	}

	resource "kubermatic_node_deployment" "acctest_nd" {
		cluster_id = kubermatic_cluster.acctest_cluster.id
		name = "%s"
		spec {
			replicas = 2
			template {
				cloud {
					azure {
						size = "%s"
					}
				}
				operating_system {
					ubuntu {
						dist_upgrade_on_boot = false
					}
				}
				versions {
					kubelet = "1.17.6"
				}
			}
		}
	}`, n, n, nodeDC, clientID, clientSecret, tenantID, subscID, n, nodeSize)
}

func TestAccKubermaticNodeDeployment_AWS_Basic(t *testing.T) {
	var nodedepl models.NodeDeployment
	testName := randomTestName()

	accessKeyID := os.Getenv(testEnvAWSAccessKeyID)
	accessKeySecret := os.Getenv(testAWSSecretAccessKey)
	vpcID := os.Getenv(testEnvAWSVPCID)
	nodeDC := os.Getenv(testEnvAWSNodeDC)
	instanceType := os.Getenv(testEnvAWSInstanceType)
	subnetID := os.Getenv(testEnvAWSSubnetID)
	availabilityZone := os.Getenv(testEnvAWSAvailabilityZone)
	diskSize := os.Getenv(testEnvAWSDiskSize)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckForAWS(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubermaticNodeDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckKubermaticNodeDeploymentAWSBasic(testName, accessKeyID, accessKeySecret, vpcID, nodeDC, instanceType, subnetID, availabilityZone, diskSize),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubermaticNodeDeploymentExists("kubermatic_node_deployment.acctest_nd", &nodedepl),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.instance_type", instanceType),
					resource.TestCheckResourceAttrPtr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.instance_type", nodedepl.Spec.Template.Cloud.Aws.InstanceType),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.disk_size", diskSize),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.volume_type", "standart"),
					resource.TestCheckResourceAttrPtr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.volume_type", nodedepl.Spec.Template.Cloud.Aws.VolumeType),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.subnet_id", subnetID),
					resource.TestCheckResourceAttrPtr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.subnet_id", &nodedepl.Spec.Template.Cloud.Aws.SubnetID),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.availability_zone", availabilityZone),
					resource.TestCheckResourceAttrPtr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.availability_zone", &nodedepl.Spec.Template.Cloud.Aws.AvailabilityZone),
					resource.TestCheckResourceAttr("kubermatic_node_deployment.acctest_nd", "spec.0.template.0.cloud.0.aws.0.assign_public_ip", "true"),
				),
			},
		},
	})
}

func testAccCheckKubermaticNodeDeploymentAWSBasic(n, keyID, keySecret, vpcID, nodeDC, instanceType, subnetID, availabilityZone, diskSize string) string {
	return fmt.Sprintf(`
	resource "kubermatic_project" "acctest_project" {
		name = "%s"
	}

	resource "kubermatic_cluster" "acctest_cluster" {
		name = "%s"
		dc_name = "%s"
		project_id = kubermatic_project.acctest_project.id

		spec {
			version = "1.17.6"
			cloud {
				aws {
					access_key_id = "%s"
					access_key_secret = "%s"
					vpc_id = "%s'
				}
			}
		}
	}

	resource "kubermatic_node_deployment" "acctest_nd" {
		cluster_id = kubermatic_cluster.acctest_cluster.id
		name = "%s"
		spec {
			replicas = 2
			template {
				cloud {
					aws {
						instance_type = "%s"
						disk_size = "%s"
						volume_type = "standart"
						subnet_id = "%s"
						availability_zone = "%s"
						assign_public_ip = true
					}
				}
				operating_system {
					ubuntu {
						dist_upgrade_on_boot = false
					}
				}
				versions {
					kubelet = "1.17.6"
				}
			}
		}
	}`, n, n, nodeDC, keyID, keySecret, vpcID, n, instanceType, diskSize, subnetID, availabilityZone)
}

func testAccCheckKubermaticNodeDeploymentExists(n string, rec *models.NodeDeployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		k := testAccProvider.Meta().(*kubermaticProviderMeta)

		projectID, seedDC, clusterID, nodeDeplID, err := kubermaticNodeDeploymentParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		p := project.NewGetNodeDeploymentParams()
		p.SetProjectID(projectID)
		p.SetClusterID(clusterID)
		p.SetDC(seedDC)
		p.SetNodeDeploymentID(nodeDeplID)
		r, err := k.client.Project.GetNodeDeployment(p, k.auth)
		if err != nil {
			return fmt.Errorf("GetNodeDeployment: %v", err)
		}
		*rec = *r.Payload

		return nil
	}
}
