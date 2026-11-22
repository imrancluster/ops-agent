# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project state

This repo is an early scaffold for **Ops Agent**, a framework-free AI agent written in Go. Right now it holds only `go.mod` (module `github.com/imrancluster/ops-agent`, `go 1.27.4`), empty package directories (`cmd/agent`, `internal/{agent,llm,tools,memory}`), and `GO-AI-AGENT.md`.

`GO-AI-AGENT.md` (written in Bangla) is the design spec and step-by-step build guide. It contains reference code for each package. When you implement a step, read the matching section first. Two adjustments apply to its code:
- Its imports use `github.com/<your-name>/ops-agent/...`. Use the real module path `github.com/imrancluster/ops-agent/...`.
- It says its code is simplified for teaching. Check what you take from it with `go vet` and tests.

Guide sections by step: §5 steps 1–3 (LLM types, tool registry, agent loop, `main.go`), §5ক providers (Ollama, OpenAI-compatible, `.env`), §6 steps 4–6 (real tools, memory/RAG, guardrails), §7 steps 7–10 (MCP, evals, observability, multi-agent, deploy), §8 AgentOps (`agent.yaml`).

## Roadmap and working mode

The work follows `roadmap/` (`README.md` is the index, `ENGINEERING.md` holds the standards and Definition of Done, and `NN-*.md` has one file per milestone). The repo owner is learning agent engineering with basic Go experience, so we work in **pair mode**:
- Claude scaffolds what each task lists under "Claude scaffolds": types, interfaces, signatures with `// TODO(you):` markers, and **failing tests**.
- The owner implements what the task lists under "You implement".
- Claude then reviews like a senior engineer and explains the concepts.

Do not implement the "You implement" parts unless asked. When a milestone meets its Definition of Done, update its status in `roadmap/README.md`.

## Commands

```bash
go build ./...
go vet ./...
go test ./...
go test ./internal/agent -run TestAgentCallsToolThenAnswers   # single test
set -a; source .env; set +a; go run ./cmd/agent               # REPL; reads prompts/system.md from the CWD
```

Go does not load `.env` by itself; source it as shown above. `.env` holds API keys, so it must stay out of git.

## Architecture (as designed in the guide)

The agent is plain Go (`net/http` + `encoding/json`, no LLM SDK or framework) and talks to model APIs directly.

- **`internal/llm`**: a provider-neutral `Provider` interface (`Send(ctx, Request) (*Response, error)`). Its message format is modeled on Anthropic Messages content blocks (`text`, `tool_use`, `tool_result`, `stop_reason`), and this is the agent's internal language. Each adapter (`ollama.go`, `claude.go`, `openai.go` for any OpenAI-compatible API) translates to and from that format. The Ollama adapter generates tool-call IDs because Ollama may omit them. `provider.go` has the `New(Config)` factory, `ConfigFromEnv(prefix)`, a `Fallback{Primary, Backup}` wrapper, and the shared `postJSON` helper. An empty `Request.Model` means each provider uses its own default model.
- **Provider selection is env-driven and local-first.** `LLM_PROVIDER` / `LLM_MODEL` / `LLM_BASE_URL` / `LLM_API_KEY` / `LLM_NUM_CTX` choose the primary provider, which defaults to `ollama` at `localhost:11434` with `qwen3:14b` and `num_ctx` 16384. Setting the optional `BACKUP_*` variables wraps the primary in `Fallback`. Switching providers should never need code changes.
- **`internal/tools`**: a `Registry` of `Tool{Def llm.Tool, Run Func, Risky bool}`. `Defs()` returns tools sorted by name so their order stays stable across calls, which keeps behavior and caching predictable. Tool descriptions are written in English; the system prompt controls the answer language.
- **`internal/agent`**: `agent.go` runs the loop: send the request, append the full assistant content to history, and while `stop_reason == "tool_use"`, run the tools and append one user message that holds only `tool_result` blocks, up to `MaxSteps`. `runTool` returns unknown-tool errors, approval denials, and tool errors to the model as `is_error` results instead of failing. It applies a 30s timeout per tool and truncates output by rune count so Bangla text isn't split mid-character. `Risky` tools require the `Approve` callback, which is the human-in-the-loop gate. The guide also plans `context.go` (context builder) and `guard.go` (guardrails).
- **`internal/memory`**: SQLite session history plus an embedding vector store for RAG. RAG is exposed through a `search_docs` tool. When trimming or summarizing history, never separate a `tool_use` from its `tool_result`.
- **`prompts/system.md`**: the system prompt. It is versioned like code and evals are rerun after every change.
- **`evals/`**: two layers. Loop unit tests use a `FakeProvider` that returns scripted `[]*llm.Response`, with no model needed. Quality evals use `evals/cases.json`, where each case has `q`, `must_call`, `must_not_call`, and `answer_has`. A prompt or model change that lowers the pass rate does not ship.
- `cmd/agent/main.go` imports `_ "time/tzdata"` so that `time.LoadLocation("Asia/Dhaka")` works on Windows.

## Design constraints from the guide

- The DB tool is **read-only**. File access is limited to fixed paths, and outbound domains come from an allowlist.
- Tool output (logs, docs, DB rows) is data and never instructions. Defenses against prompt injection live in code (approval gates, allowlists) and must not rely on the prompt alone.
- Required limits: `MaxSteps`, a per-session token budget, per-tool timeouts, and limited exponential-backoff retries on 429/5xx. The guide suggests a typed `APIError{Status int}` checked with `errors.As` in place of matching on the error string.
- Logs use structured `log/slog`: session ID, step, model, tokens, duration, tool name, and errors. Phone numbers and emails are redacted in logs.
