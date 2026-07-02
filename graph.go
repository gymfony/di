package di

// ServiceNode represents a single service in the code generation graph.
type ServiceNode struct {
	Type     string   // The name of the return type (e.g. "*myproject/cmd/di.UserService")
	CtorName string   // The name of the constructor function (e.g. "NewUserService")
	Deps     []string // List of dependency types (constructor arguments)
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
