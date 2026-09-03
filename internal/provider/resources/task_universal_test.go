package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccTaskUniversalResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	resourceName := "stonebranch_task_universal.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccTaskUniversalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "template", "tf-example-aws-deploy"),
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
				Config: testAccTaskUniversalConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "summary", "Updated task summary"),
				),
			},
		},
	})
}

func TestAccTaskUniversalResource_genericFields(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	resourceName := "stonebranch_task_universal.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskUniversalConfig_genericFields(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "text_field_1", "hello"),
					resource.TestCheckResourceAttr(resourceName, "choice_field_1", "option-a"),
					resource.TestCheckResourceAttr(resourceName, "boolean_field_1", "true"),
					resource.TestCheckResourceAttr(resourceName, "int_field_1", "42"),
					resource.TestCheckResourceAttr(resourceName, "large_text_field_1", "some multi-line\ntext value"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccTaskUniversalConfig_basic(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_universal" "test" {
  name      = %[1]q
  agent_var = "agent_name"
  template  = "tf-example-aws-deploy"
}
`, name)
}

func testAccTaskUniversalConfig_updated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_universal" "test" {
  name      = %[1]q
  summary   = "Updated task summary"
  agent_var = "agent_name"
  template  = "tf-example-aws-deploy"
}
`, name)
}

func testAccTaskUniversalConfig_genericFields(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_task_universal" "test" {
  name      = %[1]q
  agent_var = "agent_name"
  template  = "tf-example-aws-deploy"

  text_field_1       = "hello"
  choice_field_1     = "option-a"
  boolean_field_1    = true
  int_field_1        = 42
  large_text_field_1 = "some multi-line\ntext value"
}
`, name)
}
