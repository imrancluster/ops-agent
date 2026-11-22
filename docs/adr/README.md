# Architecture Decision Records

An ADR records **one** significant decision: the context, the options, what we chose, and what it costs us.
Write one whenever we pick between real alternatives that would be expensive to reverse.

Rules:
- Number them in order; never reuse a number.
- **Never edit an accepted ADR's decision.** If we change our mind, write a new ADR and mark the old one
  `Superseded by NNNN`. The history of *why* is the whole point.
- Keep them short (one screen). Copy [0000-template.md](0000-template.md) to start.

| # | Decision | Status |
| --- | --- | --- |
| [0001](0001-no-llm-sdk.md) | Talk to LLM APIs with `net/http` + `encoding/json/v2`, no SDK or agent framework | Accepted |
| [0002](0002-local-first-ollama.md) | Local Ollama is the default provider; paid APIs are opt-in | Accepted |
