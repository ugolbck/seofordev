package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ugolbck/seofordev/internal/api"
)

// ContentBriefGenerationModel handles the content brief generation process
type ContentBriefGenerationModel struct {
	width  int
	height int

	// Configuration
	config *Config

	// Input state
	keywordInput string
	inputFocused bool
	cursor       int

	// Generation state
	generating  bool
	briefID     string
	keyword     string
	brief       string
	status      string
	error       string
	creditsUsed int
	generatedAt string

	// Credit balance tracking
	credits int // -1 means not loaded yet

	// Scrolling for long briefs
	scrollOffset int

	// Spinner animation
	spinner       int
	spinnerFrames []string

	// Polling
	lastPollTime time.Time

	// Notification system
	notification     *NotificationMsg
	notificationTime time.Time

	// Navigation
	quitting bool
}

// NewContentBriefGenerationModel creates a new content brief generation model
func NewContentBriefGenerationModel() *ContentBriefGenerationModel {
	return &ContentBriefGenerationModel{
		inputFocused:  true,
		credits:       -1, // Not loaded yet
		spinnerFrames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// NewContentBriefGenerationModelWithKeyword creates a new model with pre-filled keyword
func NewContentBriefGenerationModelWithKeyword(keyword string) *ContentBriefGenerationModel {
	return &ContentBriefGenerationModel{
		keywordInput:  keyword,
		inputFocused:  false, // Start generating immediately
		credits:       -1,    // Not loaded yet
		spinnerFrames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Init implements tea.Model
func (m *ContentBriefGenerationModel) Init() tea.Cmd {
	var cmds []tea.Cmd

	// If we have a keyword and aren't focused, start generation immediately
	if m.keywordInput != "" && !m.inputFocused {
		cmds = append(cmds, m.startGeneration())
	}

	// Always fetch credits
	cmds = append(cmds, m.fetchCredits())

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// Update implements tea.Model
func (m *ContentBriefGenerationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.credits = -1
		}
		return m, nil

	case BriefGenerationStartedMsg:
		m.generating = true
		m.briefID = msg.BriefID
		m.keyword = msg.Keyword
		m.status = "pending"
		// Start both polling and spinner animation
		return m, tea.Batch(
			m.pollBriefStatus(),
			tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return TickMsg(t)
			}),
		)

	case BriefGenerationProgressMsg:
		m.status = msg.Status
		// Continue polling if not completed or failed
		if msg.Status != "completed" && msg.Status != "failed" {
			return m, m.pollBriefStatus()
		}
		return m, nil

	case BriefGenerationCompletedMsg:
		m.generating = false
		m.brief = msg.Brief
		m.status = msg.Status
		m.creditsUsed = msg.CreditsUsed
		// Refresh credits since we used some
		return m, m.fetchCredits()

	case BriefGenerationFailedMsg:
		m.generating = false
		m.error = msg.Error
		m.status = "failed"
		return m, nil

	case TickMsg:
		// Handle spinner animation and polling
		if m.generating {
			// Update spinner
			if len(m.spinnerFrames) > 0 {
				m.spinner = (m.spinner + 1) % len(m.spinnerFrames)
			}

			// Handle polling if enough time has passed
			if time.Since(m.lastPollTime) >= 500*time.Millisecond {
				return m, tea.Batch(
					m.pollBriefStatus(),
					tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
						return TickMsg(t)
					}),
				)
			}

			// Continue ticking for spinner animation
			return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return TickMsg(t)
			})
		}
		return m, nil

	case NotificationMsg:
		// Store notification and timestamp for auto-hide
		m.notification = &msg
		m.notificationTime = time.Now()
		// Auto-hide notification after 3 seconds
		return m, tea.Tick(time.Second*3, func(time.Time) tea.Msg {
			return HideNotificationMsg{}
		})

	case HideNotificationMsg:
		// Clear notification if it's been long enough
		if time.Since(m.notificationTime) >= time.Second*3 {
			m.notification = nil
		}
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m *ContentBriefGenerationModel) View() string {
	if m.quitting {
		return AppStyle.Render("Thanks for using SEO CLI! 👋")
	}

	if m.generating {
		return m.renderGenerating()
	}

	if m.brief != "" {
		return m.renderBrief()
	}

	if m.error != "" {
		return m.renderError()
	}

	return m.renderInput()
}

