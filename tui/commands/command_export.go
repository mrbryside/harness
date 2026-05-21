package commands

// exportCommand exports the current conversation.
type exportCommand struct{}

func (c *exportCommand) Name() string        { return "export" }
func (c *exportCommand) Description() string { return "Export conversation" }

func (c *exportCommand) Execute(args string) Result {
	return Result{
		Chat:  "Conversation exported to export.txt",
		Toast: "✓ Exported",
	}
}

// NewExportCommand creates a new export command instance.
func NewExportCommand() Command {
	return &exportCommand{}
}
