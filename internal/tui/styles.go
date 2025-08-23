package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary colors
	PrimaryColor   = lipgloss.Color("#7C3AED") // Purple
	SecondaryColor = lipgloss.Color("#06B6D4") // Cyan
	AccentColor    = lipgloss.Color("#F59E0B") // Amber

	// Status colors
	SuccessColor = lipgloss.Color("#10B981") // Green
	WarningColor = lipgloss.Color("#F59E0B") // Amber
	ErrorColor   = lipgloss.Color("#EF4444") // Red
	InfoColor    = lipgloss.Color("#3B82F6") // Blue

	// Neutral colors
	TextColor   = lipgloss.Color("")        // Use terminal default
	MutedColor  = lipgloss.Color("#9CA3AF") // Gray
	BorderColor = lipgloss.Color("#6B7280") // Dark gray

	// Interactive colors
	SelectedColor = lipgloss.Color("#8B5CF6") // Purple
	FocusedColor  = lipgloss.Color("#06B6D4") // Cyan
)

// Base styles
var (
	// Main container styles - consistent left alignment
	AppStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title and header styles - left aligned for consistency
	TitleStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true).
			Align(lipgloss.Left). // Changed from Center to Left
			Margin(0, 0, 1, 0)    // Reduced margins for consistency

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Italic(true).
			Align(lipgloss.Left). // Changed from Center to Left
			Margin(0, 0, 1, 0)    // Consistent margins

	// Content styles - consistent padding and margins
	ContentStyle = lipgloss.NewStyle().
			Padding(0, 1). // Reduced padding for consistency
			Margin(0, 0)   // Reduced margins

	// Status styles - unchanged
	StatusStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder())

	SuccessStatusStyle = StatusStyle.
				Background(SuccessColor).
				Foreground(lipgloss.Color("#FFFFFF"))

	WarningStatusStyle = StatusStyle.
				Background(WarningColor).
				Foreground(lipgloss.Color("#000000"))

	ErrorStatusStyle = StatusStyle.
				Background(ErrorColor).
				Foreground(lipgloss.Color("#FFFFFF"))

	InfoStatusStyle = StatusStyle.
			Background(InfoColor).
			Foreground(lipgloss.Color("#FFFFFF"))

	// Progress styles - unchanged
	ProgressBarStyle = lipgloss.NewStyle().
				Background(BorderColor).
				Border(lipgloss.RoundedBorder())

	ProgressFillStyle = lipgloss.NewStyle().
				Background(PrimaryColor).
				Border(lipgloss.RoundedBorder())

	// List and navigation styles - ensure consistent behavior
	ListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2).
			Margin(0, 0) // Reduced margins

	ListItemStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 0)

	SelectedItemStyle = ListItemStyle.
				Background(SelectedColor).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	FocusedItemStyle = ListItemStyle.
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(FocusedColor).
				Padding(0, 1)

	// Button styles - unchanged
	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor)

	ActiveButtonStyle = ButtonStyle.
				BorderForeground(PrimaryColor).
				Background(PrimaryColor).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	// Input styles - unchanged
	InputStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderColor)

	FocusedInputStyle = InputStyle.
				BorderForeground(FocusedColor)

	// Help styles - left aligned and consistent
	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Margin(0, 0).        // Consistent margins
			Align(lipgloss.Left) // Changed from Center to Left

	KeyStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	// Box styles - consistent margins
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2).
			Margin(0, 0) // Consistent margins

	HighlightBoxStyle = BoxStyle.
				BorderForeground(PrimaryColor)

	// Modal styles - unchanged
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(2, 4).
			Margin(2, 4)

	OverlayStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")).
			Width(120).
			Height(100)

	// Notification styles
	NotificationStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Margin(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(BorderColor)

	SuccessNotificationStyle = NotificationStyle.
					Background(SuccessColor).
					Foreground(lipgloss.Color("#FFFFFF")).
					BorderForeground(SuccessColor)

	ErrorNotificationStyle = NotificationStyle.
				Background(ErrorColor).
				Foreground(lipgloss.Color("#FFFFFF")).
				BorderForeground(ErrorColor)

	InfoNotificationStyle = NotificationStyle.
				Background(InfoColor).
				Foreground(lipgloss.Color("#FFFFFF")).
				BorderForeground(InfoColor)
)

// Status icons
const (
	IconPending    = "⏳"
	IconCrawling   = "🔍"
	IconAnalyzing  = "⚡"
	IconCompleted  = "✅"
	IconError      = "❌"
	IconWarning    = "⚠️"
	IconInfo       = "ℹ️"
	IconSuccess    = "🎉"
	IconLoading    = "⏳"
	IconArrowUp    = "↑"
	IconArrowDown  = "↓"
	IconArrowLeft  = "←"
	IconArrowRight = "→"
	IconEnter      = "↵"
	IconSpace      = "␣"
	IconEscape     = "⎋"
)

// Helper functions for styling
func StatusIcon(status PageStatus) string {
	switch status {
	case StatusPending:
		return IconPending
	case StatusCrawling:
		return IconCrawling
	case StatusAnalyzing:
		return IconAnalyzing
	case StatusCompleted:
		return IconCompleted
	case StatusError:
		return IconError
	case StatusWarning:
		return IconWarning
	default:
		return IconPending
	}
}

