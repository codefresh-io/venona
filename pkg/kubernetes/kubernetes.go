package kubernetes

type (
	// Kubernetes API client
	Kubernetes interface {
		CreateResource() error
		DeleteResource() error
	}

	kubernetes struct{}
)

// New build Kubernetes API
func New() Kubernetes {
	return nil
}

func (k kubernetes) CreateResource() error {
	return nil
}

func (k kubernetes) DeleteResource() error {
	return nil
}
