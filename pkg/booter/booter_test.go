package booter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chan-jui-huang/go-backend-package/pkg/booter/config"
)

// package-level mocks
type mockRegistrar struct {
	name  string
	calls *[]string
}

func (m *mockRegistrar) Boot()     { *m.calls = append(*m.calls, m.name+":Boot") }
func (m *mockRegistrar) Register() { *m.calls = append(*m.calls, m.name+":Register") }

type mockExec struct{ seq *[]string }

func (m *mockExec) BeforeExecute() { *m.seq = append(*m.seq, "Before") }
func (m *mockExec) Execute()       { *m.seq = append(*m.seq, "Execute") }
func (m *mockExec) AfterExecute()  { *m.seq = append(*m.seq, "After") }

func TestBootConfigRegistryEnvExpansion(t *testing.T) {
	tmp := t.TempDir()
	cfg := "database:\n  host: \"${MY_HOST}\"\n"
	if err := os.WriteFile(filepath.Join(tmp, "config.yml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	// set env
	os.Setenv("MY_HOST", "localhost")
	defer os.Unsetenv("MY_HOST")

	bc := NewConfig(tmp, "config.yml", false)
	bootConfigRegistry(bc)

	v := config.Registry.GetViper()
	got := v.GetString("database.host")
	if got != "localhost" {
		t.Fatalf("expected database.host to be 'localhost', got '%s'", got)
	}
}

func TestRegistrarExecuteOrder(t *testing.T) {
	calls := []string{}

	r1 := &mockRegistrar{name: "r1", calls: &calls}
	r2 := &mockRegistrar{name: "r2", calls: &calls}

	rc := NewRegistrarCenter([]Registrar{r1, r2})
	rc.Execute()

	expected := []string{"r1:Boot", "r1:Register", "r2:Boot", "r2:Register"}
	if len(calls) != len(expected) {
		t.Fatalf("expected calls %v, got %v", expected, calls)
	}
	for i := range expected {
		if calls[i] != expected[i] {
			t.Fatalf("at %d expected %s got %s", i, expected[i], calls[i])
		}
	}
}

func TestBootCallsHooksAndLoadsConfig(t *testing.T) {
	// prepare temp config
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "config.yml"), []byte("k: v\n"), 0644); err != nil {
		t.Fatal(err)
	}

	seq := []string{}
	loadCalled := false
	loadEnv := func() { loadCalled = true }
	newCfg := func() *Config { return NewConfig(tmp, "config.yml", false) }

	exec := &mockExec{seq: &seq}

	Boot(loadEnv, newCfg, exec)

	if !loadCalled {
		t.Fatalf("expected loadEnvFunc to be called")
	}

	expected := []string{"Before", "Execute", "After"}
	if len(seq) != len(expected) {
		t.Fatalf("expected seq %v got %v", expected, seq)
	}
	for i := range expected {
		if seq[i] != expected[i] {
			t.Fatalf("at %d expected %s got %s", i, expected[i], seq[i])
		}
	}
}
