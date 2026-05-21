package commands

import "testing"

// Registry starts empty.
func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()
	if len(r.All()) != 0 {
		t.Errorf("expected empty registry, got %d commands", len(r.All()))
	}
}

// Register adds a command to the registry.
func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	cmd := &helpCommand{}
	r.Register(cmd)
	if len(r.All()) != 1 {
		t.Errorf("expected 1 command, got %d", len(r.All()))
	}
}

// Get returns a command by name.
func TestRegistryGet(t *testing.T) {
	r := NewRegistry()
	cmd := &helpCommand{}
	r.Register(cmd)
	got, ok := r.Get("help")
	if !ok {
		t.Fatal("expected to find 'help' command")
	}
	if got.Name() != "help" {
		t.Errorf("expected name 'help', got %q", got.Name())
	}
}

// Get returns false for unknown commands.
func TestRegistryGetUnknown(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("unknown")
	if ok {
		t.Error("expected false for unknown command")
	}
}
