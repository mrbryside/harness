package styles

import "charm.land/lipgloss/v2"

// GlamourStyle is the Glamour markdown theme used for assistant messages.
const GlamourStyle = "dark"

// Color palette — all colors must be defined here. Never hardcode hex elsewhere.
var (
	// Background is the main chat area — the darkest tone.
	Background = lipgloss.Color("#000000")
	// PanelBg is the slightly lighter tone used for sidebar, input, and user-message panels.
	PanelBg         = lipgloss.Color("#1a1b26")
	UserBorder      = lipgloss.Color("#3d59a1") // blue left accent
	AssistantText   = lipgloss.Color("#c0caf5")
	SidebarLabel    = lipgloss.Color("#ffffff")
	SidebarValue    = lipgloss.Color("#a9b1d6")
	StatusBarBg     = lipgloss.Color("#0d0d12")
	StatusBarAccent = lipgloss.Color("#7aa2f7")
	ConnectedDot    = lipgloss.Color("#9ece6a")
	// Agent mode label colors (shown in the input footer).
	ModeBuildColor = lipgloss.Color("#7aa2f7") // blue — write/execute mode
	ModePlanColor  = lipgloss.Color("#bb9af7") // purple — plan/think mode
)
