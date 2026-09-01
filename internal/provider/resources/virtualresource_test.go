package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccVirtualResourceResource_basic(t *testing.T) {
	rName := "tf-test-virtual-resource-" + acctest.RandString(8)
	resourceName := "stonebranch_virtual_resource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccVirtualResourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "limit", "5"),
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
				Config: testAccVirtualResourceConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "limit", "10"),
					resource.TestCheckResourceAttr(resourceName, "summary", "Updated virtual resource"),
				),
			},
		},
	})
}

func TestAccVirtualResourceResource_depletable(t *testing.T) {
	rName := "tf-test-virtual-resource-" + acctest.RandString(8)
	resourceName := "stonebranch_virtual_resource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualResourceConfig_depletable(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "Depletable"),
					resource.TestCheckResourceAttr(resourceName, "limit", "100"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccVirtualResourceConfig_basic(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_virtual_resource" "test" {
  name  = %[1]q
  limit = 5
}
`, name)
}

func testAccVirtualResourceConfig_updated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_virtual_resource" "test" {
  name    = %[1]q
  limit   = 10
  summary = "Updated virtual resource"
}
`, name)
}

func testAccVirtualResourceConfig_depletable(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_virtual_resource" "test" {
  name  = %[1]q
  type  = "Depletable"
  limit = 100
}
`, name)
}
