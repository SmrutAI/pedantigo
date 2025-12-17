---
sidebar_position: 1
---

# Validation

Pedantigo validates Go structs using struct tags, similar to how `encoding/json` works.

## How It Works

1. Define a struct with `validate` tags
2. Call `pedantigo.Unmarshal[T]()` or `pedantigo.Validate()`
3. Get back validation errors (if any)

```go
type User struct {
    Name string `json:"name" validate:"required,min=2"`
}
```

TODO: Add detailed validation documentation.

:::tip
Validation happens automatically during `Unmarshal` - no separate step needed!
:::
