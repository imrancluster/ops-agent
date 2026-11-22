// Package llm is the agent's only door to language models.
//
// It defines a provider-neutral message format (content blocks: text,
// tool_use, tool_result) and the Provider interface. Each adapter (Ollama,
// Claude, OpenAI-compatible) translates that format to and from one vendor's
// wire format, so no other package ever sees vendor JSON.
//
// This package must not import any other internal package.
package llm
