package commands

// helpCommand shows available slash commands.
type helpCommand struct{}

func (c *helpCommand) Name() string        { return "help" }
func (c *helpCommand) Description() string { return "Show available commands" }

func (c *helpCommand) Execute(args string) Result {
	return Result{
		Toast: "✓ Showing help",
	}
}

// NewHelpCommand creates a new help command instance.
func NewHelpCommand() Command {
	return &helpCommand{}
}
