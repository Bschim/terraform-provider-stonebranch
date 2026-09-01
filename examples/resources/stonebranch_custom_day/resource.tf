# Stonebranch Custom Day Example
#
# This example demonstrates the four "ctype" definition styles supported by
# custom days, plus a holiday with weekend-observance rules.
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

# ctype = "Single Date": a single, specific exception date.
resource "stonebranch_custom_day" "maintenance_window" {
  name  = "tf-example-cd-maintenance"
  ctype = "Single Date"
  date  = "2026-04-15"
}

# ctype = "List of Dates": an explicit list of exception dates.
resource "stonebranch_custom_day" "company_closures" {
  name      = "tf-example-cd-closures"
  ctype     = "List of Dates"
  date_list = ["2026-07-03", "2026-11-27", "2026-12-26"]
}

# ctype = "Absolute Repeating Date": the same month+day every year.
resource "stonebranch_custom_day" "christmas" {
  name    = "tf-example-cd-christmas"
  ctype   = "Absolute Repeating Date"
  month   = "Dec"
  day     = 25
  holiday = true

  # If Christmas falls on a weekend, observe it on the nearest weekday.
  observed_rules = [
    {
      actual_day_of_week   = "Sat"
      observed_day_of_week = "Fri"
    },
    {
      actual_day_of_week   = "Sun"
      observed_day_of_week = "Mon"
    },
  ]
}

# ctype = "Relative Repeating Date": the nth weekday of a month, every year.
resource "stonebranch_custom_day" "thanksgiving" {
  name      = "tf-example-cd-thanksgiving"
  ctype     = "Relative Repeating Date"
  month     = "Nov"
  dayofweek = "Thu"
  relfreq   = "4th"
  holiday   = true
}
