---
sidebar_position: 6
---

# Streaming Validation

Validate partial JSON as it streams - perfect for LLM outputs.

## Use Case

When receiving JSON from an LLM token-by-token, you can validate the partial structure before it's complete.

```go
parser := pedantigo.NewStreamParser[User]()

// Feed tokens as they arrive
parser.Feed(`{"name": "Ali`)
parser.Feed(`ce", "email": "alice@`)
parser.Feed(`example.com"}`)

// Get validated result
user, errs := parser.Complete()
```

TODO: Add detailed streaming documentation.

:::info
Streaming validation is especially useful for real-time LLM applications.
:::
