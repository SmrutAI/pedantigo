---
sidebar_position: 3
---

# Errors

Validation error types and handling.

## ValidationError

```go
type ValidationError struct {
    Field   string // Field path (e.g., "user.email")
    Message string // Error message
    Tag     string // Constraint that failed
    Value   any    // Actual value
}
```

## Handling Errors

```go
user, errs := pedantigo.Unmarshal[User](jsonData)
for _, err := range errs {
    fmt.Printf("Field %s: %s\n", err.Field, err.Message)
}
```

TODO: Add complete error handling documentation.

:::caution
Always check the error slice - it may contain multiple validation failures.
:::
