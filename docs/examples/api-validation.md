---
sidebar_position: 3
---

# API Validation

Using Pedantigo in REST API handlers.

## HTTP Handler Example

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    user, errs := pedantigo.Unmarshal[CreateUserRequest](body)
    if len(errs) > 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]any{
            "errors": errs,
        })
        return
    }

    // user is validated, proceed with business logic
    // ...
}
```

## With Popular Frameworks

Works with Gin, Echo, Fiber, Chi, and any other Go HTTP framework.

TODO: Add framework-specific examples.

:::info
Pedantigo integrates seamlessly with any HTTP framework - just pass the request body to `Unmarshal`.
:::
