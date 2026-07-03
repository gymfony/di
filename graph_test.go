package di

import (
	"errors"
	"testing"
)

func TestDependencyGraph_Sort_Success(t *testing.T) {
	// Scenario: Successful dependency chain build
	// Config -> DB (requires Config) -> UserService (requires DB)
	g := &DependencyGraph{
		Nodes: map[string]*ServiceNode{
			"*test.UserService": {
				Type:     "*test.UserService",
				CtorName: "NewUserService",
				Deps:     []string{"*test.DB"},
			},
			"*test.DB": {
				Type:     "*test.DB",
				CtorName: "NewDB",
				Deps:     []string{"*test.Config"},
			},
			"*test.Config": {
				Type:     "*test.Config",
				CtorName: "NewConfig",
				Deps:     nil,
			},
		},
	}

	result, err := g.Sort()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 nodes in result, got %d", len(result))
	}

	// Check the correct topological order.
	// Each service must appear in the list AFTER its dependencies.
	positions := make(map[string]int)
	for i, node := range result {
		positions[node.Type] = i
	}

	if positions["*test.Config"] > positions["*test.DB"] {
		t.Error("expected *test.Config to be initialized before *test.DB")
	}
	if positions["*test.DB"] > positions["*test.UserService"] {
		t.Error("expected *test.DB to be initialized before *test.UserService")
	}
}

func TestDependencyGraph_Sort_Errors(t *testing.T) {
	tests := []struct {
		name     string
		nodes    map[string]*ServiceNode
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "Detect circular dependency",
			// Cycle: A requires B, B requires C, C requires A.
			nodes: map[string]*ServiceNode{
				"A": {Type: "A", Deps: []string{"B"}},
				"B": {Type: "B", Deps: []string{"C"}},
				"C": {Type: "C", Deps: []string{"A"}},
			},
			checkErr: func(t *testing.T, err error) {
				var cycleErr *CycleError
				if !errors.As(err, &cycleErr) {
					t.Fatalf("expected *CycleError, got %T (%v)", err, err)
				}

				// Check that the path contains a cycle.
				// The exact order depends on which map element the traversal begins with,
				// but the chain length must be equal to 4 (e.g., A -> B -> C -> A)
				if len(cycleErr.Path) != 4 {
					t.Errorf("expected cycle path length 4, got %d: %v", len(cycleErr.Path), cycleErr.Path)
				}

				// The first and last elements of the path must match, forming a loop
				first := cycleErr.Path[0]
				last := cycleErr.Path[len(cycleErr.Path)-1]
				if first != last {
					t.Errorf("expected cycle path to start and end with the same type, got %s and %s", first, last)
				}
			},
		},
		{
			name: "Detect missing provider",
			// A requires B, but B is not registered in the container
			nodes: map[string]*ServiceNode{
				"A": {Type: "A", Deps: []string{"B"}},
			},
			checkErr: func(t *testing.T, err error) {
				var missingErr *MissingProviderError
				if !errors.As(err, &missingErr) {
					t.Fatalf("expected *MissingProviderError, got %T (%v)", err, err)
				}
				if missingErr.TargetType != "B" {
					t.Errorf("expected missing TargetType 'B', got %q", missingErr.TargetType)
				}
				if missingErr.Dependant != "A" {
					t.Errorf("expected Dependant to be 'A', got %q", missingErr.Dependant)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &DependencyGraph{Nodes: tt.nodes}
			_, err := g.Sort()

			if err == nil {
				t.Fatal("expected an error, got nil")
			}

			tt.checkErr(t, err)
		})
	}
}

func TestDependencyGraph_Sort_Empty(t *testing.T) {
	// An empty graph should correctly return an empty slice without errors
	g := &DependencyGraph{Nodes: make(map[string]*ServiceNode)}
	result, err := g.Sort()

	if err != nil {
		t.Fatalf("expected no error for empty graph, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d nodes", len(result))
	}
}
