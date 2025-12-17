---
sidebar_position: 2
---

# LLM Streaming

Validate structured output from LLMs as it streams.

## Use Case

When using OpenAI, Anthropic, or other LLMs with structured output, you receive JSON token-by-token. Pedantigo can validate partial JSON as it arrives.

```go
type LLMResponse struct {
    Answer     string   `json:"answer" validate:"required"`
    Confidence float64  `json:"confidence" validate:"ge=0,le=1"`
    Sources    []string `json:"sources" validate:"max_items=10"`
}

parser := pedantigo.NewStreamParser[LLMResponse]()

// Feed tokens as they stream from LLM
for token := range llmStream {
    parser.Feed(token)
}

response, errs := parser.Complete()
```

TODO: Add complete LLM streaming example.

:::tip
Streaming validation catches malformed JSON early, before the full response arrives.
:::
