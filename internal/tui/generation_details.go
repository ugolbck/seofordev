package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// GenerationDetailsModel displays the keywords for a selected generation
type GenerationDetailsModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// Generation data
	generation      api.KeywordGenerationHistoryItem
	selectedKeyword int
	scrollOffset    int

	// Credit balance tracking
	credits int // -1 means not loaded yet

	// Navigation
	quitting bool
}

// NewGenerationDetailsModel creates a new generation details model
func NewGenerationDetailsModel(config *Config, generation api.KeywordGenerationHistoryItem) *GenerationDetailsModel {
	return &GenerationDetailsModel{
		config:          config,
		generation:      generation,
		selectedKeyword: 0,
		scrollOffset:    0,
		credits:         -1, // Not loaded yet
	}
}

// Init implements tea.Model
func (m *GenerationDetailsModel) Init() tea.Cmd {
	// Fetch credits if we have an API key
	if m.config.APIKey != "" {
		return m.fetchCredits()
	}
	return nil
}

// fetchCredits fetches the current credit balance from the API
func (m *GenerationDetailsModel) fetchCredits() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// Update implements tea.Model
func (m *GenerationDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

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
func (m *GenerationDetailsModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	if len(m.generation.Keywords) == 0 {
		return m.renderEmpty()
	}

	return m.renderKeywordsList()
}

// renderEmpty shows when there are no keywords
func (m *GenerationDetailsModel) renderEmpty() string {
	title := TitleStyle.Render("🔑 Generation Details")
	empty := ContentStyle.Render("No keywords found for this generation.")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			m.renderGenerationSummary(),
			"",
			empty,
			"",
			m.renderCompactHelp(),
		),
	)
}

// renderKeywordsList renders the main keywords list view
func (m *GenerationDetailsModel) renderKeywordsList() string {
	title := TitleStyle.Render("🔑 Generation Details")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			m.renderGenerationSummary(),
			"",
			m.renderKeywords(),
			"",
			m.renderCompactHelp(),
		),
	)
}

// renderGenerationSummary shows generation information
func (m *GenerationDetailsModel) renderGenerationSummary() string {
	generatedAt, _ := time.Parse(time.RFC3339, m.generation.GeneratedAt)
	dateStr := generatedAt.Format("Jan 2, 2006 at 3:04 PM")

	summary := lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("Generation Information"),
		ContentStyle.Render(fmt.Sprintf("Seed Keyword: %s", m.generation.SeedKeyword)),
		ContentStyle.Render(fmt.Sprintf("Generated: %s", dateStr)),
		ContentStyle.Render(fmt.Sprintf("Status: %s", m.generation.Status)),
		ContentStyle.Render(fmt.Sprintf("Keywords Found: %d", len(m.generation.Keywords))),
		ContentStyle.Render(fmt.Sprintf("Credits Used: %d", m.generation.CreditsUsed)),
	)

	return BoxStyle.Render(summary)
}

// renderKeywords renders the list of keywords
func (m *GenerationDetailsModel) renderKeywords() string {
	title := SubtitleStyle.Render("Keywords")

	// Add table header
	header := m.renderTableHeader()

	maxVisible := m.getMaxVisibleItems()
	start := m.scrollOffset
	end := start + maxVisible
	if end > len(m.generation.Keywords) {
		end = len(m.generation.Keywords)
	}

	var items []string
	items = append(items, header) // Add header first

	// Add scroll indicator if there are items above
	if start > 0 {
		scrollUp := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↑ More keywords above")
		items = append(items, scrollUp)
	}

	for i := start; i < end; i++ {
		keyword := m.generation.Keywords[i]
		isSelected := i == m.selectedKeyword
		items = append(items, m.renderKeywordItem(i, keyword, isSelected))
	}

	// Add scroll indicator if there are items below
	if end < len(m.generation.Keywords) {
		scrollDown := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↓ More keywords below")
		items = append(items, scrollDown)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, "", lipgloss.JoinVertical(lipgloss.Left, items...))
}

