package commands

// quitCommand exits the application.
type quitCommand struct{}

func (c *quitCommand) Name() string        { return "quit" }
func (c *quitCommand) Description() string { return "Exit application" }

func (c *quitCommand) Execute(args string) Result {
	return Result{
		Chat:  "Goodbye!",
		Toast: "✓ Exiting...",
	}
}

// NewQuitCommand creates a new quit command instance.
func NewQuitCommand() Command {
	return &quitCommand{}
}
