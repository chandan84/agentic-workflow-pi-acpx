module github.com/chandan84/agentic-workflow-pi-acpx/tests/e2e

go 1.22

require (
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/events v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/services/agents v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/services/flow v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/services/gateway v0.0.0
	github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator v0.0.0
	google.golang.org/grpc v1.67.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/nats-io/nats.go v1.37.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nexus-rpc/sdk-go v0.0.9 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.temporal.io/api v1.36.0 // indirect
	go.temporal.io/sdk v1.28.1 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/exp v0.0.0-20231127185646-65229373498e // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Force genproto to a split-modules-era version so imports of
// google.golang.org/genproto/googleapis/{api,rpc} are unambiguous in the
// transitive closure (temporal sdk pulls them in).
replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20240814211410-ddb44dafa142

replace (
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/authn => ../../pkg/authn
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/config => ../../pkg/config
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/events => ../../pkg/events
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir => ../../pkg/flowir
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/health => ../../pkg/health
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/obs => ../../pkg/obs
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/pgxconn => ../../pkg/pgxconn
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen => ../../pkg/protogen
	github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime => ../../pkg/runtime
	github.com/chandan84/agentic-workflow-pi-acpx/services/agents => ../../services/agents
	github.com/chandan84/agentic-workflow-pi-acpx/services/flow => ../../services/flow
	github.com/chandan84/agentic-workflow-pi-acpx/services/gateway => ../../services/gateway
	github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator => ../../services/orchestrator
)
