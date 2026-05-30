package styles

import (
	"fmt"
	"image/color"
	"strconv"

	"charm.land/lipgloss/v2"
)

// GlamourStyle is the Glamour markdown theme used for assistant messages.
const GlamourStyle = "dark"

// ColorToAnsiBg converts a color.Color to an ANSI 24-bit background SGR
// escape sequence.
func ColorToAnsiBg(c color.Color) string {
	return hexToAnsiBg(hexFromColor(c))
}

// hexToAnsiBg converts a 6-digit hex colour string to an ANSI 24-bit
// background SGR escape sequence.
func hexToAnsiBg(hex string) string {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return ""
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

// hexFromColor returns the 6-digit hex representation of a color.Color.
func hexFromColor(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// Color palette — Monokai-inspired dark theme.
// All colors must be defined here. Never hardcode hex elsewhere.
var (
	// Background — sidebar & user chat panels.
	Background = lipgloss.Color("#141414")

	// ChatBackground — main chat area and status bar.
	ChatBackground = lipgloss.Color("#0a0a0a")

	// PanelBg — input area.
	PanelBg = lipgloss.Color("#1e1e1e")

	// UserBorder — blue accent bar on user messages.
	UserBorder = lipgloss.Color("#6b9aee")

	// AssistantText — warm off-white, the classic Monokai foreground.
	AssistantText = lipgloss.Color("#f8f8f2")

	// Sidebar labels are bold white; values are muted Monokai gray.
	SidebarLabel = lipgloss.Color("#f8f8f2")
	SidebarValue = lipgloss.Color("#75715e") // Monokai comment gray

	// Status bar — flush with chat background.
	StatusBarBg     = lipgloss.Color("#0a0a0a")
	StatusBarAccent = lipgloss.Color("#66d9ef") // Monokai cyan

	// Small green dot for "connected" indicators — Monokai green.
	ConnectedDot = lipgloss.Color("#a6e22e")

	// Agent mode labels in the input footer.
	ModeBuildColor = lipgloss.Color("#6b9aee") // Blue — write/execute mode
	ModePlanColor  = lipgloss.Color("#ae81ff") // Monokai purple — plan/think mode

	// Selection highlight when dragging to copy text.
	SelectionBg = lipgloss.Color("#49483e") // Monokai selection gray

	// AccentOrange — Monokai orange used for active todos `[•]` and highlights.
	AccentOrange = lipgloss.Color("#fd971f")

	// Peach — soft warm background for selected autocomplete rows.
	Peach = lipgloss.Color("#f5c2a8")
)

// ANSI SGR sequences derived from the palette above.
// Keep these in sync with the lipgloss.Colors above.
var (
	BackgroundSGR  = hexToAnsiBg("#141414")
	ChatBgSGR      = hexToAnsiBg("#0a0a0a")
	PanelBgSGR     = hexToAnsiBg("#1e1e1e")
	SelectionBgSGR = hexToAnsiBg("#49483e")
)
