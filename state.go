package exoip

//go:generate stringer -type=State

// State represents the state : backup, master
type State int

const (
	// StateUnknown represents the initial state
	StateUnknown State = iota
	// StateBackup represents the backup state
	StateBackup
	// StateMaster represents the master state
	StateMaster
)
