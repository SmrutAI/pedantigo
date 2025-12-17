---
sidebar_position: 4
---

# Discriminated Unions

Type-safe union types with automatic discrimination based on a field value.

## Example

```go
type Message struct {
    Type    string `json:"type"`
    Content any    `json:"content"`
}

// Content can be TextContent, ImageContent, etc. based on Type
```

TODO: Add detailed discriminated unions documentation.

:::caution
Discriminated unions require careful type registration. See examples for patterns.
:::
