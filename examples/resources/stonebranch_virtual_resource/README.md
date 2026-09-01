# Stonebranch Virtual Resource Example

This example demonstrates how to create virtual resources in Stonebranch Universal Controller for concurrency control.

## Resources Created

- `stonebranch_virtual_resource.db_connections` - Renewable resource limiting concurrent database connections
- `stonebranch_virtual_resource.license_seats` - Depletable resource tracking a finite pool of license seats
- `stonebranch_task_unix.backup` - Example task referencing a virtual resource via a task variable

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
| `name` | Unique name of the virtual resource | Yes |
| `type` | Type of virtual resource: `Renewable`, `Boundary`, or `Depletable` | No |
| `limit` | Maximum concurrent usage allowed | No |
| `summary` | Description of the virtual resource | No |
| `opswise_groups` | Business services this virtual resource belongs to | No |

## Notes

- **API endpoint divergence**: the underlying UAC API endpoint for this resource is `/resources/virtual`, not `/resources/virtualresource`.
- `Renewable` resources release usage automatically once a task finishes (simple concurrency limiter).
- `Boundary` resources enforce mutual exclusion around a boundary condition.
- `Depletable` resources track a finite pool that is consumed and not automatically replenished (e.g. license seats).
- Actual task-to-virtual-resource binding (usage criteria) is configured on the task's runtime/resource criteria within UAC; this example only shows referencing the resource's name via a task variable.
