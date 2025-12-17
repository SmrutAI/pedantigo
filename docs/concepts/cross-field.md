---
sidebar_position: 3
---

# Cross-Field Validation

Validate relationships between multiple fields in a struct.

## Example

```go
type DateRange struct {
    StartDate time.Time `json:"start_date"`
    EndDate   time.Time `json:"end_date"`
}

func (d DateRange) Validate() error {
    if d.EndDate.Before(d.StartDate) {
        return errors.New("end_date must be after start_date")
    }
    return nil
}
```

TODO: Add detailed cross-field validation documentation.

:::note
Implement the `Validate() error` interface for custom cross-field logic.
:::
