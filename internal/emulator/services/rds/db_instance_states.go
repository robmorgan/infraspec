package rds

// DBInstanceState represents the status of an RDS DB instance
type DBInstanceState string

// DB instance states as defined by AWS RDS
const (
	DBInstanceStateCreating    DBInstanceState = "creating"
	DBInstanceStateAvailable   DBInstanceState = "available"
	DBInstanceStateModifying   DBInstanceState = "modifying"
	DBInstanceStateStarting    DBInstanceState = "starting"
	DBInstanceStateStopping    DBInstanceState = "stopping"
	DBInstanceStateStopped     DBInstanceState = "stopped"
	DBInstanceStateRebooting   DBInstanceState = "rebooting"
	DBInstanceStateDeleting    DBInstanceState = "deleting"
	DBInstanceStateBackingUp   DBInstanceState = "backing-up"
	DBInstanceStateFailed      DBInstanceState = "failed"
	DBInstanceStateStorageFull DBInstanceState = "storage-full"
)

// dbInstanceTransitions defines valid state transitions for DB instances
var dbInstanceTransitions = map[DBInstanceState][]DBInstanceState{
	DBInstanceStateCreating: {DBInstanceStateAvailable, DBInstanceStateFailed},
	DBInstanceStateAvailable: {
		DBInstanceStateModifying,
		DBInstanceStateStopping,
		DBInstanceStateRebooting,
		DBInstanceStateDeleting,
		DBInstanceStateBackingUp,
	},
	DBInstanceStateModifying:   {DBInstanceStateAvailable, DBInstanceStateFailed},
	DBInstanceStateStopping:    {DBInstanceStateStopped, DBInstanceStateFailed},
	DBInstanceStateStopped:     {DBInstanceStateStarting, DBInstanceStateDeleting},
	DBInstanceStateStarting:    {DBInstanceStateAvailable, DBInstanceStateFailed},
	DBInstanceStateRebooting:   {DBInstanceStateAvailable, DBInstanceStateFailed},
	DBInstanceStateBackingUp:   {DBInstanceStateAvailable},
	DBInstanceStateDeleting:    {}, // Terminal state - instance will be removed
	DBInstanceStateFailed:      {DBInstanceStateDeleting},
	DBInstanceStateStorageFull: {DBInstanceStateModifying, DBInstanceStateDeleting},
}

// IsValidDBInstanceTransition checks if a state transition is valid for DB instances
func IsValidDBInstanceTransition(from, to DBInstanceState) bool {
	allowed, ok := dbInstanceTransitions[from]
	if !ok {
		return false
	}
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	return false
}

// NewDBInstanceStateError creates an error for invalid DB instance state transitions
func NewDBInstanceStateError(identifier string, from, to DBInstanceState) *StateTransitionError {
	return &StateTransitionError{
		ResourceType: "db-instance",
		ResourceID:   identifier,
		FromState:    string(from),
		ToState:      string(to),
	}
}

// String returns the string representation of the state
func (s DBInstanceState) String() string {
	return string(s)
}

// IsTransitionalState returns true if the state is a transitional state
func (s DBInstanceState) IsTransitionalState() bool {
	switch s {
	case DBInstanceStateCreating,
		DBInstanceStateModifying,
		DBInstanceStateStarting,
		DBInstanceStateStopping,
		DBInstanceStateRebooting,
		DBInstanceStateDeleting,
		DBInstanceStateBackingUp:
		return true
	default:
		return false
	}
}

// IsStableState returns true if the state is a stable state
func (s DBInstanceState) IsStableState() bool {
	switch s {
	case DBInstanceStateAvailable,
		DBInstanceStateStopped,
		DBInstanceStateFailed,
		DBInstanceStateStorageFull:
		return true
	default:
		return false
	}
}
