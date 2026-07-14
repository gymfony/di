package di

import "strings"

// ServiceNode represents a single service in the code generation graph.
type ServiceNode struct {
	Type     string   // The name of the return type (e.g. "*myproject/cmd/di.UserService")
	CtorName string   // The name of the constructor function (e.g. "NewUserService")
	Deps     []string // List of dependency types (constructor arguments)
	Tags     []string // The names of the tags this service is bound to (e.g. ["api_routes"])
}

// DependencyGraph accumulates all nodes for analysis
type DependencyGraph struct {
	Nodes map[string]*ServiceNode
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[string]*ServiceNode),
	}
}

const (
	newNode byte = iota
	inProgressNode
	finishedNode
)

func (g *DependencyGraph) Sort() ([]*ServiceNode, error) {
	var result []*ServiceNode
	visited := make(map[string]byte, len(g.Nodes))
	result = make([]*ServiceNode, 0, len(g.Nodes))

	var stack []string
	var visit func(string) error

	visit = func(nodeType string) error {
		state := visited[nodeType]
		if state == inProgressNode {
			cyclePath := append([]string{}, stack...)
			cyclePath = append(cyclePath, nodeType)
			return &CycleError{Path: cyclePath}
		}
		if state == finishedNode {
			return nil
		}

		visited[nodeType] = inProgressNode
		stack = append(stack, nodeType)

		node, exists := g.Nodes[nodeType]
		if !exists {
			var dependant string
			if len(stack) > 1 {
				dependant = stack[len(stack)-2]
			}
			return &MissingProviderError{
				TargetType: nodeType,
				Dependant:  dependant,
			}
		}

		if err := g.visitDeps(node, visit); err != nil {
			return err
		}

		stack = stack[:len(stack)-1]
		visited[nodeType] = finishedNode
		result = append(result, node)

		return nil
	}

	for nodeType := range g.Nodes {
		if visited[nodeType] == newNode {
			if err := visit(nodeType); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// visitDeps iterates through all node dependencies and runs visit on them.
func (g *DependencyGraph) visitDeps(node *ServiceNode, visit func(string) error) error {
	for _, depType := range node.Deps {
		if strings.HasPrefix(depType, "[]") {
			if err := g.visitTagDeps(strings.TrimPrefix(depType, "[]"), visit); err != nil {
				return err
			}
			continue
		}

		if err := visit(depType); err != nil {
			return err
		}
	}
	return nil
}

// visitTagDeps finds all nodes marked with a given tag and calls visit on them.
func (g *DependencyGraph) visitTagDeps(cleanTagType string, visit func(string) error) error {
	for _, potentialNode := range g.Nodes {
		for _, t := range potentialNode.Tags {
			if t == cleanTagType {
				if err := visit(potentialNode.Type); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
