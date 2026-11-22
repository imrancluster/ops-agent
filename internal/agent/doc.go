// Package agent runs the agent loop: ask the model what to do, execute the
// tools it requests (subject to guardrails), feed the results back, and repeat
// until the model answers or a limit (steps, budget, time) is reached.
//
// The model only decides; this package acts, and enforces the rules.
package agent
