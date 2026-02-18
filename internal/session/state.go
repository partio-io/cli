package session

// State represents the lifecycle state of a session.
type State string

const (
	StateIdle   State = "idle"
	StateActive State = "active"
	StateEnded  State = "ended"
)
