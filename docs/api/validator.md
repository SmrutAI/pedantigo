---
sidebar_position: 1
---

# Validator API

Core validation functions.

## Simple API

```go
// Unmarshal JSON and validate
user, errs := pedantigo.Unmarshal[User](jsonData)

// Generate JSON schema
schema := pedantigo.Schema[User]()

// Validate existing struct
errs := pedantigo.Validate(user)
```

## Validator Object API

```go
// Create reusable validator
validator := pedantigo.New[User]()

// Use validator
user, errs := validator.Unmarshal(jsonData)
schema := validator.Schema()
```

TODO: Add complete API documentation.

:::info
See [pkg.go.dev](https://pkg.go.dev/github.com/smrutai/pedantigo) for full API reference.
:::
