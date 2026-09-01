# Stonebranch Custom Day Example

This example demonstrates how to create custom days in Stonebranch Universal Controller. Custom days define calendar exception dates and can be referenced by calendars, or marked as holidays with weekend-observance rules.

## Resources Created

- `stonebranch_custom_day.maintenance_window` - `Single Date`: one specific exception date
- `stonebranch_custom_day.company_closures` - `List of Dates`: an explicit list of exception dates
- `stonebranch_custom_day.christmas` - `Absolute Repeating Date`: the same month+day every year, marked as a holiday with weekend-observance rules
- `stonebranch_custom_day.thanksgiving` - `Relative Repeating Date`: the nth weekday of a month, every year (e.g. 4th Thursday of November)

## Usage

```bash
# Set your API token
export STONEBRANCH_API_TOKEN="your-token"

# Initialize and apply
terraform init
terraform plan
terraform apply
```

## Attributes

| Attribute | Description | Required |
|-----------|-------------|----------|
| `name` | Unique name of the custom day | Yes |
| `ctype` | Definition style: `Single Date`, `List of Dates`, `Absolute Repeating Date`, or `Relative Repeating Date` | Yes |
| `date` | Specific date (`yyyy-MM-dd`), used when `ctype` is `Single Date` | No |
| `date_list` | List of specific dates (`yyyy-MM-dd`), used when `ctype` is `List of Dates` | No |
| `month` | Month (`Jan`-`Dec`), used when `ctype` is `Absolute Repeating Date` or `Relative Repeating Date` | No |
| `day` | Day of month, used when `ctype` is `Absolute Repeating Date` | No |
| `dayofweek` | Day of week (`Sun`-`Sat`), used when `ctype` is `Relative Repeating Date` | No |
| `relfreq` | Relative frequency (`1st`, `2nd`, `3rd`, `4th`, `Last`, `Every`, `Nth`, `Last Day`, `Last Business Day`), used when `ctype` is `Relative Repeating Date` | No |
| `nth_amount` / `nth_type` | Nth day-of-month value/type, used when `relfreq` is `Nth` | No |
| `adjustment` / `adjustment_amount` / `adjustment_type` | Offset applied to the resolved date (`None`, `Less`, `Plus`) | No |
| `holiday` | Marks this custom day as a holiday, enabling `observed_rules` | No |
| `period` | Marks this custom day as a period (not allowed when `ctype` is `Single Date`) | No |
| `observed_rules` | Weekend-observance rules (`actual_day_of_week` / `observed_day_of_week`), used when `holiday` is `true` | No |
| `comments` | Description of the custom day | No |

## Notes

- The `ctype` field groups above are mutually exclusive; the provider validates the correct field combination for the selected `ctype` at plan time.
- `category` is a server-computed, read-only attribute (`Day`, `Holiday`, or `Period`) derived from the `holiday`/`period` flags.
- The `ctype`/`relfreq` enum values are not documented in Stonebranch's OpenAPI spec; they were determined by empirically probing a live UAC instance.
