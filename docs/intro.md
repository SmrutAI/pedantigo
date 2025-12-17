---
sidebar_position: 1
slug: /
---

# Welcome to Pedantigo

**Pedantigo** is a Pydantic-inspired validation library for Go. It brings the elegance of Python's Pydantic to Go with struct tag-based validation, auto-generated JSON schemas, and high-performance caching.

## Why Pedantigo?

Pedantigo was built while developing [smrut.ai](https://smrut.ai), where reliable JSON validation is critical for AI agent interactions. We needed Pydantic's power in Go.

- **Familiar API**: If you've used Pydantic, you'll feel right at home
- **Struct Tags**: Go-idiomatic validation using struct tags
- **JSON Schema**: Auto-generate JSON schemas from your Go structs
- **Performance**: Built-in caching delivers 240x speedup on schema generation
- **Streaming**: Validate partial JSON as it streams (perfect for LLM outputs)
- **Discriminated Unions**: Type-safe union types with automatic discrimination

## Quick Example

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type User struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=0,max=150"`
}

func main() {
    jsonData := []byte(`{"name": "Alice", "email": "alice@example.com", "age": 30}`)

    // Unmarshal and validate in one call
    user, errs := pedantigo.Unmarshal[User](jsonData)
    if len(errs) > 0 {
        fmt.Println("Validation errors:", errs)
        return
    }

    fmt.Printf("Valid user: %+v\n", user)
}
```

## Features at a Glance

| Feature | Description |
|---------|-------------|
| **Struct Tags** | `validate:"required,email,min=18"` |
| **JSON Schema** | Auto-generated with constraints |
| **Caching** | 240x faster schema generation |
| **Streaming** | Validate partial JSON |
| **Unions** | Discriminated union types |
| **Cross-field** | Validate relationships between fields |

## Next Steps

- [Installation](./getting-started/installation) - Get Pedantigo installed
- [Quick Start](./getting-started/quickstart) - Build your first validated struct
- [Concepts](./concepts/validation) - Understand how Pedantigo works

:::tip Ready to dive in?
Start with the [Installation guide](./getting-started/installation) to get Pedantigo running in minutes.
:::
