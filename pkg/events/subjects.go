// Package events is the catalog of NATS JetStream subjects and a thin
// publish/subscribe wrapper. Every service references these constants —
// docs/spec/events.md is generated from this file.
package events

// Subjects produced by each bounded context.
const (
	// agents.*
	SubjectAgentCreated   = "agents.agent.created"
	SubjectGroupCreated   = "agents.group.created"
	SubjectResourceLinked = "agents.resource.linked"

	// flows.*
	SubjectFlowCreated         = "flows.flow.created"
	SubjectFlowVersionCreated  = "flows.flow.version.created"
	SubjectFlowGenerationLog   = "flows.flow.generation.log"
	SubjectFlowGenerationDone  = "flows.flow.generation.done"
	SubjectFlowGenerationError = "flows.flow.generation.error"

	// runs.*
	SubjectRunStarted          = "runs.run.started"
	SubjectRunUpdated          = "runs.run.updated"
	SubjectRunCompleted        = "runs.run.completed"
	SubjectRunFailed           = "runs.run.failed"
	SubjectRunCheckpointWait   = "runs.checkpoint.awaiting"
	SubjectRunCheckpointDone   = "runs.checkpoint.resolved"
	SubjectRunSegmentStart     = "runs.segment.started"
	SubjectRunSegmentEnd       = "runs.segment.finished"
	SubjectRunWorkItemEmitted  = "runs.workitem.emitted"

	// audit.*
	SubjectAuditEvent = "audit.event"
)

// AllSubjects returns the catalog in a stable order for docs and tests.
func AllSubjects() []string {
	return []string{
		SubjectAgentCreated, SubjectGroupCreated, SubjectResourceLinked,
		SubjectFlowCreated, SubjectFlowVersionCreated,
		SubjectFlowGenerationLog, SubjectFlowGenerationDone, SubjectFlowGenerationError,
		SubjectRunStarted, SubjectRunUpdated, SubjectRunCompleted, SubjectRunFailed,
		SubjectRunCheckpointWait, SubjectRunCheckpointDone,
		SubjectRunSegmentStart, SubjectRunSegmentEnd, SubjectRunWorkItemEmitted,
		SubjectAuditEvent,
	}
}
