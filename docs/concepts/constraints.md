---
sidebar_position: 2
---

# Constraints

Constraints are validation rules applied via struct tags.

## Syntax

```go
type Example struct {
    Field string `validate:"constraint1,constraint2=value"`
}
```

## Common Constraints

| Constraint | Description |
|------------|-------------|
| `required` | Field must not be empty |
| `min=N` | Minimum value/length |
| `max=N` | Maximum value/length |
| `email` | Valid email format |

TODO: Add full constraints documentation.

:::info
See [Constraints Reference](/docs/constraints/string) for the complete list.
:::
