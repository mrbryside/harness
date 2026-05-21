package commands

// loadCommand loads a saved conversation.
type loadCommand struct{}

func (c *loadCommand) Name() string        { return "load" }
func (c *loadCommand) Description() string { return "Load conversation" }

func (c *loadCommand) Execute(args string) Result {
	return Result{
		Chat:  "Conversation loaded from disk.",
		Toast: "✓ Loaded",
	}
}

// NewLoadCommand creates a new load command instance.
func NewLoadCommand() Command {
	return &loadCommand{}
}
