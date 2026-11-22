# M05: Multiple Providers

**Guide:** §5ক (factory, Claude, OpenAI-compatible, Fallback) · **Week:** 4 · **Depends on:** M01, M03

## Concept

Your agent now runs on a free local model. Real systems also need paid models, for harder reasoning, better Bangla, or a backup when the local machine is down. Because the agent only knows `llm.Provider`, adding a provider means adding **one adapter**, without touching the agent at all.

- **Factory:** `llm.New(Config)` turns configuration into a concrete provider.
- **Decorator:** `Fallback{Primary, Backup}` *is also* a `Provider`, so it wraps other providers without the agent noticing. Retries (M08), rate limiting and routing (M13) use the same pattern.
- **History is neutral**, so you can even switch providers in the middle of a conversation.

## How industry does it

- An **LLM gateway** layer gives every provider one interface, with fallbacks, retries, and cost tracking. Some companies build this as an internal service; you're building it as a package.
- **Contract tests per adapter** use recorded or golden JSON. Vendor APIs drift, and these tests catch it.
- Choosing between local and paid is **measured, not guessed**: the same eval set runs on each provider (M09).
- API keys come from the environment or a secret manager, and are never logged or included in prompts.

## Go you'll use

- **Composition:** a struct that holds interface values (`Fallback`) is itself an implementation of the interface.
- **Shared helpers:** `postJSON` is used by both Ollama and OpenAI. Extract it once there are two real users of it, not before.
- **Type switch / mapping:** OpenAI uses `tool_calls` + `role:"tool"` messages, and its arguments arrive as a **JSON string**, not an object.

## Tasks

### Task 1: Claude adapter
- **Claude scaffolds:** `claude.go`, plus contract tests checking the headers (`x-api-key`, `anthropic-version`) and parsing a tool-use response.
- **You implement:** `Send` (the format is nearly the identity mapping, since our types are Anthropic-shaped). A configurable base URL lets the tests point at `httptest`.

### Task 2: OpenAI-compatible adapter
- **You implement:** translation in both directions. The tricky parts:
  - The system prompt becomes a `system` message.
  - `arguments` is a string, so it has to be unmarshaled into a `RawMessage`.
  - `finish_reason` `tool_calls` maps to `StopToolUse`.
  - Some models want `max_completion_tokens` instead of `max_tokens`.

### Task 3: Factory and config
- **Claude scaffolds:** `Config`, `ConfigFromEnv(prefix)`, `New(Config)` with table tests (unknown provider, defaults, missing API key for a paid provider gives an error).
- **You implement:** `Validate()` so that a misconfiguration fails at startup instead of on the first request.

### Task 4: Fallback
- **You implement:** `Fallback.Send`. Tests: primary OK means the backup is never called; primary fails means the backup is called with `Model` cleared; **context cancelled** means *no* fallback (the user gave up, so a paid call would be wasted).
- Decide which errors justify a fallback (a 400 is your bug, so falling back won't help). Record the decision in an ADR.

### Task 5: Side-by-side check
- **You implement:** run the same question on Ollama and on Claude Haiku (`claude-haiku-5-5`). Write down quality, latency and tokens in `notes/m05.md`.

## Check yourself

1. Why does `Fallback` implement `Provider` instead of being an `if` statement inside the agent?
2. Why should a 400 error *not* trigger a fallback?
3. What has to stay true about the history format for a provider switch mid-conversation to work?
4. Where are the API keys read, and why only there?
