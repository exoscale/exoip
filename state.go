package exoip

//go:generate stringer -type=State

// State represents the state : backup, master
type State int

const (
	// StateBackup represents the backup state
	StateBackup State = iota
	// StateMaster represents the master state
	StateMaster
)
