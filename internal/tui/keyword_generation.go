package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// KeywordGenerationModel handles keyword generation input and results
type KeywordGenerationModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// Input state
	seedKeyword  string
	inputFocused bool
	generating   bool

	// Results
	results         []api.KeywordData
	selectedKeyword int
	scrollOffset    int
	generationID    string
	creditsUsed     int
	error           error

	// Credit balance tracking
	credits int // -1 means not loaded yet

	// Navigation
	quitting bool
}

// NewKeywordGenerationModel creates a new keyword generation model
func NewKeywordGenerationModel(config *Config) *KeywordGenerationModel {
	return &KeywordGenerationModel{
		config:          config,
		inputFocused:    true,
		selectedKeyword: 0,
		scrollOffset:    0,
		credits:         -1, // Not loaded yet
	}
}

// Init implements tea.Model
func (m *KeywordGenerationModel) Init() tea.Cmd {
	// Fetch credits if we have an API key
	if m.config.APIKey != "" {
		return m.fetchCredits()
	}
	return nil
}

// Update implements tea.Model
func (m *KeywordGenerationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case KeywordGenerationCompleteMsg:
		m.generating = false
		if msg.Error != nil {
			m.error = msg.Error
		} else {
			m.results = msg.Results
			m.generationID = msg.GenerationID
			m.creditsUsed = msg.CreditsUsed
			m.inputFocused = false
			m.selectedKeyword = 0
			m.scrollOffset = 0
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
func (m *KeywordGenerationModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	if m.error != nil {
		return m.renderError()
	}

	if m.generating {
		return m.renderGenerating()
	}

	if len(m.results) > 0 {
		return m.renderResults()
	}

	return m.renderInput()
}

// renderInput shows the keyword input interface
func (m *KeywordGenerationModel) renderInput() string {
	title := TitleStyle.Render("✨ Generate Keywords")

	// Input section
	inputLabel := SubtitleStyle.Render("Enter seed keyword:")

	inputStyle := InputStyle
	if m.inputFocused {
		inputStyle = FocusedInputStyle
	}

	inputBox := inputStyle.Render(fmt.Sprintf(" %s ", m.seedKeyword))
	if m.inputFocused {
		inputBox = inputStyle.Render(fmt.Sprintf(" %s█", m.seedKeyword))
	}

	// Instructions
	instructions := ContentStyle.Render("Enter a keyword to generate related SEO keywords and search volume data.")

	// Help
	help := RenderStatusBar(map[string]string{
		"Enter": "Generate keywords",
		"Esc":   "Back to keyword menu",
	}, m.credits, m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			inputLabel,
			"",
			inputBox,
			"",
			instructions,
			"",
			help,
		),
	)
}

// renderGenerating shows the generation progress
func (m *KeywordGenerationModel) renderGenerating() string {
	title := TitleStyle.Render("✨ Generate Keywords")

	seedInfo := SubtitleStyle.Render(fmt.Sprintf("Seed keyword: %s", m.seedKeyword))

	loading := InfoStatusStyle.Render("🔍 Generating keywords...")

	instructions := ContentStyle.Render("This may take a few moments while we fetch keyword data...")

	help := RenderStatusBar(map[string]string{
		"Esc": "Cancel",
	}, m.credits, m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			seedInfo,
			"",
			loading,
			"",
			instructions,
			"",
			help,
		),
	)
}

// renderResults shows the generated keywords
func (m *KeywordGenerationModel) renderResults() string {
	title := TitleStyle.Render("✨ Keyword Results")

	// Summary
	summary := m.renderSummary()

	// Keywords list
	keywordsList := m.renderKeywordsList()

	// Help
	help := RenderStatusBar(map[string]string{
		"↑↓":    "Navigate keywords",
		"Enter": "Select keyword (future: generate content)",
		"Esc":   "Back to keyword menu",
	}, m.credits, m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			summary,
			"",
			keywordsList,
			"",
			help,
		),
	)
}

// renderSummary shows generation summary
func (m *KeywordGenerationModel) renderSummary() string {
	stats := []string{
		fmt.Sprintf("Seed Keyword: %s", m.seedKeyword),
		fmt.Sprintf("Keywords Found: %d", len(m.results)),
		fmt.Sprintf("Credits Used: %d", m.creditsUsed),
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("📈 Generation Summary:"),
		"",
		ContentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, stats...)),
	)
}

