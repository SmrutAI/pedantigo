---
sidebar_position: 2
---

# Numeric Constraints

Constraints for numeric fields (int, float, etc.).

| Constraint | Example | Description |
|------------|---------|-------------|
| `min=N` | `min=0` | Minimum value |
| `max=N` | `max=100` | Maximum value |
| `gt=N` | `gt=0` | Greater than |
| `lt=N` | `lt=100` | Less than |
| `ge=N` | `ge=1` | Greater or equal |
| `le=N` | `le=99` | Less or equal |

## Example

```go
type Product struct {
    Price    float64 `validate:"gt=0"`
    Quantity int     `validate:"min=0,max=1000"`
    Rating   float64 `validate:"ge=0,le=5"`
}
```

TODO: Add complete numeric constraints documentation.
