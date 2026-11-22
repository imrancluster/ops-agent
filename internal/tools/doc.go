// Package tools holds the tools the agent can call. Each tool pairs the
// definition the model reads (name, description, JSON Schema) with the Go
// function that actually runs it, so the two can never drift apart.
//
// Tools get least privilege, enforced in code: read-only database handles,
// file paths from config, never from the model.
package tools
