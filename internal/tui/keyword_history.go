package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// KeywordHistoryModel displays the list of past keyword generations
type KeywordHistoryModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// History data
	generations        []api.KeywordGenerationHistoryItem
	selectedGeneration int
	loading            bool
	error              error

	// Credit balance tracking
	credits int // -1 means not loaded yet

	// Scrolling
	listScrollOffset int // Track scroll position in generation list

	// Navigation
	quitting bool
}

// NewKeywordHistoryModel creates a new keyword history model
func NewKeywordHistoryModel(config *Config) *KeywordHistoryModel {
	return &KeywordHistoryModel{
		config:             config,
		selectedGeneration: 0,
		loading:            true,
		generations:        []api.KeywordGenerationHistoryItem{},
		credits:            -1, // Not loaded yet
	}
}

// Init implements tea.Model
func (m *KeywordHistoryModel) Init() tea.Cmd {
	// Fetch both history and credits
	var cmds []tea.Cmd
	cmds = append(cmds, m.fetchKeywordHistory())
	if m.config.APIKey != "" {
		cmds = append(cmds, m.fetchCredits())
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (m *KeywordHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case KeywordHistoryLoadedMsg:
		m.loading = false
		if msg.Error != nil {
			m.error = msg.Error
		} else {
			m.generations = msg.Generations
			// Start at the top (most recent generation) and reset scroll
			m.selectedGeneration = 0
			m.listScrollOffset = 0
		}
		return m, nil

	case CreditsMsg:
		// Update credit balance
		if msg.Error == nil {
			m.credits = msg.Credits
		} else {
			// Keep current credits value on error
			m.credits = -1
		}
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m *KeywordHistoryModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	if m.error != nil {
		return m.renderError()
	}

	if m.loading {
		return m.renderLoading()
	}

	if len(m.generations) == 0 {
		return m.renderEmpty()
	}

	return m.renderGenerationList()
}

// renderLoading shows a loading state
func (m *KeywordHistoryModel) renderLoading() string {
	title := TitleStyle.Render("📊 Keyword History")
	loading := InfoStatusStyle.Render("🔍 Loading your keyword generation history...")
	help := RenderStatusBar(map[string]string{"Esc": "Back to keyword menu"}, m.credits, m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			loading,
			"",
			help,
		),
	)
}

// renderEmpty shows when no generations are found
func (m *KeywordHistoryModel) renderEmpty() string {
	title := TitleStyle.Render("📊 Keyword History")
	empty := ContentStyle.Render("No keyword generations found. Generate your first keywords to see them here!")
	help := RenderStatusBar(map[string]string{"Esc": "Back to keyword menu"}, m.credits, m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			empty,
			"",
			help,
		),
	)
}

// renderGenerationList shows the list of generations
func (m *KeywordHistoryModel) renderGenerationList() string {
	title := TitleStyle.Render("📊 Keyword History")

	// Summary
	summary := m.renderSummary()

	// Generation list
	generationList := m.renderGenerationItems()

	// Help
	help := RenderStatusBar(map[string]string{
		"↑↓":    "Navigate/Scroll",
		"Enter": "View Keywords",
		"Esc":   "Back to keyword menu",
	}, m.credits, m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			summary,
			"",
			generationList,
			"",
			help,
		),
	)
}

// renderSummary shows generation statistics
func (m *KeywordHistoryModel) renderSummary() string {
	if len(m.generations) == 0 {
		return ""
	}

	// Calculate statistics
	var totalKeywords int

	for _, generation := range m.generations {
		totalKeywords += len(generation.Keywords)
	}

	stats := []string{
		fmt.Sprintf("Total Generations: %d", len(m.generations)),
		fmt.Sprintf("Total Keywords Generated: %d", totalKeywords),
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("📈 Summary:"),
		"",
		ContentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, stats...)),
	)
}

