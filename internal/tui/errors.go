package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ugolbck/seofordev/internal/api"
)

// RenderCreditError renders the insufficient credits error with upgrade options
func RenderCreditError(err *api.InsufficientCreditsError) string {
	title := ErrorStatusStyle.Render("❌ Insufficient Credits")

	mainMessage := fmt.Sprintf(
		"You only have %d credits but need %d to complete this audit",
		err.CurrentBalance,
		err.CreditsRequired,
	)

	shortage := err.CreditsRequired - err.CurrentBalance
	shortageMsg := fmt.Sprintf("You need %d more credits to complete this audit", shortage)

	canProcessMsg := fmt.Sprintf("With your current balance, you can process %d pages", err.PagesThatCanBeProcessed)

	upgradeOptions := []string{
		"🚀 Upgrade to our lifetime deal at https://seofor.dev/pricing",
		"📉 Reduce scope with --max-pages flag",
	}

	optionsList := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(AccentColor).Bold(true).Render("Options:"),
		"",
	)

	for _, option := range upgradeOptions {
		optionsList = lipgloss.JoinVertical(lipgloss.Left,
			optionsList,
			lipgloss.NewStyle().Foreground(InfoColor).Render("• "+option),
		)
	}

	help := HelpStyle.Render("Press ctrl+c to exit and adjust your audit scope")

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			ContentStyle.Render(mainMessage),
			ContentStyle.Render(shortageMsg),
			ContentStyle.Render(canProcessMsg),
			"",
			optionsList,
			"",
			help,
		),
	)
}
