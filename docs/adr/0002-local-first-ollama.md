# 0002. Local Ollama is the default LLM provider; paid APIs are opt-in

- **Status:** Accepted
- **Date:** 2026-11-22
- **Deciders:** Imran Sarder

## Context

The agent needs an LLM for every step of its loop, and during development we call it
constantly: manual testing, unit experiments, and later eval suites of 30+ cases run
several times each. Forces:

- **Cost.** Paid APIs charge per token, and an agent re-sends the whole history on every
  step. A runaway loop or a large eval run can cost real money overnight.
- **Privacy.** The agent reads the app's database and logs, which contain user data.
  With a local model, that data never leaves our machines.
- **Hardware we already own.** Mac-1 (Apple M5, 32 GB unified memory) runs a 14B model
  comfortably. `qwen3:14b` (about 9 GB on disk) is already installed and supports tool calling.
- **Quality gap.** Local models are weaker than frontier models at hard reasoning, long
  tool chains and Bangla. Some tasks will need a paid model.
- **Offline use.** A local-first agent keeps working with poor or no internet, which
  matters for deployments in rural Bangladesh (guide §9).

## Options considered

1. **Paid API only (e.g. Claude):** best quality and simplest setup, but every call costs
   money, user data leaves our network, and nothing works offline.
2. **Local only (Ollama):** free and private, but we accept lower quality with no way to
   escalate hard tasks, and the agent stops when Mac-1 is down.
3. **Local by default, paid as opt-in fallback, switchable through `.env`:** free and
   private for daily work; a paid model only when configured as the primary or as the
   backup. Costs us an extra adapter layer and the `Fallback` wrapper.

## Decision

We choose option 3. `LLM_PROVIDER` defaults to `ollama` with `qwen3:14b` and
`num_ctx` 16384. Paid providers (Claude, any OpenAI-compatible API) are enabled only
through `LLM_*` or `BACKUP_*` environment variables, with no code changes. The
`llm.Provider` interface from ADR 0001 is what makes this switch possible.

We picked the 14B size because it balances tool-calling quality against speed and memory
on Mac-1. 4B–8B models choose the wrong tool too often, and 30B models leave too little
memory for other work.

## Consequences

- **Good:**
  - Development and eval runs are free, so we can run them as often as we want.
  - App data stays on our own machines by default.
  - The agent works offline on the LAN.
  - Comparing local and paid models is a config change, so the M09 evals can measure the
    real quality gap instead of guessing it.
- **Bad / accepted costs:**
  - Lower quality on hard reasoning, long tool chains and Bangla answers.
  - The first call is slow while the model loads into memory (we use a 5-minute HTTP timeout).
  - A smaller context window: if `num_ctx` is too small, old messages are dropped silently.
  - Small models sometimes return a wrong tool name or broken JSON, so the agent loop must
    handle bad tool calls gracefully (M03).
  - We maintain three adapters instead of one, plus the fallback logic (M05).
  - Mac-1 becomes a dependency: if it is off, the agent needs a configured backup.
- **Revisit when:**
  - The M09 eval pass rate on `qwen3:14b` is more than 15 percentage points below a paid
    model on the same cases; or
  - A new local model clearly beats `qwen3:14b` on our evals (re-run evals, then switch
    the default); or
  - The agent runs for real users, where answer quality may matter more than cost; or
  - Paid API spending through the fallback goes above the daily budget we set in M08.
