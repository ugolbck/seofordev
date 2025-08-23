package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Placeholder models for features not yet implemented

// NewHelpModel creates a placeholder help interface
func NewHelpModel() tea.Model {
	return &stubModel{
		title:   "❓ Help & Documentation",
		content: "Learn how to use SEO Developer Tools effectively",
	}
}


// stubModel is a simple placeholder for unimplemented features
type stubModel struct {
	title   string
	content string
	width   int
	height  int
}

func (m *stubModel) Init() tea.Cmd {
	return nil
}

func (m *stubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	return m, nil
}

func (m *stubModel) View() string {
	help := RenderKeyHelp(map[string]string{
		"Enter": "Back to menu",
		"Esc":   "Back to menu",
		"q":     "Quit",
	})

	return AppStyle.Render(
		TitleStyle.Render(m.title) + "\n\n" +
			ContentStyle.Render(m.content+"\n\n(Coming soon!)") + "\n\n" +
			help,
	)
}

// contentBriefStubModel is a placeholder for content brief features that returns to content brief menu
type contentBriefStubModel struct {
	title   string
	content string
	width   int
	height  int
}

func (m *contentBriefStubModel) Init() tea.Cmd {
	return nil
}

func (m *contentBriefStubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			return m, func() tea.Msg { return BackToContentBriefMenuMsg{} }
		}
	}

	return m, nil
}

func (m *contentBriefStubModel) View() string {
	help := RenderKeyHelp(map[string]string{
		"Enter": "Back to Content Brief menu",
		"Esc":   "Back to Content Brief menu",  
		"q":     "Quit",
	})

	return AppStyle.Render(
		TitleStyle.Render(m.title) + "\n\n" +
			ContentStyle.Render(m.content+"\n\n(Coming soon!)") + "\n\n" +
			help,
	)
}