// renderInput shows the keyword input form
func (m *ContentBriefGenerationModel) renderInput() string {
	title := TitleStyle.Render("📄 Generate Content Brief")

	// API Key status
	var apiStatus string
	if m.config != nil && m.config.APIKey != "" {
		apiStatus = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Render("✅ API Key configured")
	} else {
		apiStatus = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Render("❌ API Key not set (go to Settings)")
	}

	// Input field
	prompt := ContentStyle.Render("Enter keyword phrase:")

	var input string
	if m.inputFocused {
		// Show cursor
		beforeCursor := m.keywordInput[:m.cursor]
		afterCursor := m.keywordInput[m.cursor:]
		input = beforeCursor + "█" + afterCursor
	} else {
		input = m.keywordInput
	}

	inputField := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(0, 1).
		Width(50).
		Render(input)

	// Instructions
	instructions := ContentStyle.Render("Press Enter to generate brief, or Esc to go back")

	// Status bar
	statusBar := RenderStatusBar(map[string]string{
		"Enter": "Generate Brief",
		"Esc":   "Back",
	}, m.credits, m.config != nil && m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			apiStatus,
			"",
			prompt,
			inputField,
			"",
			instructions,
			"",
			statusBar,
		),
	)
}

// renderGenerating shows the generation progress
func (m *ContentBriefGenerationModel) renderGenerating() string {
	title := TitleStyle.Render("📄 Generating Content Brief")

	keyword := ContentStyle.Render(fmt.Sprintf("Keyword: %s", m.keyword))

	var statusText string
	switch m.status {
	case "pending":
		statusText = "⏳ Starting generation..."
	case "processing":
		statusText = "🤖 AI is generating your content brief..."
	default:
		statusText = fmt.Sprintf("Status: %s", m.status)
	}

	status := ContentStyle.Render(statusText)

	// Animated spinner
	var spinnerIcon string
	if len(m.spinnerFrames) > 0 {
		spinnerIcon = m.spinnerFrames[m.spinner]
	} else {
		spinnerIcon = "⠋" // Fallback
	}
	spinner := lipgloss.NewStyle().
		Foreground(AccentColor).
		Render(fmt.Sprintf("%s Processing...", spinnerIcon))

	instructions := ContentStyle.Render("Please wait while we generate your content brief...")

	// Status bar
	statusBar := RenderStatusBar(map[string]string{
		"Esc": "Cancel",
	}, m.credits, m.config != nil && m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			keyword,
			status,
			"",
			spinner,
			"",
			instructions,
			"",
			statusBar,
		),
	)
}