func StatusColor(status PageStatus) lipgloss.Color {
	switch status {
	case StatusCompleted:
		return SuccessColor
	case StatusError:
		return ErrorColor
	case StatusWarning:
		return WarningColor
	case StatusCrawling, StatusAnalyzing:
		return InfoColor
	default:
		return MutedColor
	}
}

func ScoreColor(score int) lipgloss.Color {
	switch {
	case score >= 80:
		return SuccessColor
	case score >= 60:
		return WarningColor
	default:
		return ErrorColor
	}
}

// RenderButton renders a button with consistent highlighting behavior
func RenderButton(text string, isSelected bool) string {
	if isSelected {
		return lipgloss.NewStyle().
			Background(SelectedColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Render(text)
	} else {
		return lipgloss.NewStyle().
			Foreground(TextColor).
			Render(text)
	}
}

// Progress bar rendering
func RenderProgressBar(current, total int, width int) string {
	if total == 0 {
		return ProgressBarStyle.Width(width).Render(lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, "No data"))
	}

	percentage := float64(current) / float64(total)
	filledWidth := int(float64(width) * percentage)

	filled := ProgressFillStyle.Width(filledWidth).Render("")
	empty := ProgressBarStyle.Width(width - filledWidth).Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top, filled, empty)
}

// Enhanced key binding help - consistent left alignment
func RenderKeyHelp(bindings map[string]string) string {
	// Create a consistent container with left alignment
	container := lipgloss.NewStyle().
		Align(lipgloss.Left). // Changed from Center to Left
		Margin(0, 0)          // Consistent margins

	// Define a consistent order for common keys
	keyOrder := []string{
		"↑↓",     // Vertical navigation
		"←→",     // Horizontal navigation
		"Enter",  // Primary action
		"e",      // Export
		"n",      // New/Create
		"h",      // History/Help
		"r",      // Retry/Refresh
		"b",      // Brief/Back (context-specific)
		"Tab",    // Tab navigation
		"Ctrl+U", // Clear input
		"Esc",    // Back/cancel
		"Ctrl+C", // Force quit
		"q",      // Quit
	}

	// Create ordered pairs
	var pairs []string
	usedKeys := make(map[string]bool)

	// First, add keys in the defined order
	for _, key := range keyOrder {
		if desc, exists := bindings[key]; exists {
			keyText := KeyStyle.Render(key)
			descText := lipgloss.NewStyle().
				Foreground(MutedColor).
				Render(desc)

			pair := lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Width(12).Render(keyText),
				descText,
			)
			pairs = append(pairs, pair)
			usedKeys[key] = true
		}
	}

	// Then add any remaining keys that weren't in the predefined order
	for key, desc := range bindings {
		if !usedKeys[key] {
			keyText := KeyStyle.Render(key)
			descText := lipgloss.NewStyle().
				Foreground(MutedColor).
				Render(desc)

			pair := lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Width(12).Render(keyText),
				descText,
			)
			pairs = append(pairs, pair)
		}
	}

	// Join all pairs with proper spacing
	return container.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			pairs...,
		),
	)
}

// RenderStatusBar renders help text and credit balance in a consistent status bar
func RenderStatusBar(helpBindings map[string]string, credits int, hasAPIKey bool) string {
	// Render help text on the left
	helpText := RenderKeyHelp(helpBindings)

	// Render credit balance on the right
	var creditText string
	if !hasAPIKey {
		creditText = lipgloss.NewStyle().
			Foreground(MutedColor).
			Render("Credits: N/A (no API key)")
	} else if credits < 0 {
		// Credits not loaded yet or error
		creditText = lipgloss.NewStyle().
			Foreground(MutedColor).
			Render("Credits: Loading...")
	} else {
		// Color-code the credit balance
		var creditColor lipgloss.Color
		switch {
		case credits >= 100:
			creditColor = SuccessColor // Green for healthy balance
		case credits >= 20:
			creditColor = WarningColor // Yellow for moderate balance
		default:
			creditColor = ErrorColor // Red for low balance
		}

		creditText = lipgloss.NewStyle().
			Foreground(creditColor).
			Bold(true).
			Render(fmt.Sprintf("Credits: %d", credits))
	}

	// Create a flexible layout that puts help on left and credits on right
	// We'll use a simple approach - left align help, then add credits with spacing
	return lipgloss.JoinVertical(lipgloss.Left,
		helpText,
		"",
		lipgloss.NewStyle().
			Align(lipgloss.Right).
			Render(creditText),
	)
}

// RenderNotification renders a notification message with appropriate styling
func RenderNotification(msg NotificationMsg) string {
	var style lipgloss.Style
	var icon string

	switch msg.Type {
	case NotificationSuccess:
		style = SuccessNotificationStyle
		icon = "✅"
	case NotificationError:
		style = ErrorNotificationStyle
		icon = "❌"
	case NotificationInfo:
		style = InfoNotificationStyle
		icon = "ℹ️"
	default:
		style = NotificationStyle
		icon = "•"
	}

	return style.Render(fmt.Sprintf("%s %s", icon, msg.Message))
}
