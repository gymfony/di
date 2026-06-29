package di

// Provider describes a single service constructor and its tags.
type Provider struct {
	Ctor interface{}  // Constructor function, e.g., NewUserService
	Tags []tagBinding // Tags bound to this provider
}

type tagBinding struct {
	name string
}

// Set combines several providers into a logical group (analogous to sets in wire).
type Set struct {
	Providers []Provider
}

// NewSet collects providers into a single set.
func NewSet(elements ...interface{}) Set {
	var set Set
	for _, element := range elements {
		switch el := element.(type) {
		case Provider:
			set.Providers = append(set.Providers, el)
		case Set:
			set.Providers = append(set.Providers, el.Providers...)
		default:
			// If a pure constructor function without tags is passed
			set.Providers = append(set.Providers, Provider{Ctor: el})
		}
	}
	return set
}

// TagKey associates a tag name with a strong type T.
type TagKey[T any] struct {
	name string
}

// NewTagKey creates a typed key for the tag.
func NewTagKey[T any](name string) TagKey[T] {
	return TagKey[T]{name: name}
}

// Name returns the debug tag name.
func (k TagKey[T]) Name() string {
	return k.name
}

// Tag binds the constructor to a specific TagKey[T].
// The compiler guarantees that the type T is the interface/type under which the tag is registered.
func Tag[T any](ctor interface{}, key TagKey[T]) Provider {
	return Provider{
		Ctor: ctor,
		Tags: []tagBinding{
			{name: key.name},
		},
	}
}