// renderTableHeader renders the table header
func (m *GenerationDetailsModel) renderTableHeader() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Border(lipgloss.Border{
			Top:    "─",
			Bottom: "─",
		}, false, false, true, false)

	// Calculate responsive column widths based on terminal width
	totalWidth := m.width - 8 // Account for padding and borders
	if totalWidth < 70 {
		totalWidth = 70 // Minimum width
	}

	// Define column widths (proportional to content importance)
	keywordWidth := int(float64(totalWidth) * 0.45)    // 45% for keyword
	volumeWidth := int(float64(totalWidth) * 0.18)     // 18% for volume
	difficultyWidth := int(float64(totalWidth) * 0.18) // 18% for difficulty
	cpcWidth := int(float64(totalWidth) * 0.15)        // 15% for CPC

	// Ensure minimum widths
	if keywordWidth < 20 {
		keywordWidth = 20
	}
	if volumeWidth < 8 {
		volumeWidth = 8
	}
	if difficultyWidth < 8 {
		difficultyWidth = 8
	}
	if cpcWidth < 6 {
		cpcWidth = 6
	}

	// Create header row
	keywordCol := lipgloss.NewStyle().Width(keywordWidth).Render("Keyword")
	volumeCol := lipgloss.NewStyle().Width(volumeWidth).Align(lipgloss.Center).Render("Volume")
	difficultyCol := lipgloss.NewStyle().Width(difficultyWidth).Align(lipgloss.Center).Render("Difficulty")
	cpcCol := lipgloss.NewStyle().Width(cpcWidth).Align(lipgloss.Center).Render("CPC")

	headerRow := lipgloss.JoinHorizontal(lipgloss.Left,
		keywordCol,
		"  ",
		volumeCol,
		"  ",
		difficultyCol,
		"  ",
		cpcCol,
	)

	return headerStyle.Render(headerRow)
}

// renderKeywordItem renders a single keyword item in table format
func (m *GenerationDetailsModel) renderKeywordItem(index int, keyword api.KeywordData, isSelected bool) string {
	// Calculate responsive column widths
	totalWidth := m.width - 8
	if totalWidth < 70 {
		totalWidth = 70
	}

	keywordWidth := int(float64(totalWidth) * 0.45)
	volumeWidth := int(float64(totalWidth) * 0.18)
	difficultyWidth := int(float64(totalWidth) * 0.18)
	cpcWidth := int(float64(totalWidth) * 0.15)

	// Ensure minimum widths
	if keywordWidth < 20 {
		keywordWidth = 20
	}
	if volumeWidth < 8 {
		volumeWidth = 8
	}
	if difficultyWidth < 8 {
		difficultyWidth = 8
	}
	if cpcWidth < 6 {
		cpcWidth = 6
	}

	// Format keyword (truncate if too long)
	keywordText := keyword.Keyword
	if len(keywordText) > keywordWidth-2 {
		keywordText = keywordText[:keywordWidth-5] + "..."
	}

	// Format volume and handle null
	var volumeText string
	if keyword.Volume != nil {
		volumeText = fmt.Sprintf("%d", *keyword.Volume)
	} else {
		volumeText = "/"
	}

	// Format difficulty and handle null
	var difficultyText string
	if keyword.Difficulty != nil {
		difficultyText = fmt.Sprintf("%.1f", *keyword.Difficulty)
	} else {
		difficultyText = "/"
	}

	// Format CPC and handle null
	var cpcText string
	if keyword.CPC != nil {
		cpcText = fmt.Sprintf("$%.2f", *keyword.CPC)
	} else {
		cpcText = "/"
	}

	// Build the table row content (without styling)
	rowContent := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Width(keywordWidth).Render(keywordText),
		"  ",
		lipgloss.NewStyle().Width(volumeWidth).Align(lipgloss.Center).Render(volumeText),
		"  ",
		lipgloss.NewStyle().Width(difficultyWidth).Align(lipgloss.Center).Render(difficultyText),
		"  ",
		lipgloss.NewStyle().Width(cpcWidth).Align(lipgloss.Center).Render(cpcText),
	)

	// Apply selection styling - ensure it covers the entire row width
	if isSelected {
		// Use a style that ensures full width coverage and overrides all individual styling
		selectedStyle := lipgloss.NewStyle().
			Background(SelectedItemStyle.GetBackground()).
			Foreground(SelectedItemStyle.GetForeground()).
			Bold(SelectedItemStyle.GetBold()).
			Width(totalWidth) // Ensure it spans the full width

		return selectedStyle.Render(rowContent)
	}

	// For non-selected items, apply individual cell styling
	keywordCol := lipgloss.NewStyle().Width(keywordWidth).Render(keywordText)

	// Format volume with colors and handle null
	var volumeTextStyled string
	if keyword.Volume != nil {
		volumeColor := m.getVolumeColor(*keyword.Volume)
		volumeTextStyled = lipgloss.NewStyle().
			Foreground(volumeColor).
			Bold(true).
			Width(volumeWidth).
			Align(lipgloss.Center).
			Render(volumeText)
	} else {
		volumeTextStyled = lipgloss.NewStyle().
			Foreground(MutedColor).
			Width(volumeWidth).
			Align(lipgloss.Center).
			Render(volumeText)
	}

	// Format difficulty with colors and handle null
	var difficultyTextStyled string
	if keyword.Difficulty != nil {
		difficultyColor := m.getDifficultyColor(*keyword.Difficulty)
		difficultyTextStyled = lipgloss.NewStyle().
			Foreground(difficultyColor).
			Bold(true).
			Width(difficultyWidth).
			Align(lipgloss.Center).
			Render(difficultyText)
	} else {
		difficultyTextStyled = lipgloss.NewStyle().
			Foreground(MutedColor).
			Width(difficultyWidth).
			Align(lipgloss.Center).
			Render(difficultyText)
	}

	// Format CPC and handle null
	var cpcTextStyled string
	if keyword.CPC != nil {
		cpcTextStyled = lipgloss.NewStyle().
			Width(cpcWidth).
			Align(lipgloss.Center).
			Render(cpcText)
	} else {
		cpcTextStyled = lipgloss.NewStyle().
			Foreground(MutedColor).
			Width(cpcWidth).
			Align(lipgloss.Center).
			Render(cpcText)
	}

	// Build the styled table row for non-selected items
	styledRow := lipgloss.JoinHorizontal(lipgloss.Left,
		keywordCol,
		"  ",
		volumeTextStyled,
		"  ",
		difficultyTextStyled,
		"  ",
		cpcTextStyled,
	)

	// For non-selected items, use the regular list item style
	return ListItemStyle.Width(totalWidth).Render(styledRow)
}

