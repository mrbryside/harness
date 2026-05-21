package commands

// clearCommand clears the chat history.
type clearCommand struct{}

func (c *clearCommand) Name() string        { return "clear" }
func (c *clearCommand) Description() string { return "Clear chat history" }

func (c *clearCommand) Execute(args string) Result {
	return Result{
		Chat:  "Chat history cleared.",
		Toast: "✓ Chat cleared",
	}
}

// NewClearCommand creates a new clear command instance.
func NewClearCommand() Command {
	return &clearCommand{}
}
