package di

import (
	"github.com/lkmavi/saferefl"
)

type ContainerBuilder struct {
	// We use the string representation of the type from saferefl as the key,
	// or the metadata of the types provided by the library
	providers map[string]Provider
	tags      map[string][]string
}

func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{
		providers: make(map[string]Provider),
		tags:      make(map[string][]string),
	}
}

// AddProvider registers a single provider, retrieving information about the type T
func AddProvider[T any](cb *ContainerBuilder, p Provider) error {
	// Obtaining a type name T without heap allocations
	typeKey := typeKeyOf[T]()

	if _, exists := cb.providers[typeKey]; exists {
		return &DuplicateBindingError{Type: typeKey}
	}

	cb.providers[typeKey] = p

	// Registering tags
	for _, tag := range p.Tags {
		cb.tags[tag.name] = append(cb.tags[tag.name], typeKey)
	}

	return nil
}

// typeKeyOf returns a stable string key for type T using saferefl's TypeInfo.
func typeKeyOf[T any]() string {
	ti := saferefl.TypeOf[T]()
	return ti.String()
}

func (cb *ContainerBuilder) AddSet(set Set) error {
	// TODO: This is where the parsing logic for Stage 2 will come in.
	return nil
}