// renderBrief shows the generated brief
func (m *ContentBriefGenerationModel) renderBrief() string {
	title := TitleStyle.Render("📄 Content Brief Generated")

	keyword := ContentStyle.Render(fmt.Sprintf("Keyword: %s", m.keyword))
	credits := ContentStyle.Render(fmt.Sprintf("Credits used: %d", m.creditsUsed))

	// Brief content with scrollable area
	briefTitle := SubtitleStyle.Render("Generated Brief:")

	// Calculate available height for content
	contentHeight := m.height - 15 // Account for title, metadata, status bar, etc.
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Split content into lines and handle scrolling
	contentLines := strings.Split(m.brief, "\n")

	// Apply scroll offset
	startLine := m.scrollOffset
	endLine := startLine + contentHeight
	if endLine > len(contentLines) {
		endLine = len(contentLines)
	}

	var visibleLines []string
	if startLine < len(contentLines) {
		visibleLines = contentLines[startLine:endLine]
	}

	// Show scroll indicators
	var scrollIndicators []string
	if m.scrollOffset > 0 {
		scrollIndicators = append(scrollIndicators,
			lipgloss.NewStyle().Foreground(MutedColor).Italic(true).Render("↑ More content above"))
	}

	scrollIndicators = append(scrollIndicators, visibleLines...)

	if endLine < len(contentLines) {
		scrollIndicators = append(scrollIndicators,
			lipgloss.NewStyle().Foreground(MutedColor).Italic(true).Render("↓ More content below"))
	}

	// Fixed width content frame to prevent dynamic resizing
	fixedFrameWidth := 80
	if m.width > 0 && m.width-10 < fixedFrameWidth {
		fixedFrameWidth = m.width - 10 // Adjust for smaller terminals
	}

	briefContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(1).
		Width(fixedFrameWidth).
		Height(contentHeight).
		Render(strings.Join(scrollIndicators, "\n"))

	// Status bar with consistent ordering
	var statusBarHelp map[string]string
	if len(contentLines) > contentHeight {
		statusBarHelp = map[string]string{
			"↑↓":  "Scroll content",
			"e":   "Export to clipboard",
			"n":   "New Brief",
			"h":   "View History",
			"Esc": "Back",
		}
	} else {
		statusBarHelp = map[string]string{
			"e":   "Export to clipboard",
			"n":   "New Brief",
			"h":   "View History",
			"Esc": "Back",
		}
	}

	statusBar := RenderStatusBar(statusBarHelp, m.credits, m.config != nil && m.config.APIKey != "")

	// Build content with optional notification
	content := []string{
		title,
		"",
		keyword,
		credits,
		"",
		briefTitle,
		briefContent,
	}

	// Add notification if present
	if m.notification != nil {
		content = append(content, "", RenderNotification(*m.notification))
	}

	content = append(content, "", statusBar)

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

// renderError shows generation errors
func (m *ContentBriefGenerationModel) renderError() string {
	title := TitleStyle.Render("📄 Brief Generation Failed")

	errorText := lipgloss.NewStyle().
		Foreground(ErrorColor).
		Render(fmt.Sprintf("❌ Error: %s", m.error))

	instructions := ContentStyle.Render("Press 'r' to retry or 'n' for a new brief")

	// Status bar
	statusBar := RenderStatusBar(map[string]string{
		"r":   "Retry",
		"n":   "New Brief",
		"Esc": "Back",
	}, m.credits, m.config != nil && m.config.APIKey != "")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			errorText,
			"",
			instructions,
			"",
			statusBar,
		),
	)
}

// handleKeypress handles keyboard input
func (m *ContentBriefGenerationModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		if m.generating {
			// Cancel generation - for now just go back
			return m, func() tea.Msg { return BackToContentBriefMenuMsg{} }
		}
		return m, func() tea.Msg { return BackToContentBriefMenuMsg{} }
	}

	// Handle different states
	if m.inputFocused {
		return m.handleInputKeypress(msg)
	}

	if m.brief != "" {
		return m.handleBriefKeypress(msg)
	}

	if m.error != "" {
		return m.handleErrorKeypress(msg)
	}

	return m, nil
}

// handleInputKeypress handles input when in keyword entry mode
func (m *ContentBriefGenerationModel) handleInputKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if strings.TrimSpace(m.keywordInput) != "" {
			m.inputFocused = false
			return m, m.startGeneration()
		}
		return m, nil

	case "left":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "right":
		if m.cursor < len(m.keywordInput) {
			m.cursor++
		}
		return m, nil

	case "backspace":
		if m.cursor > 0 {
			m.keywordInput = m.keywordInput[:m.cursor-1] + m.keywordInput[m.cursor:]
			m.cursor--
		}
		return m, nil

	case "delete":
		if m.cursor < len(m.keywordInput) {
			m.keywordInput = m.keywordInput[:m.cursor] + m.keywordInput[m.cursor+1:]
		}
		return m, nil

	case "home":
		m.cursor = 0
		return m, nil

	case "end":
		m.cursor = len(m.keywordInput)
		return m, nil

	default:
		// Add regular characters
		if len(msg.String()) == 1 {
			char := msg.String()
			m.keywordInput = m.keywordInput[:m.cursor] + char + m.keywordInput[m.cursor:]
			m.cursor++
		}
		return m, nil
	}
}

