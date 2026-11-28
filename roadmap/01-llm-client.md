# M01: Talking to an LLM

**Guide:** §5 step 1, §5ক (Ollama adapter) · **Week:** 1 · **Depends on:** M00

## Concept

An LLM API is a **stateless function**: you send it the *whole* conversation (system prompt + messages + tool definitions) and it returns *one* reply. It has no memory between calls. Whatever "memory" the agent seems to have is your code re-sending the history every time. That has two consequences:

- **Cost and latency grow with the history.** Every call pays for every token again.
- **You own the conversation state.** That is why later milestones can store it, trim it, or switch providers in the middle of a conversation.

A reply is made of **content blocks**: `text` (the model is talking) or `tool_use` (the model wants your code to run a tool). The **`stop_reason`** field tells you why the model stopped:
- `end_turn`: it is done.
- `tool_use`: it is waiting for tool results.
- `max_tokens`: it was cut off.

**Tokens** are word pieces, and both price and context limits are counted in tokens.

## How industry does it

Production agents almost always put a **provider-neutral layer** between the agent and the vendor APIs. Vendors change APIs, prices and model names, and companies switch providers for cost or privacy reasons. Within this project, only the adapter packages know any vendor's JSON format. The guide uses Anthropic's content-block format as the neutral format, and each adapter translates to and from it. This is the **Adapter pattern**, and the agent depends only on the `Provider` **interface** (dependency inversion).

## Go you'll use

> **JSON v2 (ADR 0001):** this project uses `encoding/json/v2` + `encoding/json/jsontext`, not the
> guide's v1. Read `json.RawMessage` below as **`jsontext.Value`**, prefer the `omitzero` tag option,
> and remember v2 matches field names **case-sensitively** by default.

- **Struct tags and `omitempty`:** `json:"tool_use_id,omitempty"` controls the field name and leaves the field out when it is empty.
- **`json.RawMessage`:** keeps tool input/schema JSON as raw bytes, because its shape isn't known ahead of time.
- **Interfaces:** `type Provider interface { Send(ctx, Request) (*Response, error) }`. Any type that has this method satisfies the interface automatically; Go has no `implements` keyword.
- **`http.NewRequestWithContext`:** the HTTP request is cancelled when the `ctx` is cancelled.
- **`sync/atomic`:** generates tool-call IDs safely across goroutines, needed because Ollama may omit IDs.
- **`net/http/httptest`:** a real HTTP server inside your test, so you can test an adapter without the network.

## Tasks

### Task 1: Neutral types (`internal/llm/types.go`)
- **Claude scaffolds:** the file with `ContentBlock`, `Message`, `Tool`, `Request`, `Usage`, `Response`, and the `Provider` interface (guide step 1), plus constants for roles, block types and stop reasons (`StopToolUse`, `StopEndTurn`, ...) so code never compares magic strings.
- **Claude scaffolds a test:** `types_test.go` marshals a `Request` and compares it with `testdata/request.golden.json`.
- **You implement:** make the golden test pass. Then explain why `Input` is a `json.RawMessage` and not a `map[string]any`.

### Task 2: The `APIError` type
- An improvement over the guide, which formats status codes into strings: `type APIError struct { Provider string; Status int; Body string }` with an `Error()` method, plus `func (e *APIError) Retryable() bool` (true for 429/5xx; named `Retryable`, not `Temporary`, see notes/m01.md).
- **You implement:** the struct, the methods, and a table-driven test covering 200/400/401/404/429/500/502/503/529.

### Task 3: The Ollama adapter (`internal/llm/ollama.go`)
- **Claude scaffolds:** the `Ollama` struct, the `ollama*` wire types, the function signatures `toOllama(Request) ollamaRequest` and `fromOllama(ollamaResponse) *Response`, and `ollama_test.go`, which uses `httptest` to check the request body and to return a canned tool-call response.
- **You implement:** both translation functions and `Send`:
  - Turn `tool_result` blocks into `role:"tool"` messages.
  - Generate tool-call IDs, keeping track of which ID maps to which tool name.
  - Map to `StopToolUse` when `tool_calls` is non-empty (Ollama 0.40 sends `done_reason: "stop"` even for tool calls, see notes/m00.md §6), `StopMaxTokens` on `done_reason: "length"`, otherwise `StopEndTurn`.
  - Use the tool-call `id` when Ollama sends one; generate an ID only when it's missing.
  - Return `*APIError` for status codes other than 200.
- **Done when:** the contract tests pass, including one where Ollama returns two tool calls in a single reply.

### Task 4: First real call (`cmd/hello/main.go`)
- **You implement:** a 20-line program that sends "Introduce yourself in one line" to the local `qwen3:14b` and prints the text and the token usage. Run it twice to see for yourself that the API is stateless.
- **Experiment (guide exercise):** set `System` to "You are an angry teacher" and compare the answers.

## Improvements over the guide's code

| Guide | Industry upgrade | Why |
| --- | --- | --- |
| `fmt.Errorf("llm api %d: %s")` | typed `*APIError` + `errors.As` | retry logic shouldn't parse strings |
| magic strings `"tool_use"` | typed constants | typos become compile errors |
| `resp.Content[0].Text` | a `Text()` helper that skips non-text blocks | the first block may be a tool call or thinking output, not text |
| no response size limit | `io.LimitReader` on error bodies | a misbehaving server can't flood memory |

## Check yourself

1. If the API is stateless, how does the model "know" what the user said two turns ago?
2. What breaks if the agent package imports Ollama's wire types directly?
3. Why test the adapter against `httptest` rather than against the real Ollama?
4. What should the agent do when `stop_reason` is `max_tokens`?
