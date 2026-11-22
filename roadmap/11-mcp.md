# M11: MCP (Model Context Protocol)

**Guide:** §7 step 7 · **Week:** 10 · **Depends on:** M04

## Concept: the "USB port" for tools

So far, your tools only work inside *your* agent. **MCP** is an open standard for connecting tools and data to *any* agent. Once your `query_db` is an MCP server, Claude Desktop, Claude Code, IDEs and other agents can all use it, and your agent can in turn use the thousands of MCP servers others have built.

- **MCP server:** offers tools (and resources and prompts) through a standard JSON-RPC protocol over stdio or HTTP.
- **MCP client:** lists the server's tools, converts them into `llm.Tool`, and forwards calls to the server.
- **MCP vs A2A:** MCP connects an *agent to tools*. A2A (M13) connects an *agent to agents*.

## How industry does it

- Tools are increasingly shipped **as MCP servers**, decoupled from any one agent. Teams publish internal MCP servers for databases, ticketing and docs.
- **Trust boundaries matter more:** a third-party MCP server's tool *descriptions* can themselves contain prompt injections. Allowlist the servers you connect to, and review their tool descriptions.
- Use the official SDK (`github.com/modelcontextprotocol/go-sdk`, maintained together with Google). Its API is still evolving, so **always copy code from its README and `examples/`**, not from memory.

## Go you'll use

- **Adapter pattern once more:** MCP tool → `tools.Tool`. Your existing `Run` functions are reused almost unchanged, which is the payoff of designing tools as closures in M02 and M04.
- **`os/exec`** for launching a stdio MCP server as a subprocess, with proper cleanup on `ctx` cancellation.
- **Generics** in the SDK's typed tool handlers.

## Tasks

### Task 1: Hello MCP
- **You implement:** run the SDK README's "Getting started" greet server exactly as written. Connect to it from Claude Code (`claude mcp add`) and call it.

### Task 2: Ops tools as an MCP server
- **Claude scaffolds:** `cmd/mcp-server` wiring, after checking the current SDK API together with you.
- **You implement:** register `query_db`, `search_logs` and `search_docs` by reusing the existing tool `Run` functions. The risky `create_ticket` stays **out** of the server. Explain why in an ADR.

### Task 3: MCP client in the agent
- **You implement:** `internal/tools/mcpclient.go`, which connects to the configured servers (an allowlist in config), lists their tools, wraps them as `tools.Tool` with a name prefix (`mcp_<server>_<tool>`) to avoid name collisions, and forwards calls to the server.
- **Done when:** your agent uses its own tools *through* MCP and the M09 eval pass rate doesn't change.

## Check yourself

1. What problem does MCP solve that a Go package of tools doesn't?
2. Why might an MCP server's tool description be dangerous?
3. Why keep `create_ticket` out of the shared server?
4. Where does human approval happen when a tool comes from MCP?
