package commands

// settingsCommand opens settings.
type settingsCommand struct{}

func (c *settingsCommand) Name() string        { return "settings" }
func (c *settingsCommand) Description() string { return "Open settings" }

func (c *settingsCommand) Execute(args string) Result {
	return Result{
		Chat:  "Settings panel opened.",
		Toast: "✓ Settings",
	}
}

// NewSettingsCommand creates a new settings command instance.
func NewSettingsCommand() Command {
	return &settingsCommand{}
}
