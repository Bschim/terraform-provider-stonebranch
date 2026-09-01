package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccUniversalTemplateResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-ut")
	resourceName := "stonebranch_universal_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccUniversalTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "agent_type", "Any"),
					resource.TestCheckResourceAttr(resourceName, "variable_prefix", "TF"),
					resource.TestCheckResourceAttr(resourceName, "exit_codes", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "sys_id"),
					resource.TestCheckResourceAttr(resourceName, "template_type", "Script"),
					resource.TestCheckResourceAttr(resourceName, "exit_code_processing", "Success Exitcode Range"),
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
				Config: testAccUniversalTemplateConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated universal template"),
					resource.TestCheckResourceAttr(resourceName, "script", "echo updated"),
					resource.TestCheckResourceAttr(resourceName, "use_common_script", "true"),
				),
			},
		},
	})
}

func TestAccUniversalTemplateResource_windows(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-ut")
	resourceName := "stonebranch_universal_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUniversalTemplateConfig_windows(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "agent_type", "Windows"),
					resource.TestCheckResourceAttr(resourceName, "script_windows", "echo hi"),
					resource.TestCheckResourceAttr(resourceName, "elevate_user", "true"),
				),
			},
		},
	})
}

func TestAccUniversalTemplateResource_environment(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-ut")
	resourceName := "stonebranch_universal_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUniversalTemplateConfig_environment(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.name", "FOO"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.value", "bar"),
				),
			},
			// Update: omit environment entirely to confirm full-replace clears it.
			{
				Config: testAccUniversalTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environment.#", "0"),
				),
			},
		},
	})
}

func TestAccUniversalTemplateResource_fields(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-ut")
	resourceName := "stonebranch_universal_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUniversalTemplateConfig_fields(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.name", "region"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.field_mapping", "Choice Field 1"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.field_type", "Choice"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.choices.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.choices.0.field_value", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "fields.1.name", "tags"),
					resource.TestCheckResourceAttr(resourceName, "fields.1.field_mapping", "Array Field 1"),
					resource.TestCheckResourceAttr(resourceName, "fields.1.field_type", "Array"),
					resource.TestCheckResourceAttr(resourceName, "fields.1.array_field_value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fields.1.array_field_value.0.name", "env"),
					resource.TestCheckResourceAttr(resourceName, "fields.2.name", "notes"),
					resource.TestCheckResourceAttr(resourceName, "fields.2.field_mapping", "Text Field 1"),
					resource.TestCheckResourceAttr(resourceName, "fields.2.field_type", "Text"),
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
			// Update: replace with a different non-empty field set. Note: the
			// live API rejects updates that would reduce `fields` to empty, so
			// this step intentionally does not test clearing fields entirely.
			{
				Config: testAccUniversalTemplateConfig_fieldsUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.name", "priority"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.field_mapping", "Integer Field 1"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.field_type", "Integer"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccUniversalTemplateConfig_basic(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_universal_template" "test" {
  name            = %[1]q
  variable_prefix = "TF"
  agent_type      = "Any"
  exit_codes      = "0"
}
`, name)
}

func testAccUniversalTemplateConfig_updated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_universal_template" "test" {
  name              = %[1]q
  description       = "Updated universal template"
  variable_prefix   = "TF"
  agent_type        = "Any"
  exit_codes        = "0"
  use_common_script = true
  script            = "echo updated"
}
`, name)
}

func testAccUniversalTemplateConfig_windows(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_universal_template" "test" {
  name            = %[1]q
  variable_prefix = "TF"
  agent_type      = "Windows"
  exit_codes      = "0"
  script_windows  = "echo hi"
  elevate_user    = true
}
`, name)
}

func testAccUniversalTemplateConfig_environment(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_universal_template" "test" {
  name            = %[1]q
  variable_prefix = "TF"
  agent_type      = "Any"
  exit_codes      = "0"

  environment = [
    { name = "FOO", value = "bar" },
    { name = "BAZ", value = "qux" },
  ]
}
`, name)
}

func testAccUniversalTemplateConfig_fields(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_universal_template" "test" {
  name            = %[1]q
  variable_prefix = "TF"
  agent_type      = "Any"
  exit_codes      = "0"

  fields = [
    {
      name          = "region"
      label         = "Region"
      field_mapping = "Choice Field 1"
      field_type    = "Choice"
      choices = [
        { field_value = "us-east-1", field_value_label = "US East 1" },
        { field_value = "us-west-2", field_value_label = "US West 2" },
      ]
    },
    {
      name              = "tags"
      label             = "Tags"
      field_mapping     = "Array Field 1"
      field_type        = "Array"
      array_name_title  = "Key"
      array_value_title = "Value"
      array_field_value = [
        { name = "env", value = "prod" },
      ]
    },
    {
      name          = "notes"
      label         = "Notes"
      field_mapping = "Text Field 1"
      field_type    = "Text"
    },
  ]
}
`, name)
}

func testAccUniversalTemplateConfig_fieldsUpdated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_universal_template" "test" {
  name            = %[1]q
  variable_prefix = "TF"
  agent_type      = "Any"
  exit_codes      = "0"

  fields = [
    {
      name          = "priority"
      label         = "Priority"
      field_mapping = "Integer Field 1"
      field_type    = "Integer"
      int_field_min = 1
      int_field_max = 5
    },
  ]
}
`, name)
}
