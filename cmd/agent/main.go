// Command agent is the Ops Agent entry point: it loads configuration, wires
// the LLM provider, tools and memory together, and starts the REPL.
//
// This is the only place that reads environment variables.
package main

import "fmt"

func main() {
	// TODO(M03): replace with config loading, dependency wiring and the REPL.
	fmt.Println("ops-agent: not implemented yet (see roadmap/03-agent-loop.md)")
}
