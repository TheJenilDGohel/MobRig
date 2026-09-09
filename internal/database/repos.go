package database

// Repositories is the aggregate entry point for all database repositories.
// Create one with NewRepositories(db) after opening the database.
type Repositories struct {
	Devices      *DeviceRepository
	Quirks       *QuirkRepository
	Sessions     *SessionRepository
	Flows        *FlowRepository
	Fingerprints *FingerprintRepository
	Failures     *FailurePatternRepository
}

// NewRepositories constructs all repositories from a single Database handle.
func NewRepositories(db *Database) *Repositories {
	return &Repositories{
		Devices:      NewDeviceRepository(db),
		Quirks:       NewQuirkRepository(db),
		Sessions:     NewSessionRepository(db),
		Flows:        NewFlowRepository(db),
		Fingerprints: NewFingerprintRepository(db),
		Failures:     NewFailurePatternRepository(db),
	}
}
