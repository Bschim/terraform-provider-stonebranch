package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccTaskRecurringResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-recurring")
	targetName := acctest.RandomWithPrefix("tf-test-recurring-target")
	resourceName := "stonebranch_task_recurring.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccTaskRecurringConfig_basic(rName, targetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "target_task", targetName),
					resource.TestCheckResourceAttr(resourceName, "recurrence_type", "Interval"),
					resource.TestCheckResourceAttr(resourceName, "recurrence_interval", "15"),
					resource.TestCheckResourceAttr(resourceName, "recurrence_interval_unit", "Minutes"),
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
				Config: testAccTaskRecurringConfig_updated(rName, targetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "recurrence_interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "summary", "Updated recurring task"),
				),
			},
		},
	})
}

func TestAccTaskRecurringResource_withTimeWindow(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-recurring")
	targetName := acctest.RandomWithPrefix("tf-test-recurring-target")
	resourceName := "stonebranch_task_recurring.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskRecurringConfig_withTimeWindow(rName, targetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "time_window", "true"),
					resource.TestCheckResourceAttr(resourceName, "interval_start_time", "01:00"),
					resource.TestCheckResourceAttr(resourceName, "interval_end_time", "23:40"),
				),
			},
		},
	})
}

func TestAccTaskRecurringResource_withNumberOfRecurrences(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-recurring")
	targetName := acctest.RandomWithPrefix("tf-test-recurring-target")
	resourceName := "stonebranch_task_recurring.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskRecurringConfig_withNumberOfRecurrences(rName, targetName, "5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "indefinite_recurrences", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_of_recurrences", "5"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccTaskRecurringConfig_target(targetName string) string {
	return fmt.Sprintf(`
resource "stonebranch_task_unix" "target" {
  name       = %[1]q
  command    = "echo target"
  agent_var  = "agent_name"
  exit_codes = "0"
}
`, targetName)
}

func testAccTaskRecurringConfig_basic(name, targetName string) string {
	return sbacctest.ProviderConfig() + testAccTaskRecurringConfig_target(targetName) + fmt.Sprintf(`
resource "stonebranch_task_recurring" "test" {
  name                     = %[1]q
  target_task              = stonebranch_task_unix.target.name
  recurrence_type          = "Interval"
  recurrence_interval      = "15"
  recurrence_interval_unit = "Minutes"
  indefinite_recurrences   = true
}
`, name)
}

func testAccTaskRecurringConfig_updated(name, targetName string) string {
	return sbacctest.ProviderConfig() + testAccTaskRecurringConfig_target(targetName) + fmt.Sprintf(`
resource "stonebranch_task_recurring" "test" {
  name                     = %[1]q
  target_task              = stonebranch_task_unix.target.name
  summary                  = "Updated recurring task"
  recurrence_type          = "Interval"
  recurrence_interval      = "30"
  recurrence_interval_unit = "Minutes"
  indefinite_recurrences   = true
}
`, name)
}

func testAccTaskRecurringConfig_withTimeWindow(name, targetName string) string {
	return sbacctest.ProviderConfig() + testAccTaskRecurringConfig_target(targetName) + fmt.Sprintf(`
resource "stonebranch_task_recurring" "test" {
  name                     = %[1]q
  target_task              = stonebranch_task_unix.target.name
  recurrence_type          = "Interval"
  recurrence_interval      = "20"
  recurrence_interval_unit = "Minutes"
  indefinite_recurrences   = false

  time_window         = true
  interval_start_time = "01:00"
  interval_end_time   = "23:40"
}
`, name)
}

func testAccTaskRecurringConfig_withNumberOfRecurrences(name, targetName, numberOfRecurrences string) string {
	return sbacctest.ProviderConfig() + testAccTaskRecurringConfig_target(targetName) + fmt.Sprintf(`
resource "stonebranch_task_recurring" "test" {
  name                     = %[1]q
  target_task              = stonebranch_task_unix.target.name
  recurrence_type          = "Interval"
  recurrence_interval      = "15"
  recurrence_interval_unit = "Minutes"
  indefinite_recurrences   = false
  number_of_recurrences    = %[2]q
}
`, name, numberOfRecurrences)
}
