# Stonebranch Calendar Example
#
# This example demonstrates how to create calendars in Stonebranch.
#
# Usage:
#   export STONEBRANCH_API_TOKEN="your-token"
#   terraform init
#   terraform plan
#   terraform apply

terraform {
  required_providers {
    stonebranch = {
      source = "stonebranch/stonebranch"
    }
  }
}

provider "stonebranch" {
  # Uses STONEBRANCH_API_TOKEN environment variable
}

# Basic calendar with required quarter settings
resource "stonebranch_calendar" "basic" {
  name = "tf-example-calendar-basic"

  # Quarter settings are required by the API
  first_quarter_month  = "Jan"
  first_quarter_day    = "1"
  second_quarter_month = "Apr"
  second_quarter_day   = "1"
  third_quarter_month  = "Jul"
  third_quarter_day    = "1"
  fourth_quarter_month = "Oct"
  fourth_quarter_day   = "1"
}

# Calendar with standard business days (Monday-Friday)
resource "stonebranch_calendar" "business" {
  name              = "tf-example-calendar-business"
  comments          = "Standard business calendar for weekday operations"
  business_days     = "Monday,Tuesday,Wednesday,Thursday,Friday"
  first_day_of_week = "Monday"
}

# Calendar with fiscal quarters (calendar year)
resource "stonebranch_calendar" "fiscal" {
  name     = "tf-example-calendar-fiscal"
  comments = "Fiscal calendar with standard quarters"

  business_days     = "Monday,Tuesday,Wednesday,Thursday,Friday"
  first_day_of_week = "Monday"

  # Q1: January 1
  first_quarter_month = "Jan"
  first_quarter_day   = "1"

  # Q2: April 1
  second_quarter_month = "Apr"
  second_quarter_day   = "1"

  # Q3: July 1
  third_quarter_month = "Jul"
  third_quarter_day   = "1"

  # Q4: October 1
  fourth_quarter_month = "Oct"
  fourth_quarter_day   = "1"
}

# Calendar for weekend operations
resource "stonebranch_calendar" "weekend" {
  name              = "tf-example-calendar-weekend"
  comments          = "Calendar for weekend-only operations"
  business_days     = "Saturday,Sunday"
  first_day_of_week = "Sunday"
}

# Calendar with custom-day exceptions (holidays)
resource "stonebranch_custom_day" "new_years" {
  name  = "tf-example-custom-day-new-years"
  ctype = "Single Date"
  date  = "2027-01-01"
}

resource "stonebranch_calendar" "with_holidays" {
  name              = "tf-example-calendar-with-holidays"
  comments          = "Business calendar with holiday exceptions"
  business_days     = "Monday,Tuesday,Wednesday,Thursday,Friday"
  first_day_of_week = "Monday"

  custom_days = [stonebranch_custom_day.new_years.name]
}

# Use calendar in a time trigger
resource "stonebranch_trigger_time" "daily_with_calendar" {
  name        = "tf-example-trigger-with-calendar"
  description = "Trigger that uses a business calendar"

  tasks = ["some-existing-task"]

  time       = "09:00"
  time_zone  = "America/New_York"
  day_style  = "Everyday"

  # Reference the business calendar
  calendar = stonebranch_calendar.business.name
}