// handleBriefKeypress handles input when showing the generated brief
func (m *ContentBriefGenerationModel) handleBriefKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n":
		// Start a new brief
		return m.resetForNewBrief(), nil

	case "h":
		// Go to history - we'll implement this later
		return m, func() tea.Msg { return BackToContentBriefMenuMsg{} }

	case "e":
		// Export content brief to clipboard
		if m.brief != "" {
			return m, ExportContentBriefToClipboardWithNotification(m.brief)
		}
		return m, nil

	case "up", "k":
		// Scroll up
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
		return m, nil

	case "down", "j":
		// Scroll down
		contentLines := strings.Split(m.brief, "\n")
		maxScroll := len(contentLines) - 10 // Leave some content visible
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.scrollOffset < maxScroll {
			m.scrollOffset++
		}
		return m, nil
	}

	return m, nil
}

// handleErrorKeypress handles input when showing errors
func (m *ContentBriefGenerationModel) handleErrorKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		// Retry generation
		m.error = ""
		return m, m.startGeneration()

	case "n":
		// New brief
		return m.resetForNewBrief(), nil
	}

	return m, nil
}

// Helper methods

// startGeneration initiates the brief generation process
func (m *ContentBriefGenerationModel) startGeneration() tea.Cmd {
	keyword := strings.TrimSpace(m.keywordInput)
	if keyword == "" {
		return nil
	}

	return func() tea.Msg {
		if m.config == nil || m.config.APIKey == "" {
			return BriefGenerationFailedMsg{
				Error: "API key not configured. Please go to Settings to set your API key.",
			}
		}

		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GenerateBrief(keyword)
		if err != nil {
			return BriefGenerationFailedMsg{
				Error: fmt.Sprintf("Failed to start brief generation: %v", err),
			}
		}

		return BriefGenerationStartedMsg{
			BriefID: resp.ID,
			Keyword: keyword,
		}
	}
}

// pollBriefStatus checks the status of the brief generation
func (m *ContentBriefGenerationModel) pollBriefStatus() tea.Cmd {
	if m.briefID == "" {
		return nil
	}

	m.lastPollTime = time.Now()

	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetBriefStatus(m.briefID)
		if err != nil {
			return BriefGenerationFailedMsg{
				Error: fmt.Sprintf("Failed to check brief status: %v", err),
			}
		}

		switch resp.Status {
		case "completed":
			if resp.Brief != nil {
				return BriefGenerationCompletedMsg{
					BriefID:     resp.ID,
					Brief:       *resp.Brief,
					Status:      resp.Status,
					CreditsUsed: resp.CreditsUsed,
				}
			} else {
				return BriefGenerationFailedMsg{
					Error: "Brief generation completed but no content was returned",
				}
			}
		case "failed":
			return BriefGenerationFailedMsg{
				Error: "Brief generation failed on the server",
			}
		default:
			return BriefGenerationProgressMsg{
				BriefID: resp.ID,
				Status:  resp.Status,
			}
		}
	})
}

// fetchCredits fetches the current credit balance from the API
func (m *ContentBriefGenerationModel) fetchCredits() tea.Cmd {
	if m.config == nil || m.config.APIKey == "" {
		return nil
	}

	return func() tea.Msg {
		client := api.NewClient(m.config.GetEffectiveBaseURL(), m.config.APIKey)
		resp, err := client.GetCreditBalance()
		if err != nil {
			return CreditsMsg{Credits: -1, Error: err}
		}
		return CreditsMsg{Credits: resp.Credits, Error: nil}
	}
}

// resetForNewBrief resets the model for a new brief generation
func (m *ContentBriefGenerationModel) resetForNewBrief() *ContentBriefGenerationModel {
	return &ContentBriefGenerationModel{
		width:         m.width,
		height:        m.height,
		config:        m.config,
		inputFocused:  true,
		cursor:        0,
		credits:       m.credits,
		scrollOffset:  0, // Reset scroll position
		spinner:       0, // Reset spinner
		spinnerFrames: m.spinnerFrames,
	}
}
