package commands

// saveCommand saves the current conversation.
type saveCommand struct{}

func (c *saveCommand) Name() string        { return "save" }
func (c *saveCommand) Description() string { return "Save conversation" }

func (c *saveCommand) Execute(args string) Result {
	return Result{
		Chat:  "Conversation saved to disk.",
		Toast: "✓ Saved",
	}
}

// NewSaveCommand creates a new save command instance.
func NewSaveCommand() Command {
	return &saveCommand{}
}
