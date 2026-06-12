module github.com/chandan84/agentic-workflow-pi-acpx/tests/e2e

go 1.22

require (
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/events v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime v0.0.0
)

require (
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/nats-io/nats.go v1.37.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
)

replace (
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/events => ../../pkg/events
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir => ../../pkg/flowir
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime => ../../pkg/runtime
)
