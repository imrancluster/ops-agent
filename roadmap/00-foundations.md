# M00: Engineering Foundations

**Guide:** §3 (lab setup) · **Week:** 0 · **Depends on:** nothing

## Concept: why start here?

Agents are *non-deterministic*: the same input can produce a different output on the next run. That makes everything else about the project matter more, not less. When the agent misbehaves you need a reproducible build, a fast test loop, and a git history that lets you answer "what changed?". Teams that skip foundations end up debugging the model when the real bug is in their code.

## How industry does it

- Every repo has a single command that answers "is this healthy?" (`make check`), and CI runs exactly that command.
- Linting is strict from day one, because tightening it later means a painful cleanup.
- Decisions are written down as ADRs. Agent projects make many "this model or that model, SDK or no SDK" decisions, and the reasons get forgotten fast.

## Go you'll use

- **Modules:** `go.mod` declares the module path and the **language floor** (`go 1.27.4`). Features newer than the floor are off-limits.
- **`go vet`:** the compiler's own checks for suspicious code. **`golangci-lint`** runs dozens of linters (errcheck, staticcheck, gosec, revive) in a single pass.
- **`govulncheck`:** reports known CVEs that your code *actually calls*, not just everything in the dependency list.

## Tasks

### Task 1: Version control
- **Claude scaffolds:** `.gitignore` (binaries, `.env`, `*.db`, coverage files, `.claude-flow/` hook state), `.editorconfig`.
- **You implement:** `git init`, set the default branch to `main`, make the first commit with `chore: initial scaffold`.
- **Done when:** `git status` is clean and `.env` is ignored. Check with `git check-ignore .env`.

### Task 2: Toolchain check
- Your `golangci-lint` 2.12.2 was **built with Go 1.26**. It may refuse to lint a `go 1.27.4` module.
- **You implement:** run `golangci-lint run ./...`. If it complains about the Go version, reinstall it with a build made by Go 1.27 or newer (`brew upgrade golangci-lint`, or `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`).
- **Done when:** `golangci-lint version` shows it was built with Go ≥ 1.27.

### Task 3: Makefile and lint config
- **Claude scaffolds:** `Makefile` with the targets listed in ENGINEERING.md §7, and `.golangci.yml` (v2 format) with errcheck, govet, staticcheck, revive, gosec, errorlint, bodyclose, noctx, contextcheck.
- **You implement:** read every linter in the config and write one line in `notes/m00.md` on what each one catches. Then run `make check`.
- **Done when:** `make check` passes on the empty project.

### Task 4: CI pipeline
- **Claude scaffolds:** `.github/workflows/ci.yml`, which sets up Go from `go.mod` and runs `make check` + `make vuln` on every push and PR.
- **You implement:** decide whether this repo lives on GitHub. If it does, create the remote yourself and push, then confirm the workflow passes.
- **Done when:** CI is green. If you don't use GitHub, `make check` works locally.

### Task 5: First ADRs and the env template
- **Claude scaffolds:** `docs/adr/0000-template.md`, `.env.example` (all `LLM_*` / `BACKUP_*` variables from guide §5ক, with fake values).
- **You implement:** write `docs/adr/0001-no-llm-sdk.md` (why raw `net/http`, guide §1) and `0002-local-first-ollama.md` (why Ollama is the default, guide §5ক).
- **Done when:** both ADRs explain their context, decision and consequences in your own words.

### Task 6: Lab check
- **You implement:** `curl http://localhost:11434/api/tags` lists `qwen3:14b` and `nomic-embed-text` (both are already pulled on this Mac). Send one test chat from the command line:
  `curl http://localhost:11434/api/chat -d '{"model":"qwen3:14b","messages":[{"role":"user","content":"hi"}],"stream":false}'`
- **Done when:** you have seen a raw JSON response and can point to `message.content`, `done_reason`, and the token counts (`prompt_eval_count`, `eval_count`).

## Check yourself

1. Why does `go test ./...` need to pass with no network or model running?
2. What is the difference between `go vet` and `golangci-lint`?
3. Why is `.env.example` committed when `.env` is not?
4. In the raw Ollama response, where would a tool call appear?
