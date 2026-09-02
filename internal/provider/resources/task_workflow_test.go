package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccTaskWorkflowResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wf")
	resourceName := "stonebranch_task_workflow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccTaskWorkflowConfig_basic(rName),
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
				Config: testAccTaskWorkflowConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "summary", "Updated workflow"),
				),
			},
		},
	})
}

func TestAccTaskWorkflowResource_withSummary(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wf")
	resourceName := "stonebranch_task_workflow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskWorkflowConfig_withSummary(rName, "Initial summary"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "summary", "Initial summary"),
				),
			},
			{
				Config: testAccTaskWorkflowConfig_withSummary(rName, "Changed summary"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "summary", "Changed summary"),
				),
			},
		},
	})
}

func TestAccTaskWorkflowResource_withOptions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wf")
	resourceName := "stonebranch_task_workflow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskWorkflowConfig_withOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "calculate_critical_path", "true"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccTaskWorkflowConfig_basic(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_workflow" "test" {
  name = %[1]q
}
`, name)
}

func testAccTaskWorkflowConfig_updated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_workflow" "test" {
  name    = %[1]q
  summary = "Updated workflow"
}
`, name)
}

func testAccTaskWorkflowConfig_withSummary(name, summary string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_workflow" "test" {
  name    = %[1]q
  summary = %[2]q
}
`, name, summary)
}

func testAccTaskWorkflowConfig_withOptions(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_workflow" "test" {
  name                    = %[1]q
  summary                 = "Workflow with options"
  calculate_critical_path = true
}
`, name)
}

func TestAccTaskWorkflowResource_withRunCriteria(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wf")
	resourceName := "stonebranch_task_workflow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create the workflow and add the task as a vertex.
			// The Stonebranch API requires a task to already be present as a
			// vertex in the workflow before run_criteria can reference it.
			{
				Config: testAccTaskWorkflowConfig_withRunCriteria_step1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			// Step 2: now add run_criteria referencing that task.
			{
				Config: testAccTaskWorkflowConfig_withRunCriteria_step2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "run_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "run_criteria.0.type", "Skip Criteria"),
					resource.TestCheckResourceAttr(resourceName, "run_criteria.0.task", rName+"-task"),
					resource.TestCheckResourceAttr(resourceName, "run_criteria.0.business_day", "true"),
				),
			},
		},
	})
}

func testAccTaskWorkflowConfig_withRunCriteria_step1(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_unix" "test" {
  name       = "%[1]s-task"
  agent_var  = "agent_name"
  command    = "echo hello"
  exit_codes = "0"
}

resource "stonebranch_task_workflow" "test" {
  name = %[1]q
}

resource "stonebranch_workflow_vertex" "test" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.test.name
}
`, name)
}

func testAccTaskWorkflowConfig_withRunCriteria_step2(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_unix" "test" {
  name       = "%[1]s-task"
  agent_var  = "agent_name"
  command    = "echo hello"
  exit_codes = "0"
}

resource "stonebranch_task_workflow" "test" {
  name = %[1]q

  run_criteria = [
    {
      type         = "Skip Criteria"
      task         = stonebranch_task_unix.test.name
      business_day = true
    }
  ]
}

resource "stonebranch_workflow_vertex" "test" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.test.name
}
`, name)
}
