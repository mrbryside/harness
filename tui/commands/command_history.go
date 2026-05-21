package commands

// historyCommand shows chat history.
type historyCommand struct{}

func (c *historyCommand) Name() string        { return "history" }
func (c *historyCommand) Description() string { return "Show chat history" }

func (c *historyCommand) Execute(args string) Result {
	return Result{
		Chat:  "Chat history:\n[No history saved yet]",
		Toast: "✓ History shown",
	}
}

// NewHistoryCommand creates a new history command instance.
func NewHistoryCommand() Command {
	return &historyCommand{}
}
