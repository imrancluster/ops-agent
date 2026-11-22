# M13: Multi-Agent Systems and the Future

**Guide:** §7 step 10 (multi-agent), §9 (future agents) · **Week:** 12 · **Depends on:** M09, M12

## Concept

Several agents working together can handle tasks that are too broad for one context window or one set of tools. The guide names **three proven patterns**:

| Pattern | How | When |
| --- | --- | --- |
| **Router** | A small, fast model classifies the request and hands it to the right agent | Different kinds of questions (support, billing, tech) |
| **Orchestrator–Worker** | A manager agent splits the work and calls worker agents *as tools* | Large research or report tasks |
| **Evaluator–Optimizer** | One agent writes, another critiques, and they loop | Improving code or writing quality |

In Go, **"agent as a tool"** is short: a `tools.Tool` whose `Run` calls another `Agent.Run`, with **its own** small tool list and **its own** history. The main agent's context stays clean.

**Warning (guide):** multi-agent setups increase both cost and complexity. Don't go there until one agent has demonstrably failed, *as shown by your evals*.

## How industry does it

- **Start with a single agent**, and add agents only when evals show a gap.
- **Sub-agents get the least context and least tools** they need. They return *summaries*, not raw dumps.
- **Fan-out runs in parallel** with bounded concurrency, and every worker shares the session's budget.
- **Across organizations:** the **A2A protocol** (agent-to-agent, v1.0, supported by more than 150 organizations, now under the Agentic AI Foundation alongside MCP). An agent publishes an "agent card" and others call it over HTTP.
- **What's coming next (guide §9, which is the author's forecast):** long-running agents built on *durable execution*, agent identity and scoped permissions, computer-use agents in sandboxes, and small on-device models. Each of these extends a milestone you have already built.

## Go you'll use

- **`errgroup.WithContext` + `SetLimit`:** parallel workers, where one failure cancels the rest.
- **Shared budget:** atomic counters or a budget type with a mutex, passed to every sub-agent.
- **Recursion guards:** a depth limit in `ctx` so agents can't call each other forever.
- **HTTP clients for remote agents** (Win-1 data agent ↔ Mac-1 orchestrator).

## Tasks

### Task 1: Router
- **You implement:** `qwen3:4b` or `llama3.2` classifies each request into `data | docs | other`, and the request is routed to a specialized agent with a smaller tool set. Run evals before and after, and compare.

### Task 2: Agent-as-tool
- **Claude scaffolds:** `AgentTool(name, desc string, factory func() *agent.Agent) tools.Tool` with a depth guard.
- **You implement:** an orchestrator with two workers: a "data analyst" (`query_db`) and a "log investigator" (`search_logs`). Ask a question that needs both, then trace it in Jaeger (M10).

### Task 3: Parallel workers
- **You implement:** when the orchestrator calls several workers in one turn, run them concurrently with `errgroup`, deducting from a shared budget. Test with `go test -race`.

### Task 4: Remote agent and A2A
- **You implement:** run the data agent as its own HTTP service (M12), and have the orchestrator call it as a tool over HTTP. Read the A2A spec, and write in `notes/m13.md` what it would take to expose your agent as an A2A endpoint (agent card, task lifecycle).

### Task 5: Teach it (guide §10, the final task)
- **You implement:** write a post or record a video explaining the agent loop and one hard lesson from the project. Teaching is where gaps in your understanding become visible.

## Check yourself

1. When is a multi-agent system the *wrong* choice?
2. Why do sub-agents return summaries rather than their full history?
3. How do you stop two agents from calling each other forever?
4. MCP or A2A: which one would you use to let a partner company's agent ask yours a question?
5. Which of the guide's five future-proof skills (§9) did each milestone exercise?
