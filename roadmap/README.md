# Ops Agent Roadmap

This roadmap takes you from an empty repo to a production-style AI agent written in Go, built from scratch with no agent framework. It follows `GO-AI-AGENT.md` (the guide) step by step, and adds the engineering practices that real teams use when they ship agents.

## How we work: pair mode

For every task:

1. **Read** the milestone's *Concept* and *Go you'll use* sections before writing any code.
2. **Claude scaffolds** the parts listed as "Claude scaffolds": package layout, types and interfaces, function signatures with `// TODO(you):` markers, and **failing tests** that describe the behavior.
3. **You implement** the parts listed as "You implement" until `make check` passes.
4. **Review:** Claude reviews your code the way a senior engineer would. You fix what comes up, then make one commit for the task.
5. Answer the milestone's **Check yourself** questions out loud or in `notes/`. If you can't answer one, go back over that part before moving on.

To start a task, say: **"Start M01 task 2"**.

## The big picture

```
                 ┌──────────────────────────── Ops Agent (Go binary) ───────────────────────────┐
 user ── CLI/HTTP ──► Agent Loop ──► Context Builder ──► llm.Provider ──► Ollama / Claude / OpenAI
                 │       │   ▲                          (interface)      (adapters, Fallback)     │
                 │       ▼   │ tool_result                                                        │
                 │   Guardrails (approval, budget, timeouts) ──► Tool Registry ──► query_db       │
                 │       │                                                       search_logs      │
                 │       ▼                                                       search_docs (RAG)│
                 │   Memory (SQLite sessions + vectors)                          create_ticket ⚠  │
                 │   Observability (slog JSON, OpenTelemetry → Jaeger)                            │
                 └────────────────────────────────────────────────────────────────────────────────┘
```

The model only **decides** what to do: which tool to call and with what input. Your Go code does all of the actual **work**, and it is also your Go code that enforces the **rules**. Every milestone below adds one of those boxes.

## Milestones

The order follows the guide's 12-week plan (§10). There are two deliberate changes. **M00** comes first, because a professional project starts with its foundations. **Loop unit tests** move up to M03, because industry teams test the agent loop from the first day; only *quality evals* stay in M09.

| # | Milestone | Guide | Week | Status |
| --- | --- | --- | --- | --- |
| [M00](00-foundations.md) | Engineering foundations: git, Makefile, lint, CI, ADRs | §3 | 0 | ✅ |
| [M01](01-llm-client.md) | Talking to an LLM: neutral types, `Provider`, Ollama adapter | §5 step 1, §5ক | 1 | ☐ |
| [M02](02-tool-registry.md) | Tools: schema, descriptions, registry | §5 step 2 | 2 | ☐ |
| [M03](03-agent-loop.md) | The agent loop + REPL + loop unit tests | §5 step 3, §7 step 8a | 2 | ☐ |
| [M04](04-real-tools.md) | Real tools: `query_db`, `search_logs`, `create_ticket` | §6 step 4 | 3 | ☐ |
| [M05](05-multi-provider.md) | Claude + OpenAI-compatible adapters, factory, Fallback | §5ক | 4 | ☐ |
| [M06](06-session-memory.md) | Session memory in SQLite, history compaction | §6 step 5a | 5 | ☐ |
| [M07](07-rag.md) | RAG: chunking, embeddings, vector search, `search_docs` | §6 step 5b | 6 | ☐ |
| [M08](08-guardrails.md) | Guardrails: approval, budgets, retries, prompt injection | §6 step 6 | 7 | ☐ |
| [M09](09-evals.md) | Quality evals: cases, runner, local vs paid comparison | §7 step 8b | 8 | ☐ |
| [M10](10-observability.md) | Observability: structured logs, cost, OpenTelemetry | §7 step 9 | 9 | ☐ |
| [M11](11-mcp.md) | MCP: expose tools as an MCP server, add an MCP client | §7 step 7 | 10 | ☐ |
| [M12](12-service-and-agentops.md) | HTTP service, auth, `agent.yaml`, kill switch, deploy | §7 step 10a, §8 | 11 | ☐ |
| [M13](13-multi-agent.md) | Multi-agent: router, orchestrator–worker, A2A | §7 step 10b, §9 | 12 | ☐ |

Change ☐ to ✅ when a milestone meets its Definition of Done (see [ENGINEERING.md](ENGINEERING.md)).

## Files in this folder

- `ENGINEERING.md`: the standards that apply to every task (git workflow, testing, errors, the Definition of Done).
- `NN-*.md`: one file per milestone. Each has a concept explanation, how industry does it, the Go you'll use, the tasks, improvements over the guide's code, and self-check questions.
- `notes/`: create it for your own answers and learnings. Writing them down is part of the learning (guide §10: "teach it to learn it").
