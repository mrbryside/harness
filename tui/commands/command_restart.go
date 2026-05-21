package commands

// restartCommand restarts the session.
type restartCommand struct{}

func (c *restartCommand) Name() string        { return "restart" }
func (c *restartCommand) Description() string { return "Restart session" }

func (c *restartCommand) Execute(args string) Result {
	return Result{
		Chat:  "Session restarted.",
		Toast: "✓ Restarted",
	}
}

// NewRestartCommand creates a new restart command instance.
func NewRestartCommand() Command {
	return &restartCommand{}
}
