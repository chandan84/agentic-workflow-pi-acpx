// Package flowir contains the canonical Flow IR — versioned JSON understood
// by every service. The single source of truth is schema/flow-ir.v1.json; the
// Go structs below mirror it and are kept in sync by the validators.
package flowir

import _ "embed"

//go:embed schema/flow-ir.v1.json
var SchemaV1 []byte

// IR is the top-level Flow IR document.
type IR struct {
	SchemaVersion int          `json:"schemaVersion"`
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Description   string       `json:"description,omitempty"`
	Permissions   *Permissions `json:"permissions,omitempty"`
	Start         string       `json:"start"`
	Ends          []string     `json:"ends,omitempty"`
	Nodes         []Node       `json:"nodes"`
	Edges         []Edge       `json:"edges"`
	WorkItems     []WorkItem   `json:"workItems,omitempty"`
}

// Permissions caps which roles may execute or approve.
type Permissions struct {
	Roles []string `json:"roles,omitempty"`
}

// Node is one step in the IR. Kind-specific fields are loosely typed because
// authors can add custom metadata; the schema validates the known shape.
type Node struct {
	ID             string                 `json:"id"`
	Kind           string                 `json:"kind"`
	Title          string                 `json:"title,omitempty"`
	AgentID        string                 `json:"agentId,omitempty"`
	Skill          string                 `json:"skill,omitempty"`
	Prompt         string                 `json:"prompt,omitempty"`
	Fn             string                 `json:"fn,omitempty"`
	Args           map[string]any         `json:"args,omitempty"`
	Branches       []DecisionBranch       `json:"branches,omitempty"`
	Approvers      []string               `json:"approvers,omitempty"`
	TimeoutSeconds int                    `json:"timeoutSeconds,omitempty"`
	OnTimeout      string                 `json:"onTimeout,omitempty"`
	Extra          map[string]any         `json:"-"`
}

// DecisionBranch maps a condition expression to a target node id.
type DecisionBranch struct {
	When string `json:"when"`
	To   string `json:"to"`
}

// Edge is a directed connection between nodes; condition is optional and used
// only when neither endpoint is a decision (decision branches carry their own).
type Edge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition,omitempty"`
}

// WorkItem is the schema attached to a node that emits human work.
type WorkItem struct {
	NodeID string         `json:"nodeId"`
	Kind   string         `json:"kind"`
	Schema map[string]any `json:"schema,omitempty"`
}

// NodeKind constants.
const (
	NodeACP        = "acp"
	NodeAction     = "action"
	NodeCompute    = "compute"
	NodeDecision   = "decision"
	NodeCheckpoint = "checkpoint"
	NodeFork       = "fork"
	NodeJoin       = "join"
	NodeEnd        = "end"
)
