package commands

// planCommand switches to plan mode.
type planCommand struct{}

func (c *planCommand) Name() string        { return "plan" }
func (c *planCommand) Description() string { return "Switch to plan mode" }

func (c *planCommand) Execute(args string) Result {
	return Result{
		Chat:  "Switched to Plan mode.",
		Toast: "✓ Plan mode",
	}
}

// NewPlanCommand creates a new plan command instance.
func NewPlanCommand() Command {
	return &planCommand{}
}
