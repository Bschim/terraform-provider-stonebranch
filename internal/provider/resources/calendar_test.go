package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sbacctest "github.com/OptionMetrics/terraform-provider-stonebranch/internal/acctest"
)

func TestAccCalendarResource_basic(t *testing.T) {
	rName := "tf-test-calendar-" + acctest.RandString(8)
	resourceName := "stonebranch_calendar.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccCalendarConfig_basic(rName),
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
				Config: testAccCalendarConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comments", "Updated calendar"),
				),
			},
		},
	})
}

func TestAccCalendarResource_withBusinessDays(t *testing.T) {
	rName := "tf-test-calendar-" + acctest.RandString(8)
	resourceName := "stonebranch_calendar.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCalendarConfig_withBusinessDays(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "business_days", "Monday,Tuesday,Wednesday,Thursday,Friday"),
					resource.TestCheckResourceAttr(resourceName, "first_day_of_week", "Monday"),
				),
			},
		},
	})
}

func TestAccCalendarResource_withQuarters(t *testing.T) {
	rName := "tf-test-calendar-" + acctest.RandString(8)
	resourceName := "stonebranch_calendar.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCalendarConfig_withQuarters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "first_quarter_month", "Jan"),
					resource.TestCheckResourceAttr(resourceName, "first_quarter_day", "1"),
					resource.TestCheckResourceAttr(resourceName, "second_quarter_month", "Apr"),
					resource.TestCheckResourceAttr(resourceName, "second_quarter_day", "1"),
				),
			},
		},
	})
}

func TestAccCalendarResource_withCustomDays(t *testing.T) {
	rName := "tf-test-calendar-" + acctest.RandString(8)
	resourceName := "stonebranch_calendar.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sbacctest.PreCheck(t) },
		ProtoV6ProviderFactories: sbacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCalendarConfig_withCustomDays(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "custom_days.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_days.0", rName+"-cd"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccCalendarConfig_basic(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_calendar" "test" {
  name                 = %[1]q
  first_quarter_month  = "Jan"
  first_quarter_day    = "1"
  second_quarter_month = "Apr"
  second_quarter_day   = "1"
  third_quarter_month  = "Jul"
  third_quarter_day    = "1"
  fourth_quarter_month = "Oct"
  fourth_quarter_day   = "1"
}
`, name)
}

func testAccCalendarConfig_updated(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_calendar" "test" {
  name                 = %[1]q
  comments             = "Updated calendar"
  first_quarter_month  = "Jan"
  first_quarter_day    = "1"
  second_quarter_month = "Apr"
  second_quarter_day   = "1"
  third_quarter_month  = "Jul"
  third_quarter_day    = "1"
  fourth_quarter_month = "Oct"
  fourth_quarter_day   = "1"
}
`, name)
}

func testAccCalendarConfig_withBusinessDays(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_calendar" "test" {
  name                 = %[1]q
  business_days        = "Monday,Tuesday,Wednesday,Thursday,Friday"
  first_day_of_week    = "Monday"
  first_quarter_month  = "Jan"
  first_quarter_day    = "1"
  second_quarter_month = "Apr"
  second_quarter_day   = "1"
  third_quarter_month  = "Jul"
  third_quarter_day    = "1"
  fourth_quarter_month = "Oct"
  fourth_quarter_day   = "1"
}
`, name)
}

func testAccCalendarConfig_withQuarters(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_calendar" "test" {
  name                 = %[1]q
  first_quarter_month  = "Jan"
  first_quarter_day    = "1"
  second_quarter_month = "Apr"
  second_quarter_day   = "1"
  third_quarter_month  = "Jul"
  third_quarter_day    = "1"
  fourth_quarter_month = "Oct"
  fourth_quarter_day   = "1"
}
`, name)
}

func testAccCalendarConfig_withCustomDays(name string) string {
	return sbacctest.ProviderConfig() + fmt.Sprintf(`
resource "stonebranch_custom_day" "test" {
  name  = "%[1]s-cd"
  ctype = "Single Date"
  date  = "2027-01-01"
}

resource "stonebranch_calendar" "test" {
  name                 = %[1]q
  first_quarter_month  = "Jan"
  first_quarter_day    = "1"
  second_quarter_month = "Apr"
  second_quarter_day   = "1"
  third_quarter_month  = "Jul"
  third_quarter_day    = "1"
  fourth_quarter_month = "Oct"
  fourth_quarter_day   = "1"
  custom_days          = [stonebranch_custom_day.test.name]
}
`, name)
}
