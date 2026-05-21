package commands

// buildCommand switches to build mode.
type buildCommand struct{}

func (c *buildCommand) Name() string        { return "build" }
func (c *buildCommand) Description() string { return "Switch to build mode" }

func (c *buildCommand) Execute(args string) Result {
	return Result{
		Chat:  "Switched to Build mode.",
		Toast: "✓ Build mode",
	}
}

// NewBuildCommand creates a new build command instance.
func NewBuildCommand() Command {
	return &buildCommand{}
}
