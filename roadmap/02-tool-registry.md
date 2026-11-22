# M02: Tools

**Guide:** §5 step 2 · **Week:** 2 · **Depends on:** M01

## Concept

A model can't *do* anything; it can only produce text. **Tool calling** is a contract between the model and your code:

1. You send the model a list of tools, each with a **name**, a **description**, and a **JSON Schema** for its input.
2. The model replies with a `tool_use` block: "call `get_current_time` with `{}`".
3. **Your code** decides whether to run it, runs it, and sends the result back.

Because your code runs the tool, **you have full control**. This is the most important security idea in agent engineering, and M08 builds on it.

The model never sees your Go code. It sees **only the description**, so the description is half of the tool. A good description says what the tool does, **when to call it**, and **when not to**.

## How industry does it

- **Few, robust tools** beat many small ones (guide §6 rule 1). Small local models make bad choices once they have more than about six tools.
- Tool definitions are kept **next to** their implementation (a registry), so the schema and the code can't drift apart.
- The tool list goes to the model in a **stable order**. That keeps behavior reproducible and lets providers cache the prompt prefix, which cuts cost.
- Tool descriptions are versioned and tested like prompts. Changing one changes the agent's behavior.

## Go you'll use

- **Function types:** `type Func func(ctx context.Context, input json.RawMessage) (string, error)`. A tool's behavior is a value you can store in a struct.
- **Closures:** a tool constructor (`QueryDBTool(db)`) returns a `Func` that "captures" its dependencies. You'll use this a lot in M04.
- **Maps + `slices.SortFunc`:** map iteration order is random in Go, so you must sort.
- **`time.LoadLocation` + `_ "time/tzdata"`:** time zone data embedded in the binary (Windows tip from the guide).

## Tasks

### Task 1: Registry (`internal/tools/registry.go`)
- **Claude scaffolds:** `Tool`, `Func`, `Registry` with `Add`, `Get`, `Defs`, and `registry_test.go`.
- **You implement:**
  - `Add` returns an **error** on a duplicate or empty name. This is an improvement: the guide silently overwrites.
  - `Defs()` returns the tools sorted by name.
- **Done when:** the tests cover duplicates, ordering, and `Get` on an unknown name.

### Task 2: First tool (`internal/tools/time.go`)
- **Claude scaffolds:** the `TimeTool(now func() time.Time)` signature. The clock is **injected** so the test is deterministic.
- **You implement:** the tool, returning RFC 3339 time in Asia/Dhaka. Test it with a fixed clock.
- **Done when:** the test passes. Then explain why `time.Now` is not called directly.

### Task 3: Schema validation helper
- **Claude scaffolds:** `func decodeInput[T any](raw json.RawMessage) (T, error)`, a generic helper that uses `json.Decoder` with `DisallowUnknownFields`.
- **You implement:** the helper plus a test with malformed JSON and with unknown fields. Every later tool will use it, because models sometimes send broken JSON.

### Task 4: Writing a description
- **You implement:** in `notes/m02.md`, write two descriptions for an imaginary `get_weather` tool, a bad one and a good one. Say what makes the good one better: when to call, when not to, and units.

## Improvements over the guide's code

| Guide | Upgrade | Why |
| --- | --- | --- |
| `Add` overwrites silently | returns an error | two tools with the same name means a silent behavior bug |
| `time.Now()` inside the tool | injected clock | deterministic tests |
| `json.Unmarshal` into an ad-hoc struct | generic `decodeInput[T]` with strict decoding | one validated path for every tool |

## Check yourself

1. Who executes a tool: the model, or your program? Why does that matter for security?
2. Why must `Defs()` return a stable order?
3. The guide says to write descriptions in English even when answers are in Bangla. Why?
4. What happens if two tools have overlapping descriptions?
