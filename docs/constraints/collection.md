---
sidebar_position: 4
---

# Collection Constraints

Constraints for arrays, slices, and maps.

| Constraint | Example | Description |
|------------|---------|-------------|
| `min_items=N` | `min_items=1` | Minimum items |
| `max_items=N` | `max_items=10` | Maximum items |
| `unique` | `unique` | All items unique |

## Example

```go
type Order struct {
    Items []string `validate:"min_items=1,max_items=50"`
    Tags  []string `validate:"unique"`
}
```

TODO: Add complete collection constraints documentation.
