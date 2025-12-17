---
sidebar_position: 1
---

# Basic Examples

Common validation patterns.

## User Registration

```go
type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"min=13"`
}
```

## Product Listing

```go
type Product struct {
    Name        string  `json:"name" validate:"required,max=200"`
    Description string  `json:"description" validate:"max=2000"`
    Price       float64 `json:"price" validate:"gt=0"`
    SKU         string  `json:"sku" validate:"required,pattern=^[A-Z0-9-]+$"`
}
```

TODO: Add more basic examples.
