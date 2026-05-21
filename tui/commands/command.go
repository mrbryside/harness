package commands

// Result is what a command produces after execution.
type Result struct {
	// Chat is the text appended to the chat as a system message.
	// Empty means nothing is appended.
	Chat string
	// Toast is a transient notification shown at the top-right of the chat.
	// Empty means no toast is shown.
	Toast string
}

// Command is a slash-command that can be executed in the TUI.
type Command interface {
	// Name returns the command name without the leading slash (e.g. "help").
	Name() string
	// Description returns a short description shown in the autocomplete popup.
	Description() string
	// Execute runs the command and returns a result.
	Execute(args string) Result
}

// Registry holds all registered slash commands.
type Registry struct {
	cmds map[string]Command
}

// NewRegistry creates an empty command registry.
func NewRegistry() *Registry {
	return &Registry{
		cmds: make(map[string]Command),
	}
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd Command) {
	r.cmds[cmd.Name()] = cmd
}

// Get returns a command by name. The name should not include the leading slash.
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.cmds[name]
	return cmd, ok
}

// All returns all registered commands.
func (r *Registry) All() []Command {
	list := make([]Command, 0, len(r.cmds))
	for _, cmd := range r.cmds {
		list = append(list, cmd)
	}
	return list
}
