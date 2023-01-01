package updater

// Store - the main interface that describes how to
// interact with the store or repository layer
type Store interface {
	UpdaterStore
}

// Service - is the struct on which all our logic
// will be built on top of
type Service struct {
	Store Store
}

// NewService - returns a pointer to a new service
func NewService(store Store) *Service {
	return &Service{
		Store: store,
	}
}
