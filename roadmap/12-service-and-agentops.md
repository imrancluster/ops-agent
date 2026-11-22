# M12: HTTP Service and AgentOps

**Guide:** §7 step 10 (deployment), §8 (AgentOps) · **Week:** 11 · **Depends on:** M06, M08, M10

## Concept

A REPL on your laptop is a prototype. A **service** is always running, serves several users at once, survives restarts, and can be stopped instantly when it misbehaves. The guide's §8 frames it well: **treat an agent like a new employee.** It needs an ID card (`agent.yaml`), limited permissions, a budget, a supervisor (approval plus an audit log), and regular reviews (evals).

The five management principles (guide §8):

1. Every change gets a new version and an eval run.
2. A **kill switch** that stops the agent within one second.
3. A separate identity and permissions: the agent's own DB user and API key, never your admin credentials.
4. Append-only audit logs.
5. A feedback loop: 👍/👎, with the 👎 cases becoming evals.

## How industry does it

- **Stateless HTTP handlers, state in a store:** each request loads the session, runs the agent, and saves it. With a per-session lock, two concurrent requests can't corrupt one history.
- **AuthN on every endpoint.** Without it, anyone can run your agent at your expense.
- **Graceful shutdown:** stop accepting requests, finish in-flight ones, flush traces.
- **Health and readiness endpoints, a container image, config from the environment.** These are the standard twelve-factor patterns.
- **Agent registries and control planes** (the guide mentions Microsoft Agent 365): `agent.yaml` is your small version of one.
- **Approvals become asynchronous:** a risky tool call creates a *pending approval* that a human resolves later, so the service doesn't block on stdin.

## Go you'll use

- **`net/http` with method patterns** (`mux.HandleFunc("POST /chat", ...)`, available since Go 1.22), `http.Server` with timeouts (`ReadHeaderTimeout` etc.; the defaults are unsafe).
- **`signal.NotifyContext` + `srv.Shutdown(ctx)`** for graceful shutdown.
- **Per-key locking:** `map[string]*sync.Mutex` guarded by a mutex, or a striped lock.
- **Middleware:** `func(http.Handler) http.Handler` for auth, request ID, logging and recovery.
- **Constant-time token comparison:** `crypto/subtle.ConstantTimeCompare`.
- **`gopkg.in/yaml.v3`** (or JSON) for `agent.yaml`.

## Tasks

### Task 1: `cmd/server`
- **Claude scaffolds:** the server with timeouts, middleware chain, `/healthz` and `/readyz`, and `httptest` handler tests.
- **You implement:**
  - `POST /chat {session, text}`, which loads the session, runs the agent under a per-session lock with a 2-minute context, saves the result, and returns the reply.
  - **Generic error messages to the client** and full details in the logs. The guide's `http.Error(w, err.Error())` leaks internals.

### Task 2: Auth
- **You implement:** bearer-token middleware (token from env) with a constant-time compare, plus tests for a missing, wrong and valid token.

### Task 3: Async approvals
- **You implement:** a risky tool returns "pending approval #id" to the model. Add `GET /approvals` and `POST /approvals/{id}` (approve or deny), and write every decision to the audit log.

### Task 4: `agent.yaml` and kill switch
- **You implement:** load and validate `agent.yaml` (guide §8). It decides the tools, budget and prompt version. When `kill_switch: true` is set, or `POST /admin/kill` is called, new requests get 503 immediately.

### Task 5: Deploy
- **Claude scaffolds:** a multi-stage `Dockerfile` (distroless, non-root, CGO off), and the cross-compile target `GOOS=windows GOARCH=amd64`.
- **You implement:** run it in Docker locally. Optionally run it on Win-1 as a service, then point the M09 eval runner at it over HTTP.

### Task 6: Feedback
- **You implement:** `POST /feedback {session, turn, thumbs}`, and a script that turns the week's 👎 cases into draft eval cases.

## Check yourself

1. What corrupts the history if two requests for the same session run at once without a lock?
2. Why is returning `err.Error()` to the client a security problem?
3. Why does a stopped agent return 503 rather than a polite refusal from the model?
4. Which numbers would you check every morning (guide §8 table)?
