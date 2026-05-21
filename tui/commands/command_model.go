package commands

// modelCommand shows the current model.
type modelCommand struct{}

func (c *modelCommand) Name() string        { return "model" }
func (c *modelCommand) Description() string { return "Show current model" }

func (c *modelCommand) Execute(args string) Result {
	return Result{
		Chat:  "Current model: gpt-4o",
		Toast: "✓ Model info",
	}
}

// NewModelCommand creates a new model command instance.
func NewModelCommand() Command {
	return &modelCommand{}
}
