package di_test

import (
	"testing"

	"github.com/gymfony/di"
)

func TestNew(t *testing.T) {
	c := di.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}
