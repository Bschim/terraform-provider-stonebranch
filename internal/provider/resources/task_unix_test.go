package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccTaskUnixResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	resourceName := "stonebranch_task_unix.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccTaskUnixConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "command", "echo hello"),
					resource.TestCheckResourceAttrSet(resourceName, "sys_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			// ImportState - use the task name as import ID
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
			},
			// Update
			{
				Config: testAccTaskUnixConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "command", "echo updated"),
					resource.TestCheckResourceAttr(resourceName, "summary", "Updated task summary"),
				),
			},
		},
	})
}

func TestAccTaskUnixResource_withScript(t *testing.T) {
	// This test creates a script resource and a Unix task that references it
	scriptName := acctest.RandomWithPrefix("tf-test-script")
	taskName := acctest.RandomWithPrefix("tf-test-task")
	taskResourceName := "stonebranch_task_unix.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskUnixConfig_withScript(taskName, scriptName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(taskResourceName, "name", taskName),
					resource.TestCheckResourceAttr(taskResourceName, "command_or_script", "Script"),
					resource.TestCheckResourceAttr(taskResourceName, "script", scriptName),
				),
			},
		},
	})
}

func TestAccTaskUnixResource_withSummary(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	resourceName := "stonebranch_task_unix.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskUnixConfig_withSummary(rName, "Initial summary"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "summary", "Initial summary"),
				),
			},
			{
				Config: testAccTaskUnixConfig_withSummary(rName, "Changed summary"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "summary", "Changed summary"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccTaskUnixConfig_basic(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_unix" "test" {
  name       = %[1]q
  command    = "echo hello"
  agent_var  = "agent_name"
  exit_codes = "0"
}
`, name)
}

func testAccTaskUnixConfig_updated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_unix" "test" {
  name       = %[1]q
  command    = "echo updated"
  summary    = "Updated task summary"
  agent_var  = "agent_name"
  exit_codes = "0"
}
`, name)
}

func testAccTaskUnixConfig_withScript(taskName, scriptName string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_script" "test" {
  name    = %[2]q
  content = "echo 'Hello from script'"
}

resource "stonebranch_task_unix" "test" {
  name              = %[1]q
  command_or_script = "Script"
  script            = stonebranch_script.test.name
  agent_var         = "agent_name"
  exit_codes        = "0"
}
`, taskName, scriptName)
}

func testAccTaskUnixConfig_withSummary(name, summary string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_unix" "test" {
  name       = %[1]q
  command    = "echo hello"
  summary    = %[2]q
  agent_var  = "agent_name"
  exit_codes = "0"
}
`, name, summary)
}

func TestAccTaskUnixResource_withVariables(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	resourceName := "stonebranch_task_unix.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with variables
			{
				Config: testAccTaskUnixConfig_withVariables(rName, "initial_value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "variables.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "variables.0.name", "test_var1"),
					resource.TestCheckResourceAttr(resourceName, "variables.0.value", "initial_value"),
					resource.TestCheckResourceAttr(resourceName, "variables.0.description", "Test variable 1"),
					resource.TestCheckResourceAttr(resourceName, "variables.1.name", "test_var2"),
					resource.TestCheckResourceAttr(resourceName, "variables.1.value", "value2"),
				),
			},
			// Update variables
			{
				Config: testAccTaskUnixConfig_withVariables(rName, "updated_value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "variables.0.value", "updated_value"),
				),
			},
		},
	})
}

func testAccTaskUnixConfig_withVariables(name, varValue string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_unix" "test" {
  name       = %[1]q
  command    = "echo $${test_var1}"
  agent_var  = "agent_name"
  exit_codes = "0"

  variables = [
    {
      name        = "test_var1"
      value       = %[2]q
      description = "Test variable 1"
    },
    {
      name  = "test_var2"
      value = "value2"
    }
  ]
}
`, name, varValue)
}

func TestAccTaskUnixResource_withResourceManagement(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	resourceName := "stonebranch_task_unix.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskUnixConfig_withResourceManagement(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "hold_resources", "true"),
					resource.TestCheckResourceAttr(resourceName, "exclusive_tasks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "exclusive_tasks.0.task", rName+"-other"),
					resource.TestCheckResourceAttr(resourceName, "virtual_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "virtual_resources.0.resource", rName+"-vres"),
					resource.TestCheckResourceAttr(resourceName, "virtual_resources.0.amount", "1"),
				),
			},
			// Update amount
			{
				Config: testAccTaskUnixConfig_withResourceManagement(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "virtual_resources.0.amount", "2"),
				),
			},
		},
	})
}

func testAccTaskUnixConfig_withResourceManagement(name string, amount int) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_unix" "other" {
  name       = "%[1]s-other"
  agent_var  = "agent_name"
  command    = "echo other"
  exit_codes = "0"
}

resource "stonebranch_virtual_resource" "test" {
  name  = "%[1]s-vres"
  limit = 5
}

resource "stonebranch_task_unix" "test" {
  name       = %[1]q
  command    = "echo hello"
  agent_var  = "agent_name"
  exit_codes = "0"

  hold_resources = true

  exclusive_tasks = [
    {
      task = stonebranch_task_unix.other.name
    }
  ]

  virtual_resources = [
    {
      resource = stonebranch_virtual_resource.test.name
      amount   = %[2]d
    }
  ]
}
`, name, amount)
}
