package flowir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

var compiledSchema *jsonschema.Schema

func schema() (*jsonschema.Schema, error) {
	if compiledSchema != nil {
		return compiledSchema, nil
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("flow-ir.v1.json", bytes.NewReader(SchemaV1)); err != nil {
		return nil, err
	}
	s, err := c.Compile("flow-ir.v1.json")
	if err != nil {
		return nil, err
	}
	compiledSchema = s
	return s, nil
}

// ValidateJSON validates a raw IR JSON document against the v1 schema.
func ValidateJSON(raw []byte) error {
	s, err := schema()
	if err != nil {
		return err
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	if err := s.Validate(v); err != nil {
		return err
	}
	return nil
}

// Parse validates and decodes a raw IR JSON document into a typed IR.
func Parse(raw []byte) (*IR, error) {
	if err := ValidateJSON(raw); err != nil {
		return nil, err
	}
	var ir IR
	if err := json.Unmarshal(raw, &ir); err != nil {
		return nil, err
	}
	return &ir, nil
}

// DAGCheck runs the structural validators beyond JSON-schema:
//
//   - all node ids are unique
//   - start references an existing node
//   - every node reachable from start
//   - every edge endpoint exists
//   - decision nodes have at least one branch; non-decision nodes have none
//   - fork/join are balanced (one join per fork in the path)
//   - end nodes are reachable
func DAGCheck(ir *IR) []error {
	var errs []error
	ids := map[string]int{}
	for i, n := range ir.Nodes {
		if ids[n.ID] > 0 {
			errs = append(errs, fmt.Errorf("duplicate node id %q (positions %d and %d)", n.ID, ids[n.ID]-1, i))
		}
		ids[n.ID] = i + 1
	}
	if _, ok := ids[ir.Start]; !ok {
		errs = append(errs, fmt.Errorf("start %q references unknown node", ir.Start))
	}
	for _, e := range ir.Edges {
		if _, ok := ids[e.From]; !ok {
			errs = append(errs, fmt.Errorf("edge from unknown node %q", e.From))
		}
		if _, ok := ids[e.To]; !ok {
			errs = append(errs, fmt.Errorf("edge to unknown node %q", e.To))
		}
	}

	adj := map[string][]string{}
	for _, e := range ir.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	for _, n := range ir.Nodes {
		if n.Kind == NodeDecision {
			if len(n.Branches) == 0 {
				errs = append(errs, fmt.Errorf("decision node %q has no branches", n.ID))
			}
			for _, b := range n.Branches {
				if _, ok := ids[b.To]; !ok {
					errs = append(errs, fmt.Errorf("decision %q branches to unknown node %q", n.ID, b.To))
				}
				adj[n.ID] = append(adj[n.ID], b.To)
			}
		} else if len(n.Branches) > 0 {
			errs = append(errs, fmt.Errorf("non-decision node %q has branches", n.ID))
		}
	}

	reach := map[string]bool{}
	var visit func(string)
	visit = func(id string) {
		if reach[id] {
			return
		}
		reach[id] = true
		for _, next := range adj[id] {
			visit(next)
		}
	}
	if _, ok := ids[ir.Start]; ok {
		visit(ir.Start)
	}
	for _, n := range ir.Nodes {
		if !reach[n.ID] {
			errs = append(errs, fmt.Errorf("node %q is unreachable from start", n.ID))
		}
	}

	forks := 0
	joins := 0
	for _, n := range ir.Nodes {
		if n.Kind == NodeFork {
			forks++
		}
		if n.Kind == NodeJoin {
			joins++
		}
	}
	if forks != joins {
		errs = append(errs, fmt.Errorf("unbalanced fork/join: %d forks, %d joins", forks, joins))
	}

	endSeen := false
	for _, n := range ir.Nodes {
		if n.Kind == NodeEnd && reach[n.ID] {
			endSeen = true
		}
	}
	if !endSeen && len(ir.Ends) == 0 {
		errs = append(errs, errors.New("no reachable end node and no Ends declared"))
	}
	return errs
}
