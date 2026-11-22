# M04: Real Tools

**Guide:** §6 step 4 · **Week:** 3 · **Depends on:** M03

## Concept

This milestone is where a toy agent becomes a useful one. It gets three tools, each of which teaches one principle:

| Tool | Principle |
| --- | --- |
| `query_db` | **Least privilege enforced by the system:** the database connection *cannot* write, so a "please don't write" in the prompt isn't relied on |
| `search_logs` | **The model never picks paths or URLs:** your config decides which file is read |
| `create_ticket` | **Human-in-the-loop:** risky actions need an explicit approval |

Once all three exist, the question "How many new users in the last 7 days, and what caused yesterday's 500 errors?" makes the agent chain time → DB → logs **by itself**. That is the difference between an agent and a workflow.

## How industry does it

- Agents get **their own credentials** with read-only grants, and they query a **replica or copy** of the data, never production primaries.
- The SQL keyword check is only there for convenience. The real control is `mode=ro` / `PRAGMA query_only`.
- Tool output is **bounded**: a row limit, a byte limit, and a result format (JSON) the model can parse.
- Results are **structured**, and errors are actionable ("column `signup_at` does not exist; columns are: ..."), so the model can correct itself.

## Go you'll use

- **`database/sql`** with `modernc.org/sqlite`, a pure-Go driver that needs no CGO, so cross-compiling to Windows stays easy (guide §3).
- **Dynamic row scanning:** `[]any` plus pointers, for columns that aren't known in advance. Convert `[]byte` values to `string`, because otherwise JSON encodes them as base64.
- **`bufio.Scanner`** with a larger buffer for long log lines, `regexp` or `strings.Contains` for matching, and a ring buffer for "the last 100 matches".
- **Constructor closures:** `QueryDBTool(db *sql.DB) tools.Tool`.

## Tasks

### Task 1: Fixtures
- **Claude scaffolds:** `testdata/app.sql` (users, courses, enrollments with known counts), `testdata/app.log` (with 500 errors and one injection line for M08), and a `make fixtures` target.
- **You implement:** run `make fixtures`, then look at the data with `sqlite3`.

### Task 2: `query_db`
- **Claude scaffolds:** `internal/tools/querydb.go` signatures and table tests, including `DELETE`, `SELECT ...; DROP`, `WITH ... DELETE`, a huge result, a timeout, and a blob column.
- **You implement:** open the DB read-only and enable `PRAGMA query_only=ON`, keep the SELECT check, cap results at 50 rows, and propagate the context.
- **Done when:** a write attempt fails *at the database layer*, even if you delete the keyword check. Prove it with a test.

### Task 3: `search_logs`
- **You implement:** input `{"pattern": string, "since_hours": int}`. Validate it (pattern not empty and ≤ 200 chars, `since_hours` between 1 and 168), read only the configured file, and return the last 100 matches.

### Task 4: `create_ticket` (risky)
- **You implement:** write the ticket to a local `tickets.jsonl` file, which stands in for your app's API. Set `Risky: true`. Test that it **never** runs when approval is denied.

### Task 5: Config
- **Claude scaffolds:** `internal/config` with a `Load()` that reads the DB path, the log path and the ticket path, and validates that the files exist.
- **You implement:** wire the new tools in `main.go`, then ask the agent the two-part question from the Concept section.

## Improvements over the guide's code

| Guide | Upgrade | Why |
| --- | --- | --- |
| `b, _ := json.Marshal(out)` | handle the error | `errcheck` would flag it; not every value can be marshaled |
| `[]byte` columns as base64 | converted to string | the model can read the result |
| `sql.Open` on every call? | one `*sql.DB` injected | connection pooling |
| table names hard-coded in the description | generated from `sqlite_master` at startup | the description stays correct when the schema changes |

## Check yourself

1. If you deleted the `SELECT` prefix check, would `query_db` still be safe? Why?
2. Why must the model never choose the log file path?
3. Why is a ticket "risky" when a DB read is not?
4. Which tool should *fail* and which should return an empty result when nothing matches?
