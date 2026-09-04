package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccTaskFileTransferResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-ftp")
	resourceName := "stonebranch_task_file_transfer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccTaskFileTransferConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "sys_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			// ImportState
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
			},
			// Update
			{
				Config: testAccTaskFileTransferConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "summary", "Updated file transfer task"),
				),
			},
		},
	})
}

func TestAccTaskFileTransferResource_withSummary(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-ftp")
	resourceName := "stonebranch_task_file_transfer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskFileTransferConfig_withSummary(rName, "Initial summary"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "summary", "Initial summary"),
				),
			},
			{
				Config: testAccTaskFileTransferConfig_withSummary(rName, "Changed summary"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "summary", "Changed summary"),
				),
			},
		},
	})
}

func TestAccTaskFileTransferResource_udm(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-udm")
	resourceName := "stonebranch_task_file_transfer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccTaskFileTransferConfig_udm(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "server_type", "UDM"),
					resource.TestCheckResourceAttr(resourceName, "primary_broker_choice", "Agent"),
					resource.TestCheckResourceAttr(resourceName, "primary_broker", "udm-agent-01"),
					resource.TestCheckResourceAttr(resourceName, "secondary_broker_choice", "Agent Cluster"),
					resource.TestCheckResourceAttr(resourceName, "secondary_cluster_ref", "Opswise - Default Linux/Unix Cluster"),
					resource.TestCheckResourceAttr(resourceName, "command", "GET"),
					resource.TestCheckResourceAttr(resourceName, "udm_operation", "Copy"),
					resource.TestCheckResourceAttr(resourceName, "format", "Binary"),
					resource.TestCheckResourceAttr(resourceName, "form_or_script", "Form"),
					resource.TestCheckResourceAttrSet(resourceName, "sys_id"),
				),
			},
			// ImportState
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
			},
			// Update
			{
				Config: testAccTaskFileTransferConfig_udmUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "remote_filename", "/remote/outgoing/report-updated.csv"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccTaskFileTransferConfig_basic(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_file_transfer" "test" {
  name                   = %[1]q
  agent_var              = "agent_name"
  remote_server          = "test.example.com"
  remote_filename        = "/remote/path/file.txt"
  local_filename         = "/local/path/file.txt"
  remote_credentials_var = "ftp_credentials"
}
`, name)
}

func testAccTaskFileTransferConfig_updated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_file_transfer" "test" {
  name                   = %[1]q
  summary                = "Updated file transfer task"
  agent_var              = "agent_name"
  remote_server          = "test.example.com"
  remote_filename        = "/remote/path/file.txt"
  local_filename         = "/local/path/file.txt"
  remote_credentials_var = "ftp_credentials"
}
`, name)
}

func testAccTaskFileTransferConfig_withSummary(name, summary string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_file_transfer" "test" {
  name                   = %[1]q
  summary                = %[2]q
  agent_var              = "agent_name"
  remote_server          = "test.example.com"
  remote_filename        = "/remote/path/file.txt"
  local_filename         = "/local/path/file.txt"
  remote_credentials_var = "ftp_credentials"
}
`, name, summary)
}

func testAccTaskFileTransferConfig_udm(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_file_transfer" "test" {
  name        = %[1]q
  server_type = "UDM"

  agent_cluster = "Opswise - Default Linux/Unix Cluster"
  exit_codes    = "0"

  primary_broker_choice   = "Agent"
  primary_broker          = "udm-agent-01"
  secondary_broker_choice = "Agent Cluster"
  secondary_cluster_ref   = "Opswise - Default Linux/Unix Cluster"

  local_filename  = "/data/incoming"
  remote_filename = "/remote/outgoing/report.csv"

  command       = "GET"
  udm_operation = "Copy"
  format        = "Binary"

  form_or_script = "Form"
}
`, name)
}

func testAccTaskFileTransferConfig_udmUpdated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_file_transfer" "test" {
  name        = %[1]q
  server_type = "UDM"

  agent_cluster = "Opswise - Default Linux/Unix Cluster"
  exit_codes    = "0"

  primary_broker_choice   = "Agent"
  primary_broker          = "udm-agent-01"
  secondary_broker_choice = "Agent Cluster"
  secondary_cluster_ref   = "Opswise - Default Linux/Unix Cluster"

  local_filename  = "/data/incoming"
  remote_filename = "/remote/outgoing/report-updated.csv"

  command       = "GET"
  udm_operation = "Copy"
  format        = "Binary"

  form_or_script = "Form"
}
`, name)
}
