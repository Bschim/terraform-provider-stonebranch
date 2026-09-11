package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccWorkflowEdgeResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wfe")
	resourceName := "stonebranch_workflow_edge.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflowEdgeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "workflow_name", rName+"-wf"),
					resource.TestCheckResourceAttrSet(resourceName, "source_id"),
					resource.TestCheckResourceAttrSet(resourceName, "target_id"),
					resource.TestCheckResourceAttr(resourceName, "straight_edge", "true"),
				),
			},
		},
	})
}

func TestAccWorkflowEdgeResource_multipleEdges(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wfe")
	resourceName1 := "stonebranch_workflow_edge.edge1"
	resourceName2 := "stonebranch_workflow_edge.edge2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowEdgeConfig_multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "workflow_name", rName+"-wf"),
					resource.TestCheckResourceAttrSet(resourceName1, "source_id"),
					resource.TestCheckResourceAttrSet(resourceName1, "target_id"),
					resource.TestCheckResourceAttr(resourceName2, "workflow_name", rName+"-wf"),
					resource.TestCheckResourceAttrSet(resourceName2, "source_id"),
					resource.TestCheckResourceAttrSet(resourceName2, "target_id"),
				),
			},
		},
	})
}

func TestAccWorkflowEdgeResource_conditionStatus(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wfe-cs")
	resourceName := "stonebranch_workflow_edge.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowEdgeConfig_condition(rName, `
  condition = {
    type   = "Status"
    status = "Failure"
  }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "condition.type", "Status"),
					resource.TestCheckResourceAttr(resourceName, "condition.status", "Failure"),
				),
			},
			// Update the status value in place. Regression test for the Computed
			// condition leaves (exit_code/first_value/operator/second_value) being
			// left Unknown after apply, since their prior state was null.
			{
				Config: testAccWorkflowEdgeConfig_condition(rName, `
  condition = {
    type   = "Status"
    status = "Success/Failure"
  }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "condition.type", "Status"),
					resource.TestCheckResourceAttr(resourceName, "condition.status", "Success/Failure"),
				),
			},
		},
	})
}

func TestAccWorkflowEdgeResource_conditionExitCode(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wfe-ec")
	resourceName := "stonebranch_workflow_edge.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowEdgeConfig_condition(rName, `
  condition = {
    type      = "Exit Code"
    exit_code = "0"
  }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "condition.type", "Exit Code"),
					resource.TestCheckResourceAttr(resourceName, "condition.exit_code", "0"),
				),
			},
		},
	})
}

func TestAccWorkflowEdgeResource_conditionVariable(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-wfe-var")
	resourceName := "stonebranch_workflow_edge.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowEdgeConfig_condition(rName, `
  condition = {
    type         = "Variable"
    first_value  = "$${OUTPUT_LINE_COUNT}"
    operator     = "!="
    second_value = "$${FILE_COUNT}"
  }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "condition.type", "Variable"),
					resource.TestCheckResourceAttr(resourceName, "condition.first_value", "${OUTPUT_LINE_COUNT}"),
					resource.TestCheckResourceAttr(resourceName, "condition.operator", "!="),
					resource.TestCheckResourceAttr(resourceName, "condition.second_value", "${FILE_COUNT}"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccWorkflowEdgeConfig_condition(name string, conditionBlock string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_workflow" "test" {
  name = "%[1]s-wf"
}

resource "stonebranch_task_unix" "task1" {
  name       = "%[1]s-task1"
  agent_var  = "agent_name"
  command    = "echo task1"
  exit_codes = "0"
}

resource "stonebranch_task_unix" "task2" {
  name       = "%[1]s-task2"
  agent_var  = "agent_name"
  command    = "echo task2"
  exit_codes = "0"
}

resource "stonebranch_workflow_vertex" "vertex1" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.task1.name
}

resource "stonebranch_workflow_vertex" "vertex2" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.task2.name
}

resource "stonebranch_workflow_edge" "test" {
  workflow_name = stonebranch_task_workflow.test.name
  source_id     = stonebranch_workflow_vertex.vertex1.vertex_id
  target_id     = stonebranch_workflow_vertex.vertex2.vertex_id
%[2]s
}
`, name, conditionBlock)
}

func testAccWorkflowEdgeConfig_basic(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_workflow" "test" {
  name = "%[1]s-wf"
}

resource "stonebranch_task_unix" "task1" {
  name       = "%[1]s-task1"
  agent_var  = "agent_name"
  command    = "echo task1"
  exit_codes = "0"
}

resource "stonebranch_task_unix" "task2" {
  name       = "%[1]s-task2"
  agent_var  = "agent_name"
  command    = "echo task2"
  exit_codes = "0"
}

resource "stonebranch_workflow_vertex" "vertex1" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.task1.name
}

resource "stonebranch_workflow_vertex" "vertex2" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.task2.name
}

resource "stonebranch_workflow_edge" "test" {
  workflow_name = stonebranch_task_workflow.test.name
  source_id     = stonebranch_workflow_vertex.vertex1.vertex_id
  target_id     = stonebranch_workflow_vertex.vertex2.vertex_id
}
`, name)
}

func testAccWorkflowEdgeConfig_multiple(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_workflow" "test" {
  name = "%[1]s-wf"
}

resource "stonebranch_task_unix" "task1" {
  name       = "%[1]s-task1"
  agent_var  = "agent_name"
  command    = "echo task1"
  exit_codes = "0"
}

resource "stonebranch_task_unix" "task2" {
  name       = "%[1]s-task2"
  agent_var  = "agent_name"
  command    = "echo task2"
  exit_codes = "0"
}

resource "stonebranch_task_unix" "task3" {
  name       = "%[1]s-task3"
  agent_var  = "agent_name"
  command    = "echo task3"
  exit_codes = "0"
}

resource "stonebranch_workflow_vertex" "vertex1" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.task1.name
}

resource "stonebranch_workflow_vertex" "vertex2" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.task2.name
}

resource "stonebranch_workflow_vertex" "vertex3" {
  workflow_name = stonebranch_task_workflow.test.name
  task_name     = stonebranch_task_unix.task3.name
}

# task1 -> task2
resource "stonebranch_workflow_edge" "edge1" {
  workflow_name = stonebranch_task_workflow.test.name
  source_id     = stonebranch_workflow_vertex.vertex1.vertex_id
  target_id     = stonebranch_workflow_vertex.vertex2.vertex_id
}

# task1 -> task3 (fork from task1)
resource "stonebranch_workflow_edge" "edge2" {
  workflow_name = stonebranch_task_workflow.test.name
  source_id     = stonebranch_workflow_vertex.vertex1.vertex_id
  target_id     = stonebranch_workflow_vertex.vertex3.vertex_id
}
`, name)
}