// renderKeywordsList shows the list of keywords with scrolling
func (m *KeywordGenerationModel) renderKeywordsList() string {
	if len(m.results) == 0 {
		return ContentStyle.Render("No keywords found.")
	}

	maxVisible := m.getMaxVisibleItems()
	start := m.scrollOffset
	end := start + maxVisible
	if end > len(m.results) {
		end = len(m.results)
	}

	var lines []string

	// Add table header
	header := m.renderTableHeader()
	lines = append(lines, header)

	// Add scroll indicator if there are items above
	if start > 0 {
		scrollUp := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↑ More keywords above")
		lines = append(lines, scrollUp)
	}

	// Add visible keywords
	for i := start; i < end; i++ {
		line := m.renderKeywordItem(i, m.results[i])
		lines = append(lines, line)
	}

	// Add scroll indicator if there are items below
	if end < len(m.results) {
		scrollDown := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("↓ More keywords below")
		lines = append(lines, scrollDown)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubtitleStyle.Render("🔑 Keywords:"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
}

// renderTableHeader renders the table header
func (m *KeywordGenerationModel) renderTableHeader() string {
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
	keywordWidth := int(float64(totalWidth) * 0.42)    // 42% for keyword
	volumeWidth := int(float64(totalWidth) * 0.18)     // 18% for volume
	difficultyWidth := int(float64(totalWidth) * 0.18) // 18% for difficulty
	cpcWidth := int(float64(totalWidth) * 0.18)        // 18% for CPC (increased from 15%)

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
	if cpcWidth < 8 { // Increased minimum CPC width from 6 to 8
		cpcWidth = 8
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
func (m *KeywordGenerationModel) renderKeywordItem(index int, keyword api.KeywordData) string {
	// Calculate responsive column widths (must match header)
	totalWidth := m.width - 8
	if totalWidth < 70 {
		totalWidth = 70
	}

	keywordWidth := int(float64(totalWidth) * 0.42)
	volumeWidth := int(float64(totalWidth) * 0.18)
	difficultyWidth := int(float64(totalWidth) * 0.18)
	cpcWidth := int(float64(totalWidth) * 0.18)

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
	if cpcWidth < 8 {
		cpcWidth = 8
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
		// Ensure CPC text doesn't exceed column width
		if len(cpcText) > cpcWidth-1 {
			if *keyword.CPC >= 100 {
				cpcText = "$99+"
			} else {
				cpcText = fmt.Sprintf("$%.1f", *keyword.CPC)
			}
		}
	} else {
		cpcText = "/"
	}

	// Apply selection styling - ensure it covers the entire row width
	if index == m.selectedKeyword {
		// For selected items, we need to rebuild with styled columns but consistent selection background
		keywordCol := lipgloss.NewStyle().
			Width(keywordWidth).
			Background(SelectedItemStyle.GetBackground()).
			Foreground(SelectedItemStyle.GetForeground()).
			Bold(SelectedItemStyle.GetBold()).
			Render(keywordText)

		volumeCol := lipgloss.NewStyle().
			Width(volumeWidth).
			Align(lipgloss.Center).
			Background(SelectedItemStyle.GetBackground()).
			Foreground(SelectedItemStyle.GetForeground()).
			Bold(SelectedItemStyle.GetBold()).
			Render(volumeText)

		difficultyCol := lipgloss.NewStyle().
			Width(difficultyWidth).
			Align(lipgloss.Center).
			Background(SelectedItemStyle.GetBackground()).
			Foreground(SelectedItemStyle.GetForeground()).
			Bold(SelectedItemStyle.GetBold()).
			Render(difficultyText)

		cpcCol := lipgloss.NewStyle().
			Width(cpcWidth).
			Align(lipgloss.Center).
			Background(SelectedItemStyle.GetBackground()).
			Foreground(SelectedItemStyle.GetForeground()).
			Bold(SelectedItemStyle.GetBold()).
			Render(cpcText)

		// Join columns with background-styled spacers
		spacer := lipgloss.NewStyle().
			Background(SelectedItemStyle.GetBackground()).
			Render("  ")

		selectedRow := lipgloss.JoinHorizontal(lipgloss.Left,
			keywordCol,
			spacer,
			volumeCol,
			spacer,
			difficultyCol,
			spacer,
			cpcCol,
		)

		// Apply final styling to ensure consistent width and prevent text shifting
		return lipgloss.NewStyle().
			Width(totalWidth).
			Background(SelectedItemStyle.GetBackground()).
			Render(selectedRow)
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

// renderError shows error state
func (m *KeywordGenerationModel) renderError() string {
	title := ErrorStatusStyle.Render("❌ Keyword Generation Failed")
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
func (m *KeywordGenerationModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		if m.generating {
			// Cancel generation
			m.generating = false
			return m, nil
		}
		return m, func() tea.Msg { return BackToKeywordMenuMsg{} }

	case "enter":
		if m.inputFocused && !m.generating && strings.TrimSpace(m.seedKeyword) != "" {
			// Start generation
			m.generating = true
			return m, m.generateKeywords()
		} else if len(m.results) > 0 && !m.inputFocused {
			// Future: Navigate to content generation
			selectedKeyword := m.results[m.selectedKeyword]
			return m, tea.Printf("🚀 Future: Generate content for keyword '%s'", selectedKeyword.Keyword)
		}
		return m, nil

	case "up":
		if len(m.results) > 0 && !m.inputFocused {
			if m.selectedKeyword > 0 {
				m.selectedKeyword--
				// Adjust scroll if needed
				if m.selectedKeyword < m.scrollOffset {
					m.scrollOffset = m.selectedKeyword
				}
			}
		}
		return m, nil

	case "down":
		if len(m.results) > 0 && !m.inputFocused {
			if m.selectedKeyword < len(m.results)-1 {
				m.selectedKeyword++
				// Adjust scroll if needed
				maxVisible := m.scrollOffset + m.getMaxVisibleItems() - 1
				if m.selectedKeyword > maxVisible {
					m.scrollOffset = m.selectedKeyword - m.getMaxVisibleItems() + 1
				}
			}
		}
		return m, nil

	case "backspace":
		if m.inputFocused && len(m.seedKeyword) > 0 {
			m.seedKeyword = m.seedKeyword[:len(m.seedKeyword)-1]
		}
		return m, nil

	default:
		if m.inputFocused && len(msg.String()) == 1 {
			// Add character to seed keyword
			m.seedKeyword += msg.String()
		}
		return m, nil
	}
}

// generateKeywords fetches keywords from the API
func (m *KeywordGenerationModel) generateKeywords() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)

		response, err := client.GenerateKeywords(strings.TrimSpace(m.seedKeyword))
		if err != nil {
			return KeywordGenerationCompleteMsg{Error: fmt.Errorf("failed to generate keywords: %w", err)}
		}

		return KeywordGenerationCompleteMsg{
			Results:      response.Keywords,
			GenerationID: response.ID,
			CreditsUsed:  response.CreditsUsed,
		}
	}
}

// Helper methods

func (m *KeywordGenerationModel) getVolumeColor(volume int) lipgloss.Color {
	switch {
	case volume >= 10000:
		return SuccessColor
	case volume >= 1000:
		return WarningColor
	default:
		return ErrorColor
	}
}

func (m *KeywordGenerationModel) getDifficultyColor(difficulty float64) lipgloss.Color {
	switch {
	case difficulty <= 30:
		return SuccessColor
	case difficulty <= 60:
		return WarningColor
	default:
		return ErrorColor
	}
}

func (m *KeywordGenerationModel) getMaxVisibleItems() int {
	// Account for title, summary, help text, and spacing
	usedHeight := 18
	availableHeight := m.height - usedHeight
	if availableHeight <= 0 {
		availableHeight = 8
	}
	return availableHeight
}

// fetchCredits fetches the current credit balance from the API
func (m *KeywordGenerationModel) fetchCredits() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// Message types
type KeywordGenerationCompleteMsg struct {
	Results      []api.KeywordData
	GenerationID string
	CreditsUsed  int
	Error        error
}

type BackToKeywordMenuMsg struct{}
