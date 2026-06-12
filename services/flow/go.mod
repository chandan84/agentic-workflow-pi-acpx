module github.com/chandan84/agentic-workflow-pi-acpx/services/flow

go 1.22

require (
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/config v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/events v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/health v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/obs v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime v0.0.0
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/nats-io/nats.go v1.37.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/authn => ../../pkg/authn
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/config => ../../pkg/config
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/events => ../../pkg/events
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir => ../../pkg/flowir
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/health => ../../pkg/health
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/obs => ../../pkg/obs
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen => ../../pkg/protogen
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime => ../../pkg/runtime
)
