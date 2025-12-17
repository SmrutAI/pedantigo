---
sidebar_position: 2
---

# Options

Configuration options for validators.

## Available Options

```go
validator := pedantigo.New[User](
    pedantigo.WithStrictMode(),
    pedantigo.WithCustomValidator("custom", customFn),
)
```

| Option | Description |
|--------|-------------|
| `WithStrictMode()` | Fail on unknown fields |
| `WithCustomValidator()` | Register custom validator |

TODO: Add complete options documentation.
