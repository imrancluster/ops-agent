# M03: The Agent Loop

**Guide:** §5 step 3, plus §7 step 8a (FakeProvider tests, moved earlier) · **Week:** 2 · **Depends on:** M01, M02

## Concept: the heart of every agent

Every agent, from a toy one to the largest coding agent, runs this same loop:

```
history += user message
repeat up to MaxSteps:
    resp = LLM(system, history, tools)
    history += resp                      # keep the FULL assistant turn, including tool_use blocks
    if resp.stop_reason != "tool_use":   # model is done
        return text(resp)
    results = run each tool_use block
    history += user message containing ONLY tool_result blocks
fail: "could not finish in MaxSteps"
```

The guide predicts (§9) that future agents will run **the same loop**. What changes over time is what surrounds it: how long the agent runs, how much power it has, and how it is governed. Learn this loop properly and you have the foundation for everything else.

Three rules that cause API errors when broken:

1. The assistant turn must be stored **exactly** as received, including its `tool_use` blocks.
2. Every `tool_use` ID must get a matching `tool_result` in the **very next** user message.
3. That user message contains **only** `tool_result` blocks.

## How industry does it

- **Bounded autonomy:** a step limit (`MaxSteps`), per-tool timeouts, and later a token budget (M08). An agent stuck in a loop can burn through money overnight.
- **Errors are observations:** an unknown tool, a denied approval, or a failing tool all go back to the model as `is_error: true`, so the model can try another approach. The agent only fails on *infrastructure* errors.
- **Deterministic loop tests** use a scripted fake LLM. They are free, fast, and part of CI. This is how teams test agents without spending money on model calls.
- **The system prompt is code:** it lives in its own file, under git, and evals are rerun whenever it changes.

## Go you'll use

- **`context.WithTimeout` + `defer cancel()`:** a 30-second deadline per tool.
- **Slices of structs:** `history []llm.Message`. Note that `append` may reallocate.
- **`[]rune` truncation:** cuts strings by character, not by byte, so Bangla text isn't split in the middle of a character.
- **`bufio.Reader`:** reads REPL input line by line.
- **Test doubles:** `FakeProvider` implements `llm.Provider` with scripted responses and records the requests it receives.

## Tasks

### Task 1: FakeProvider (`internal/llm/llmtest/fake.go`)
- **Claude scaffolds:** the package and the type signature.
- **You implement:** `Send` returns the next scripted response, **records every request** so tests can inspect the history that was sent, and returns an error once the script is used up.

### Task 2: Loop tests first (TDD)
- **Claude scaffolds:** `agent_test.go` with failing, table-driven cases:
  - A tool call followed by an answer (guide's test).
  - An unknown tool, which must produce an `is_error` result and let the loop continue.
  - A risky tool that is denied.
  - A tool that returns an error.
  - Hitting `MaxSteps`.
  - Two tool calls in a single turn.
  - Very long output, truncated by rune count.
  - Context cancelled in the middle of the loop.
- **You implement:** nothing yet. Read every test and predict how it will fail.

### Task 3: The loop (`internal/agent/agent.go`)
- **Claude scaffolds:** the `Agent` struct and `Run`/`runTool` signatures with TODOs.
- **You implement:** `Run` and `runTool` until every Task 2 test passes.
- **Done when:** `go test -race ./internal/agent` is green.

### Task 4: The system prompt and REPL
- **Claude scaffolds:** `prompts/system.md` (guide §5), and `cmd/agent/main.go` wiring with config loaded once.
- **You implement:** the REPL loop with an approval prompt. Run `make run`, ask "What day is it today?", and watch the logs show `stop=tool_use` and then `stop=end_turn`.

## Improvements over the guide's code

| Guide | Upgrade | Why |
| --- | --- | --- |
| `MaxTokens: 2048` hard-coded | an `Agent` field with a default | different tasks need different limits |
| `max_tokens` stop treated as done | detected and reported | a cut-off answer is a bug, not an answer |
| `history` lives inside `Agent` | kept, but documented as **not goroutine-safe** (one Agent per session) | sets up the session design in M12 |
| tests in week 8 | tests now | loop bugs are the most expensive kind |

## Check yourself

1. Draw the `history` slice after the user asks "what day is it?" and the agent answers. How many messages does it hold, and with which block types?
2. Why is a failed tool returned to the model instead of being returned as a Go error?
3. What would happen if the assistant's `tool_use` turn were not stored in `history`?
4. Is this a workflow or an agent? Who decides the order of the steps (guide §1)?
