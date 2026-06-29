package di

import "fmt"

// CycleError is raised when a cyclic dependency is detected.
type CycleError struct {
	Path []string // A chain of types that formed a cycle
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("di: circular dependency detected: %v", e.Path)
}

// MissingProviderError occurs when a constructor argument is missing from the container.
type MissingProviderError struct {
	TargetType string // The type that couldn't be resolved
	Dependant  string // Who requested it
}

func (e *MissingProviderError) Error() string {
	return fmt.Sprintf("di: missing provider for type %s (required by %s)", e.TargetType, e.Dependant)
}

// DuplicateBindingError occurs when two constructors are registered for the same type.
type DuplicateBindingError struct {
	Type string
}

func (e *DuplicateBindingError) Error() string {
	return fmt.Sprintf("di: duplicate provider binding for type %s", e.Type)
}
