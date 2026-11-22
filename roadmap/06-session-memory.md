# M06: Session Memory

**Guide:** §6 step 5 (part a) · **Week:** 5 · **Depends on:** M03

## Concept: three kinds of memory

| Kind | Remembers | How |
| --- | --- | --- |
| Short-term | this conversation | the `history` slice (already exists) |
| Session | the conversation, even after a restart | SQLite `sessions(id, history_json, updated_at)` |
| Knowledge | your docs and FAQs | RAG (M07) |

Memory is a *context engineering* problem: what should the model see, and when? The context window is limited and every token costs money. So long histories have to be **compacted**: older turns get summarized into one short message by the LLM.

The hard constraint is that **a `tool_use` and its `tool_result` must never be separated** when you trim. If they are, the API rejects the request.

## How industry does it

- **A storage interface plus implementations:** `memory.Store` with `Load`/`Save`, a SQLite implementation, and an in-memory one for tests.
- **Schema migrations** are versioned (numbered SQL files, applied at startup) and never edited by hand in production.
- **Compaction policies:** trigger on a token or message threshold, keep the last N turns verbatim, and summarize the rest. Many production agents also keep important facts in a separate long-term memory.
- **Durable execution** (guide §9): saving state after every step lets long-running agents resume after a crash. This milestone is the first step toward it.

## Go you'll use

- **`database/sql` transactions,** `ON CONFLICT ... DO UPDATE` (upsert).
- **`embed.FS`:** ship the migration `.sql` files inside the binary.
- **Interface design:** the consumer (the agent) defines the small interface it needs.
- **Pure functions for tricky logic:** `SafeCutIndex(history) int` is easy to test exhaustively.

## Tasks

### Task 1: Store interface and in-memory implementation
- **Claude scaffolds:** `internal/memory/store.go` (the interface), `memstore.go`, and a **shared conformance test suite** that runs against every implementation.
- **You implement:** the in-memory store until the suite passes.

### Task 2: SQLite store with migrations
- **Claude scaffolds:** `migrations/0001_sessions.sql` and a tiny migration runner signature.
- **You implement:** the runner (with a `schema_migrations` table) and `SQLiteStore`, which runs the same conformance suite.

### Task 3: Safe cut point
- **You implement:** `SafeCutIndex(history []llm.Message, keepLast int) int`, which never splits a `tool_use`/`tool_result` pair. It needs table tests with tricky cases: a pair at the boundary, several tool calls in one turn, and a history shorter than `keepLast`.

### Task 4: Compaction
- **You implement:** `Compact(ctx, provider, history)` summarizes everything before the cut into a single message. Test it with `FakeProvider`, then wire it into the agent behind a threshold.

### Task 5: Resume
- **You implement:** add a `--session <id>` flag to the REPL. Restarting the program with the same ID continues the conversation.

## Check yourself

1. Why does a longer history cost more on *every* call, not only once?
2. Show an example cut that would break the API, and explain why.
3. What do you lose when you summarize? When should a fact go into long-term memory instead?
4. Why does one conformance suite run against both stores?
