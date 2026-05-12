package pipeline

// SyncAccountsJob is the unit of work for one user.
type SyncAccountsJob struct {
	UserID string // TODO: match domain type (e.g. uuid.UUID)
	// TODO: add metadata (e.g. provider hints)
}
type RefreshPortfolioJob struct {
	UserID   string // TODO: match domain type (e.g. uuid.UUID)
	Simulate bool
	// TODO: add metadata (e.g. provider hints)
}
