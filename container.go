package di

// Container holds service bindings and resolves dependencies.
type Container struct {
	bindings map[string]any
}

// New returns a new empty Container.
func New() *Container {
	return &Container{
		bindings: make(map[string]any),
	}
}
