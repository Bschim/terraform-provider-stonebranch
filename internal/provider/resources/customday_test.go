package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccCustomDayResource_singleDate(t *testing.T) {
	rName := "tf-test-custom-day-" + acctest.RandString(8)
	resourceName := "stonebranch_custom_day.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccCustomDayConfig_singleDate(rName, "2026-01-01"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ctype", "Single Date"),
					resource.TestCheckResourceAttr(resourceName, "date", "2026-01-01"),
					resource.TestCheckResourceAttr(resourceName, "category", "Day"),
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
				Config: testAccCustomDayConfig_singleDate(rName, "2026-06-15"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "date", "2026-06-15"),
				),
			},
		},
	})
}

func TestAccCustomDayResource_listOfDates(t *testing.T) {
	rName := "tf-test-custom-day-" + acctest.RandString(8)
	resourceName := "stonebranch_custom_day.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDayConfig_listOfDates(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ctype", "List of Dates"),
					resource.TestCheckResourceAttr(resourceName, "date_list.#", "2"),
				),
			},
		},
	})
}

func TestAccCustomDayResource_absoluteRepeating(t *testing.T) {
	rName := "tf-test-custom-day-" + acctest.RandString(8)
	resourceName := "stonebranch_custom_day.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDayConfig_absoluteRepeating(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ctype", "Absolute Repeating Date"),
					resource.TestCheckResourceAttr(resourceName, "month", "Dec"),
					resource.TestCheckResourceAttr(resourceName, "day", "25"),
				),
			},
		},
	})
}

func TestAccCustomDayResource_relativeRepeatingHoliday(t *testing.T) {
	rName := "tf-test-custom-day-" + acctest.RandString(8)
	resourceName := "stonebranch_custom_day.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDayConfig_relativeRepeatingHoliday(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ctype", "Relative Repeating Date"),
					resource.TestCheckResourceAttr(resourceName, "month", "Nov"),
					resource.TestCheckResourceAttr(resourceName, "dayofweek", "Thu"),
					resource.TestCheckResourceAttr(resourceName, "relfreq", "4th"),
					resource.TestCheckResourceAttr(resourceName, "holiday", "true"),
					resource.TestCheckResourceAttr(resourceName, "category", "Holiday"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccCustomDayConfig_singleDate(name, date string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_custom_day" "test" {
  name  = %[1]q
  ctype = "Single Date"
  date  = %[2]q
}
`, name, date)
}

func testAccCustomDayConfig_listOfDates(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_custom_day" "test" {
  name      = %[1]q
  ctype     = "List of Dates"
  date_list = ["2026-02-01", "2026-02-14"]
}
`, name)
}

func testAccCustomDayConfig_absoluteRepeating(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_custom_day" "test" {
  name  = %[1]q
  ctype = "Absolute Repeating Date"
  month = "Dec"
  day   = 25
}
`, name)
}

func testAccCustomDayConfig_relativeRepeatingHoliday(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_custom_day" "test" {
  name      = %[1]q
  ctype     = "Relative Repeating Date"
  month     = "Nov"
  dayofweek = "Thu"
  relfreq   = "4th"
  holiday   = true
}
`, name)
}
