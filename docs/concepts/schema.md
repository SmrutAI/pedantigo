---
sidebar_position: 5
---

# JSON Schema Generation

Auto-generate JSON schemas from your Go structs.

## Example

```go
schema := pedantigo.Schema[User]()
jsonBytes, _ := json.MarshalIndent(schema, "", "  ")
fmt.Println(string(jsonBytes))
```

## Caching

Schemas are cached automatically - first call takes ~10ms, subsequent calls `<100ns` (240x faster).

TODO: Add detailed schema generation documentation.

:::tip
Use generated schemas for API documentation, form validation, or LLM structured output.
:::
