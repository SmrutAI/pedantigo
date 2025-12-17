---
sidebar_position: 3
---

# Format Constraints

Constraints for common formats.

| Constraint | Description |
|------------|-------------|
| `email` | Valid email address |
| `url` | Valid URL |
| `uuid` | Valid UUID |
| `uri` | Valid URI |
| `ipv4` | Valid IPv4 address |
| `ipv6` | Valid IPv6 address |

## Example

```go
type Contact struct {
    Email   string `validate:"required,email"`
    Website string `validate:"url"`
    ID      string `validate:"uuid"`
}
```

TODO: Add complete format constraints documentation.

:::tip
Format constraints use standard patterns and are optimized for performance.
:::
