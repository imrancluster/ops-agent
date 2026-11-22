# 0001. Use net/http for LLM APIs; no SDK or agent framework

- **Status:** Accepted
- **Date:** 2026-11-22
- **Deciders:** Imran Sarder

## Context

An SDK or agent framework would help us build an agent faster. However, the main goal of this
project is to understand how an agent works internally, so speed of delivery matters less than
visibility. Other forces:

- **Three different providers.** Ollama, Claude and OpenAI-compatible APIs each use a different
  JSON format. No single SDK covers all three, so we would need three SDKs anyway.
- **Go's standard library is enough.** `net/http` and `encoding/json/v2` cover everything these APIs need.
- **Dependencies are a risk.** Every SDK brings in transitive packages that we must update and
  scan with `govulncheck`.

## Options considered

1. **Agent framework (LangChain-style):** fastest start, but it hides the loop, tool calling and
   history handling, which are exactly what we want to learn. It is a heavy dependency, and the
   Go ecosystem for these frameworks is less mature than Python's.
2. **Official vendor SDKs (Anthropic / OpenAI Go SDK):** typed, maintained, and they handle retries
   and streaming. But each one covers only its own vendor (and Ollama has its own API), so we would
   still need our own neutral layer, and the SDK hides the raw JSON that we want to see.
3. **Raw `net/http` + `encoding/json/v2`:** full visibility into every request and response, zero
   dependencies, and one neutral format. We write and maintain the adapters ourselves.

## Decision

We will call LLM APIs directly with `net/http` and `encoding/json/v2` (option 3). All vendor-specific
code stays inside the `internal/llm` adapters behind the `Provider` interface, so replacing an
adapter with an SDK later changes one file and nothing else.

We use `encoding/json/v2` because it is stable in our Go floor (1.27) and is the
standard library's future direction. The guide's example code uses v1, so we adapt it
as we go (see the costs below).

## Consequences

- **Good:**
  - We see every request and response, so we can debug problems at the JSON level.
  - No third-party dependencies for LLM access: nothing extra to update or scan.
  - One neutral message format works across all providers, and we can switch providers
    in the middle of a conversation.
  - After building it ourselves, we will understand what an SDK or framework does for us,
    which helps us choose one wisely in future projects.
- **Bad / accepted costs:** we now own work that an SDK would do for us:
  - Retries, backoff and timeouts (M08).
  - Keeping up with API changes (new fields, version headers, renamed parameters such as
    `max_completion_tokens`).
  - Contract tests for every adapter, because nobody else tests our JSON.
  - Streaming, if we ever need it, is harder to write by hand.
  - The guide's code uses `encoding/json` (v1), so we translate it to v2: for example,
    `json.RawMessage` becomes `jsontext.Value`, `omitzero` replaces most uses of
    `omitempty`, and field-name matching is case-sensitive by default. Golden-file
    contract tests catch any difference in the JSON we send.
- **Revisit when:**
  - The learning milestones are finished and the project moves toward production use; or
  - We need streaming, prompt caching or batch APIs on two or more providers; or
  - Keeping the adapters in sync with API changes costs more than about one day per month.
