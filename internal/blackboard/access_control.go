package blackboard

// AccessLevel represents read/write access to a blackboard level
type AccessLevel int

const (
	AccessNone AccessLevel = iota
	AccessRead
	AccessReadWrite
)

// Blackboard levels
const (
	Level0Meta       = 0  // Project meta info
	Level1Facts      = 1  // Certain facts
	Level2Conjecture = 2  // Conjectures
	Level3Progress   = 3  // Work progress
	Level4Discovery  = 4  // New discoveries
	Level4_5Aux      = 45 // Auxiliary knowledge (error logs, diagnostics)
	Level5Resolution = 5  // Meeting resolutions
)

// RoleAccessMatrix defines read/write permissions per role per level
// Maps role -> level -> access
var RoleAccessMatrix = map[string]map[int]AccessLevel{
	"ceo_secretary": {
		Level0Meta:       AccessReadWrite,
		Level1Facts:      AccessReadWrite,
		Level2Conjecture: AccessReadWrite,
		Level3Progress:   AccessRead,
		Level4Discovery:  AccessRead,
		Level4_5Aux:      AccessRead,
		Level5Resolution: AccessRead,
	},
	"hr": {
		Level0Meta:       AccessRead,
		Level1Facts:      AccessRead,
		Level2Conjecture: AccessRead,
		Level3Progress:   AccessRead,
		Level4Discovery:  AccessRead,
		Level4_5Aux:      AccessRead,
		Level5Resolution: AccessRead,
	},
	"pm": {
		Level0Meta:       AccessReadWrite,
		Level1Facts:      AccessReadWrite,
		Level2Conjecture: AccessReadWrite,
		Level3Progress:   AccessReadWrite,
		Level4Discovery:  AccessReadWrite,
		Level4_5Aux:      AccessReadWrite,
		Level5Resolution: AccessRead,
	},
	"developer": {
		Level0Meta:       AccessRead,
		Level1Facts:      AccessRead,
		Level2Conjecture: AccessReadWrite,
		Level3Progress:   AccessReadWrite,
		Level4Discovery:  AccessReadWrite,
		Level4_5Aux:      AccessReadWrite,
		Level5Resolution: AccessRead,
	},
	"tester": {
		Level0Meta:       AccessRead,
		Level1Facts:      AccessRead,
		Level2Conjecture: AccessReadWrite,
		Level3Progress:   AccessReadWrite,
		Level4Discovery:  AccessReadWrite,
		Level4_5Aux:      AccessReadWrite,
		Level5Resolution: AccessRead,
	},
	"designer": {
		Level0Meta:       AccessRead,
		Level1Facts:      AccessReadWrite,
		Level2Conjecture: AccessReadWrite,
		Level3Progress:   AccessReadWrite,
		Level4Discovery:  AccessReadWrite,
		Level4_5Aux:      AccessRead,
		Level5Resolution: AccessRead,
	},
	"reviewer": {
		Level0Meta:       AccessRead,
		Level1Facts:      AccessReadWrite,
		Level2Conjecture: AccessRead,
		Level3Progress:   AccessRead,
		Level4Discovery:  AccessReadWrite,
		Level4_5Aux:      AccessReadWrite,
		Level5Resolution: AccessRead,
	},
	"ops": {
		Level0Meta:       AccessRead,
		Level1Facts:      AccessRead,
		Level2Conjecture: AccessReadWrite,
		Level3Progress:   AccessReadWrite,
		Level4Discovery:  AccessReadWrite,
		Level4_5Aux:      AccessReadWrite,
		Level5Resolution: AccessRead,
	},
	"doc": {
		Level0Meta:       AccessRead,
		Level1Facts:      AccessRead,
		Level2Conjecture: AccessRead,
		Level3Progress:   AccessReadWrite,
		Level4Discovery:  AccessRead,
		Level4_5Aux:      AccessRead,
		Level5Resolution: AccessRead,
	},
}

// CanRead checks if a role can read a given blackboard level
func CanRead(role string, level int) bool {
	levels, ok := RoleAccessMatrix[role]
	if !ok {
		return false
	}
	access, ok := levels[level]
	if !ok {
		return false
	}
	return access >= AccessRead
}

// CanWrite checks if a role can write to a given blackboard level
func CanWrite(role string, level int) bool {
	levels, ok := RoleAccessMatrix[role]
	if !ok {
		return false
	}
	access, ok := levels[level]
	if !ok {
		return false
	}
	return access >= AccessReadWrite
}

// CategoryToLevel maps blackboard categories to levels
func CategoryToLevel(category string) int {
	switch category {
	case CategoryGoal:
		return Level0Meta
	case CategoryFact:
		return Level1Facts
	case CategoryDiscovery:
		return Level4Discovery
	case CategoryDecision:
		return Level1Facts
	case CategoryProgress:
		return Level3Progress
	case CategoryResolution:
		return Level5Resolution
	case CategoryAuxiliary:
		return Level4_5Aux
	case CategoryAcceptanceCrit:
		return Level1Facts
	case CategoryVerification:
		return Level3Progress
	default:
		return Level2Conjecture
	}
}
