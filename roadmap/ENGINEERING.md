# Engineering Standards

These rules apply to every milestone. They are what separates a demo agent from one a company would run in production. The guide makes the same point (§2): about 80% of agent engineering is ordinary good software engineering.

## 1. Repository and git workflow

- **Trunk-based development.** `main` must always build and pass its tests. Do each task on a short-lived branch (`m01-ollama-adapter`) and merge it when it's green.
- **Conventional Commits**, one logical change per commit:
  `feat(llm): add Ollama adapter`, `test(agent): cover MaxSteps exhaustion`, `refactor(tools): ...`, `docs(roadmap): ...`, `chore(ci): ...`.
- Never commit secrets. `.env` is in `.gitignore`, and `.env.example` documents every variable with fake values.
- **ADRs** (Architecture Decision Records) go in `docs/adr/NNNN-title.md`. Write one whenever you pick between real alternatives, for example "no LLM SDK", "pure-Go SQLite", or "neutral message format = Anthropic-style blocks". Each ADR is short: context, decision, consequences. Industry teams use ADRs so that six months from now nobody has to ask "why did we do this?"

## 2. Code layout and dependency direction

```
cmd/agent        → wires everything together (the only place that reads env/config)
internal/agent   → depends on llm, tools, memory interfaces
internal/tools   → depends on llm (for llm.Tool)
internal/llm     → depends on nothing internal
internal/memory  → depends on llm (for llm.Message)
```

- Dependencies only point **inward**, toward `llm`. `llm` must never import `agent`. If a cycle shows up, the design is wrong.
- **Accept interfaces, return structs.** The consumer defines the interface it needs, and keeps it small (`Provider` has a single method).
- **No package-level mutable globals.** Build dependencies in `main` and pass them in (dependency injection by hand).
- **Configuration is parsed and validated once at startup, and a bad config fails fast.** Library packages never call `os.Getenv`.

## 3. Errors

- Wrap with context: `fmt.Errorf("ollama chat: %w", err)`. Never discard an error silently. If you deliberately ignore one, say why in a comment.
- Use **typed errors** for errors callers need to branch on (`*llm.APIError{Status int}`), and check them with `errors.As` / `errors.Is`, never by matching strings.
- **Agent-specific rule:** an error from a *tool* is **data for the model** (an `is_error` tool result), not a crash. An error from the *LLM provider* or from *infrastructure* goes up to the caller.

## 4. Context, timeouts, concurrency

- Every function that does I/O takes `ctx context.Context` as its first parameter and respects cancellation.
- Every outbound call has a deadline: the LLM client, each tool call, each HTTP request.
- Any goroutine you start must have a clear way to stop. Use `errgroup` for fan-out work.
- Run `go test -race` in CI. Anything shared between sessions (HTTP service, M12) needs a mutex or must be owned by a single goroutine.

## 5. Testing strategy (the agent testing pyramid)

| Layer | What | Tool | Cost |
| --- | --- | --- | --- |
| Unit | Pure functions, registry, loop logic with `FakeProvider` | `testing`, table-driven tests | free, deterministic |
| Contract | Each adapter's JSON in/out against a fake server | `net/http/httptest` + golden files in `testdata/` | free, deterministic |
| Integration | Real SQLite, real Ollama (opt-in) | build tag `//go:build integration` | local only |
| Evals | Quality against real models | `cmd/eval`, `evals/cases.json` | slow, costs money on paid APIs |

- `go test ./...` must **never** need the network or a model. Real-model tests sit behind `-tags=integration`.
- When tests call randomness or time, inject a clock or seed so the results are deterministic.

## 6. Observability and security defaults

- Use `log/slog` only, with JSON in production. Log events with fields (`"tool", name, "ms", d`), not formatted sentences.
- Never log API keys. Redact emails and phone numbers in anything that is logged or stored.
- Treat all tool output, logs and documents as **untrusted data**, never as instructions.
- Every tool gets the least privilege it needs, enforced in code (read-only DB connection, fixed file paths, domain allowlists). The prompt alone is never the enforcement.

## 7. Tooling (`make` targets, set up in M00)

| Target | Does |
| --- | --- |
| `make fmt` | `gofmt` + `goimports` |
| `make lint` | `golangci-lint run` |
| `make test` | `go test -race -cover ./...` |
| `make vuln` | `govulncheck ./...` |
| `make check` | fmt-check + vet + lint + test (CI runs exactly this) |
| `make run` | load `.env`, `go run ./cmd/agent` |

## 8. Definition of Done (for every task)

- [ ] `make check` is green.
- [ ] New behavior has tests, and the failure paths are tested too, not just the happy path.
- [ ] Every exported identifier has a doc comment that explains *why*, not just *what*.
- [ ] `CLAUDE.md` / an ADR is updated if the architecture or a decision changed.
- [ ] There is one Conventional Commit for the task.
- [ ] You can explain the code in your own words, without reading it.
