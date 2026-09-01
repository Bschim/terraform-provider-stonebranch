package resources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccEmailTemplateResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-et")
	cName := acctest.RandomWithPrefix("tf-test-etconn")
	resourceName := "stonebranch_email_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccEmailTemplateConfig_basic(rName, cName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "email_connection", cName),
					resource.TestCheckResourceAttr(resourceName, "to", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "subject", "Test Subject"),
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
				Config: testAccEmailTemplateConfig_updated(rName, cName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "subject", "Updated Subject"),
					resource.TestCheckResourceAttr(resourceName, "body", "Updated body content"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated email template"),
				),
			},
		},
	})
}

func TestAccEmailTemplateResource_ccOnly(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-et")
	cName := acctest.RandomWithPrefix("tf-test-etconn")
	resourceName := "stonebranch_email_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig_ccOnly(rName, cName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "cc", "cc@example.com"),
					resource.TestCheckNoResourceAttr(resourceName, "to"),
				),
			},
		},
	})
}

func TestAccEmailTemplateResource_noRecipients(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-et")
	cName := acctest.RandomWithPrefix("tf-test-etconn")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEmailTemplateConfig_noRecipients(rName, cName),
				ExpectError: regexp.MustCompile(`At least one of "to", "cc", or "bcc" must be set`),
			},
		},
	})
}

// Test configuration helpers

func testAccEmailTemplateConfig_basic(name, connName string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_email_connection" "test" {
  name          = %[2]q
  smtp          = "smtp.example.com"
  smtp_port     = 25
  email_address = "test@example.com"
}

resource "stonebranch_email_template" "test" {
  name       = %[1]q
  email_connection = stonebranch_email_connection.test.name
  to         = "test@example.com"
  subject    = "Test Subject"
  body       = "Test body content"
}
`, name, connName)
}

func testAccEmailTemplateConfig_updated(name, connName string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_email_connection" "test" {
  name          = %[2]q
  smtp          = "smtp.example.com"
  smtp_port     = 25
  email_address = "test@example.com"
}

resource "stonebranch_email_template" "test" {
  name        = %[1]q
  email_connection = stonebranch_email_connection.test.name
  to          = "test@example.com"
  subject     = "Updated Subject"
  body        = "Updated body content"
  description = "Updated email template"
}
`, name, connName)
}

func testAccEmailTemplateConfig_ccOnly(name, connName string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_email_connection" "test" {
  name          = %[2]q
  smtp          = "smtp.example.com"
  smtp_port     = 25
  email_address = "test@example.com"
}

resource "stonebranch_email_template" "test" {
  name       = %[1]q
  email_connection = stonebranch_email_connection.test.name
  cc         = "cc@example.com"
}
`, name, connName)
}

func testAccEmailTemplateConfig_noRecipients(name, connName string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_email_connection" "test" {
  name          = %[2]q
  smtp          = "smtp.example.com"
  smtp_port     = 25
  email_address = "test@example.com"
}

resource "stonebranch_email_template" "test" {
  name       = %[1]q
  email_connection = stonebranch_email_connection.test.name
}
`, name, connName)
}
