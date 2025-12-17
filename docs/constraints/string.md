---
sidebar_position: 1
---

# String Constraints

Constraints for string fields.

| Constraint | Example | Description |
|------------|---------|-------------|
| `min=N` | `min=2` | Minimum length |
| `max=N` | `max=100` | Maximum length |
| `len=N` | `len=10` | Exact length |
| `pattern=regex` | `pattern=^[a-z]+$` | Regex pattern |

## Example

```go
type Profile struct {
    Username string `validate:"required,min=3,max=20,pattern=^[a-z0-9_]+$"`
    Bio      string `validate:"max=500"`
}
```

TODO: Add complete string constraints documentation.
