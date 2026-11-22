# M10: Observability

**Guide:** §7 step 9 · **Week:** 9 · **Depends on:** M03, M08

## Concept

When an agent gives a wrong answer, "what did it do?" is a multi-step question: which model, which prompt version, which tools in what order, with what inputs, how long each step took, and what it all cost. **Observability** is the ability to answer that afterwards, from data alone.

There are three signals:

- **Logs:** structured events (`slog` JSON) with a **session ID** on every line.
- **Metrics:** counters and histograms: cost per day, steps per task, tool failure rate, latency.
- **Traces:** one session drawn as a tree: `session → llm call → tool → llm call`, with timings. View them in Jaeger.

## How industry does it

- **OpenTelemetry (OTel)** is the vendor-neutral standard. You instrument once and export anywhere.
- The **GenAI semantic conventions** give standard attribute names for model, token counts and operation (`gen_ai.*`), so dashboards work across tools.
- **Cost is a first-class metric.** Tokens are multiplied by a price table, aggregated per session, per day and per prompt version.
- **Privacy:** prompts and completions are often *not* stored by default, or are redacted (M08), because they contain user data.
- **Daily review** (guide §8): cost, error rate, sessions stuck at MaxSteps, pending approvals.

## Go you'll use

- **`log/slog`:** `slog.New(slog.NewJSONHandler(...))`, `logger.With("session", id)`, and passing a logger through dependencies, not a global.
- **`context.Context` values** only for request-scoped data such as the trace/span context. OTel does this for you.
- **The `go.opentelemetry.io/otel` SDK:** tracer provider, spans, attributes, and the OTLP exporter.
- **Middleware/decorator:** a `TracingProvider` wraps `llm.Provider`, so the agent code doesn't change.

## Tasks

### Task 1: Structured logging
- **Claude scaffolds:** the logger setup in `main` (text in dev, JSON in prod, chosen from config), with the logger injected into the Agent.
- **You implement:** log events for each step and each tool call with session, step, model, tokens, ms, tool and error. Apply the redaction from M08.

### Task 2: Cost accounting
- **You implement:** a price table in config (per million input/output tokens per model, with local models at 0), `Usage` → USD, and a total per session. The M08 budget uses it.

### Task 3: Tracing
- **Claude scaffolds:** an OTel init function that exports to OTLP, plus `docker compose` for Jaeger.
- **You implement:** spans for the session, each LLM call (as a decorator) and each tool call, with GenAI attributes. Run a question, then open the trace in Jaeger.

### Task 4: Daily report
- **You implement:** `cmd/report` reads the JSON logs (or the session DB) and prints yesterday's cost, error rate, average steps, and the slowest and most expensive questions.

## Check yourself

1. The average step count suddenly rises. What does that usually mean (guide §8)?
2. Why add tracing as a decorator rather than inside the agent loop?
3. Why might you choose *not* to log full prompts?
4. What is the difference between a log, a metric and a trace? Give one agent example of each.
