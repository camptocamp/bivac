package orchestrators

// Orchestrator implements a container Orchestrator interface
type Orchestrator interface {
	GetName() string
}
