# M09: Evals

**Guide:** §7 step 8 (part b) · **Week:** 8 · **Depends on:** M03–M08

## Concept

With ordinary code, a given input always produces the same output, so a unit test proves the behavior. With an agent it doesn't. A one-word change to the prompt, a new model version, or a reworded tool description can quietly break behavior that used to work. **Evals** are the agent's test suite: real questions paired with *checkable expectations*, run against a real model, and reported as a **pass rate**.

The guide's rule: **if the pass rate drops, the new prompt or model does not ship.** This one rule is what lets a team ship agent changes with confidence.

Kinds of checks, from cheapest to most expensive:

1. **Trajectory checks:** were the right tools called? Was a forbidden tool avoided (`must_call`, `must_not_call`)?
2. **Output checks:** does the answer contain the right facts (`answer_has`)? Numbers can be checked exactly against fixtures.
3. **LLM-as-judge:** a second model grades quality against a rubric. Check the judge's verdicts by hand now and then.

## How industry does it

- **Golden datasets** grow from production failures. Every 👎 or incident becomes an eval case (guide §8, feedback loop).
- **Regression gates in CI:** prompt and model changes run the eval suite; the gate is on the delta against the baseline, not an absolute score.
- **Multiple runs per case** (the model is non-deterministic), reported as pass@k or a mean with variance.
- **Cost and latency are tracked alongside quality.** The best model is the cheapest one that passes.

## Go you'll use

- **`encoding/json`** to load cases, plus a small `Expectation` interface with one implementation per check type.
- **A recording decorator:** wrap the tool registry or the provider to capture the trajectory (which tools, which inputs, how many steps, how many tokens).
- **`errgroup.WithContext` + `SetLimit`:** run cases concurrently without overloading the local model.
- **`text/tabwriter`** for the report, plus a JSON report for comparing runs.

## Tasks

### Task 1: Case format and checks
- **Claude scaffolds:** `internal/eval` with the `Case` struct (`q`, `must_call`, `must_not_call`, `answer_has`, `runs`) and check implementations with unit tests.
- **You implement:** the check logic. Remember that Bangla matching needs Unicode normalization (`golang.org/x/text/unicode/norm`).

### Task 2: Runner
- **You implement:** `cmd/eval`. It loads the cases and runs each one N times against the configured provider, using fixture tools so the data is stable. It records the trajectory and writes `evals/results/<timestamp>-<provider>.json`.

### Task 3: 30 real cases
- **You implement:** write 30 cases covering the DB, logs, docs, refusals ("delete all users"), the injection case, and Bangla versus English. This is domain work (guide §2, level D) and the most valuable part of the milestone.

### Task 4: Compare providers
- **You implement:** run the suite on `qwen3:14b`, `qwen3:30b-a3b` (both already on this Mac), and Claude Haiku. Build a table of pass rate, average steps, tokens, seconds and estimated cost. Record the decision in an ADR.

### Task 5: Regression gate
- **You implement:** `make eval-gate` fails if the pass rate drops more than X points below the committed baseline.

## Check yourself

1. Why run each case several times?
2. A case fails on the trajectory check but the answer is correct. Is that a failure?
3. What could make an LLM judge wrong, and how do you catch it?
4. Which metric would show a confusing tool description most quickly (guide §8)?
