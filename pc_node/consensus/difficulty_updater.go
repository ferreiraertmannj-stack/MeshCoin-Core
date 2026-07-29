package consensus

// DifficultyUpdater defines the interface for updating target and difficulty.
// The real implementation relies on the chain history (retargeting windows).
type DifficultyUpdater interface {
	UpdateDifficulty(currentHeight uint64) (newTarget string, newDifficulty string, changed bool)
	GetCurrentTarget() string
	GetCurrentDifficulty() string
}
