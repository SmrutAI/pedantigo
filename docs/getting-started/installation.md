---
sidebar_position: 1
---

# Installation

## Requirements

- Go 1.21 or later

## Install

```bash
go get github.com/smrutai/pedantigo
```

## Verify Installation

```go
package main

import "github.com/smrutai/pedantigo"

func main() {
    // If this compiles, you're good!
    _ = pedantigo.New[struct{}]()
}
```

TODO: Add more detailed installation instructions.
