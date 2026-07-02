package digen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type mockLogger struct{ t *testing.T }

func (m *mockLogger) Debug(msg string, args ...any) { m.t.Logf("[DEBUG] "+msg, args...) }
func (m *mockLogger) Info(msg string, args ...any)  { m.t.Logf("[INFO] "+msg, args...) }
func (m *mockLogger) Warn(msg string, args ...any)  { m.t.Logf("[WARN] "+msg, args...) }
func (m *mockLogger) Error(msg string, args ...any) { m.t.Logf("[ERROR] "+msg, args...) }

func TestParser_Parse_Success(t *testing.T) {
	tmpDir := t.TempDir()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	// We move from di/digen to the root of our module
	diRootPath := filepath.Dir(wd)

	goModCode := fmt.Sprintf(`
module mytestproject

go 1.22

require github.com/gymfony/di v0.0.0

replace github.com/gymfony/di => %s
`, diRootPath)

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModCode), 0o644); err != nil {
		t.Fatalf("failed to create go.mod file: %v", err)
	}

	testCode := `
package main

import "github.com/gymfony/di"

type Mailer struct{}
type UserService struct{ m *Mailer }

func NewMailer() *Mailer { return &Mailer{} }
func NewUserService(m *Mailer) *UserService { return &UserService{m: m} }

var Services = di.NewSet(
	NewMailer,
	NewUserService,
)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(testCode), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Hard-code Go to validate local replace and create go.sum
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	// We initialize environment variables and disable the proxy so that Go doesn't access the internet beyond v0.0.0.
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GO111MODULE=on")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run go mod tidy: %v, output: %s", err, string(output))
	}

	logger := NewParserLogger(&mockLogger{t: t})
	parser := NewParser(logger)

	args := []string{"-project-path", tmpDir}
	graph, err := parser.Parse(args)

	if err != nil {
		t.Fatalf("Parser.Parse returned unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("expected graph to be not nil")
	}

	// ПWe check the nodes in the graph (the module name is now mytestproject)
	mailerNode, exists := graph.Nodes["*mytestproject.Mailer"]
	if !exists {
		t.Fatalf("expected *mytestproject.Mailer to be registered, current nodes: %v", graph.Nodes)
	}
	if mailerNode.CtorName != "NewMailer" {
		t.Errorf("expected ctor NewMailer, got %s", mailerNode.CtorName)
	}

	userNode, exists := graph.Nodes["*mytestproject.UserService"]
	if !exists {
		t.Fatal("expected *mytestproject.UserService to be registered")
	}
	if len(userNode.Deps) != 1 || userNode.Deps[0] != "*mytestproject.Mailer" {
		t.Errorf("expected dependency [*mytestproject.Mailer], got %v", userNode.Deps)
	}
}

func TestParser_Parse_EmptyPath(t *testing.T) {
	logger := NewParserLogger(&mockLogger{t: t})
	parser := NewParser(logger)

	args := []string{"-project-path", ""}
	_, err := parser.Parse(args)

	if err == nil {
		t.Error("expected error when project path is empty, got nil")
	}
}
