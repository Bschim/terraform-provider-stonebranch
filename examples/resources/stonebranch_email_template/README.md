# Stonebranch Email Template Example

This example demonstrates how to create email templates in Stonebranch Universal Controller. Email templates define reusable subject/body content and recipients for email notifications, sent via a `stonebranch_email_connection`.

## Resources Created

- `stonebranch_email_connection.notifications` - the SMTP connection used to send emails
- `stonebranch_email_template.job_failure` - a template with a `to` recipient
- `stonebranch_email_template.weekly_report` - a template with `cc`/`bcc` recipients only

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
| `name` | Unique name of the email template | Yes |
| `email_connection` | Name of the `stonebranch_email_connection` used to send emails from this template | Yes |
| `to` / `cc` / `bcc` | Comma-separated recipient lists. At least one of these must be set | No* |
| `subject` | Subject line of the email | No |
| `body` | Body content of the email | No |
| `reply_to` | Reply-To address | No |
| `description` | Description of the email template | No |
| `opswise_groups` | List of business service names this email template belongs to | No |

\* At least one of `to`, `cc`, or `bcc` must be set; the provider validates this at plan time.

## Notes

- The underlying UAC API field for `name` is `templateName`.
- `PUT` (update) is a full replace for this resource: omitting an optional field (e.g. `cc`) on an update clears it server-side, it does not leave the existing value untouched.
