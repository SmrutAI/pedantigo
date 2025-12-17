---
sidebar_position: 2
---

# Quick Start

This guide will help you validate your first struct with Pedantigo.

## Basic Example

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type User struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=0,max=150"`
}

func main() {
    jsonData := []byte(`{"name": "Alice", "email": "alice@example.com", "age": 30}`)

    user, errs := pedantigo.Unmarshal[User](jsonData)
    if len(errs) > 0 {
        fmt.Println("Validation errors:", errs)
        return
    }

    fmt.Printf("Valid user: %+v\n", user)
}
```

TODO: Add more quickstart content.
