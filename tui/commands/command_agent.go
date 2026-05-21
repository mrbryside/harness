package commands

// agentCommand switches the AI agent.
type agentCommand struct{}

func (c *agentCommand) Name() string        { return "agent" }
func (c *agentCommand) Description() string { return "Switch AI agent" }

func (c *agentCommand) Execute(args string) Result {
	return Result{
		Chat:  "Available agents: coder, reviewer, architect",
		Toast: "✓ Agent list",
	}
}

// NewAgentCommand creates a new agent command instance.
func NewAgentCommand() Command {
	return &agentCommand{}
}
