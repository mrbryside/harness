package commands

// themeCommand changes the color theme.
type themeCommand struct{}

func (c *themeCommand) Name() string        { return "theme" }
func (c *themeCommand) Description() string { return "Change color theme" }

func (c *themeCommand) Execute(args string) Result {
	return Result{
		Chat:  "Available themes: dark, light",
		Toast: "✓ Theme list",
	}
}

// NewThemeCommand creates a new theme command instance.
func NewThemeCommand() Command {
	return &themeCommand{}
}