// renderCompactHelp renders the help text
func (m *GenerationDetailsModel) renderCompactHelp() string {
	help := map[string]string{
		"↑↓":    "Navigate keywords",
		"b":     "Generate content brief",
		"Enter": "Select keyword (future: generate content)",
		"Esc":   "Back to keyword history",
		"q":     "Quit",
	}

	return RenderStatusBar(help, m.credits, m.config.APIKey != "")
}

// handleKeypress handles keyboard input
func (m *GenerationDetailsModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// Return to keyword history
		return m, func() tea.Msg {
			return BackToKeywordMenuMsg{}
		}

	case "up":
		if m.selectedKeyword > 0 {
			m.selectedKeyword--
			// Adjust scroll if needed
			if m.selectedKeyword < m.scrollOffset {
				m.scrollOffset = m.selectedKeyword
			}
		}
		return m, nil

	case "down":
		if m.selectedKeyword < len(m.generation.Keywords)-1 {
			m.selectedKeyword++
			// Adjust scroll if needed
			maxVisible := m.getMaxVisibleItems()
			if m.selectedKeyword >= m.scrollOffset+maxVisible {
				m.scrollOffset = m.selectedKeyword - maxVisible + 1
			}
		}
		return m, nil

	case "enter":
		if len(m.generation.Keywords) > 0 && m.selectedKeyword < len(m.generation.Keywords) {
			selectedKeyword := m.generation.Keywords[m.selectedKeyword]
			// Future: Navigate to content generation
			return m, tea.Printf("🚀 Future: Generate content for keyword '%s'", selectedKeyword.Keyword)
		}
		return m, nil

	case "b":
		// Generate content brief for selected keyword
		if len(m.generation.Keywords) > 0 && m.selectedKeyword < len(m.generation.Keywords) {
			selectedKeyword := m.generation.Keywords[m.selectedKeyword]
			// Navigate to content brief generation with pre-filled keyword
			generationModel := NewContentBriefGenerationModelWithKeyword(selectedKeyword.Keyword)
			generationModel.config = m.config
			return generationModel, generationModel.Init()
		}
		return m, nil
	}

	return m, nil
}

// getMaxVisibleItems calculates how many items can be displayed
func (m *GenerationDetailsModel) getMaxVisibleItems() int {
	// Account for title, summary, help text, and spacing
	// Title: ~1 line, Summary: ~6 lines, Help: ~3 lines, spacing: ~4 lines
	usedHeight := 14
	availableHeight := m.height - usedHeight

	// Each keyword item takes 1 line
	itemHeight := 1
	maxItems := availableHeight / itemHeight

	// Ensure we show at least 12-15 items if possible
	if maxItems < 12 {
		maxItems = 12
	}

	return maxItems
}

// Helper methods

func (m *GenerationDetailsModel) getVolumeColor(volume int) lipgloss.Color {
	switch {
	case volume >= 10000:
		return SuccessColor
	case volume >= 1000:
		return WarningColor
	default:
		return ErrorColor
	}
}

func (m *GenerationDetailsModel) getDifficultyColor(difficulty float64) lipgloss.Color {
	switch {
	case difficulty <= 30:
		return SuccessColor
	case difficulty <= 60:
		return WarningColor
	default:
		return ErrorColor
	}
}