// renderGenerationItems renders the list of generation items
func (m *KeywordHistoryModel) renderGenerationItems() string {
	if len(m.generations) == 0 {
		return ""
	}

	maxVisible := m.getMaxVisibleItems()

	// Calculate visible range
	start := m.listScrollOffset
	end := start + maxVisible
	if end > len(m.generations) {
		end = len(m.generations)
	}

	// Ensure start is valid
	if start >= len(m.generations) {
		start = 0
		m.listScrollOffset = 0
	}

	var lines []string

	// Add scroll indicator if there are items above
	if start > 0 {
		scrollUp := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↑ More generations above")
		lines = append(lines, scrollUp)
	}

	// Add visible generation items
	for i := start; i < end; i++ {
		line := m.renderGenerationItem(i, m.generations[i])
		lines = append(lines, line)
	}

	// Add scroll indicator if there are items below
	if end < len(m.generations) {
		scrollDown := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↓ More generations below")
		lines = append(lines, scrollDown)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("🔑 Your Keyword Generations:"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
}

// renderGenerationItem renders a single generation item
func (m *KeywordHistoryModel) renderGenerationItem(index int, generation api.KeywordGenerationHistoryItem) string {
	// Parse the creation date
	generatedAt, err := time.Parse(time.RFC3339, generation.GeneratedAt)
	var dateStr string
	if err != nil {
		dateStr = generation.GeneratedAt // Fallback to raw string
	} else {
		dateStr = generatedAt.Format("Jan 02, 2006 15:04")
	}

	// Status icon
	statusIcon := m.getStatusIcon(generation.Status)

	// Build the line as plain text first
	line := fmt.Sprintf("%s %s - %d keywords - %d credits - %s",
		statusIcon,
		generation.SeedKeyword,
		len(generation.Keywords),
		generation.CreditsUsed,
		dateStr,
	)

	// Apply selection styling to the complete line
	style := ListItemStyle
	if index == m.selectedGeneration {
		style = SelectedItemStyle
	}

	return style.Render(line)
}

// getStatusIcon returns the appropriate icon for the generation status
func (m *KeywordHistoryModel) getStatusIcon(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "running":
		return "⚡"
	case "pending":
		return "⏳"
	default:
		return "❓"
	}
}

// renderError shows error state
func (m *KeywordHistoryModel) renderError() string {
	title := ErrorStatusStyle.Render("❌ Failed to Load Keyword History")
	message := ContentStyle.Render(m.error.Error())
	help := RenderStatusBar(map[string]string{"Esc": "Back to keyword menu"}, m.credits, m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			message,
			"",
			help,
		),
	)
}

// handleKeypress handles keyboard input
func (m *KeywordHistoryModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, func() tea.Msg { return BackToKeywordMenuMsg{} }

	case "up", "k":
		if m.selectedGeneration > 0 {
			m.selectedGeneration--
			// Adjust scroll offset to keep selected item visible
			if m.selectedGeneration < m.listScrollOffset {
				m.listScrollOffset = m.selectedGeneration
			}
		}
		return m, nil

	case "down", "j":
		if m.selectedGeneration < len(m.generations)-1 {
			m.selectedGeneration++
			// Adjust scroll offset to keep selected item visible
			maxVisible := m.listScrollOffset + m.getMaxVisibleItems() - 1
			if m.selectedGeneration > maxVisible {
				m.listScrollOffset = m.selectedGeneration - m.getMaxVisibleItems() + 1
			}
		}
		return m, nil

	case "enter":
		if len(m.generations) > 0 && m.selectedGeneration < len(m.generations) {
			selectedGeneration := m.generations[m.selectedGeneration]
			// Navigate to generation details view
			return m, func() tea.Msg {
				return NavigateToGenerationDetailsMsg{
					Generation: selectedGeneration,
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// fetchKeywordHistory fetches the keyword history from the API
func (m *KeywordHistoryModel) fetchKeywordHistory() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)

		history, err := client.GetKeywordHistory()
		if err != nil {
			return KeywordHistoryLoadedMsg{Error: fmt.Errorf("failed to fetch keyword history: %w", err)}
		}

		return KeywordHistoryLoadedMsg{Generations: history.Generations}
	}
}

// fetchCredits fetches the current credit balance from the API
func (m *KeywordHistoryModel) fetchCredits() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// getMaxVisibleItems calculates how many generation items can fit on screen
func (m *KeywordHistoryModel) getMaxVisibleItems() int {
	// Account for title, summary, help text, spacing, and potential scroll indicators
	// Title: ~1 line, Summary: ~5 lines, Help: ~3 lines, spacing: ~4 lines
	// Scroll indicators: ~2 lines (if needed)
	// Leave some buffer for safety
	availableHeight := m.height - 18
	if availableHeight <= 0 {
		availableHeight = 8 // Minimum reasonable height
	}
	return availableHeight
}

// Message types
type KeywordHistoryLoadedMsg struct {
	Generations []api.KeywordGenerationHistoryItem
	Error       error
}

type NavigateToGenerationDetailsMsg struct {
	Generation api.KeywordGenerationHistoryItem
}
