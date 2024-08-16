package styles

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var Theme = huh.ThemeCatppuccin()

const Divider = "--------------------------------------------"

var BaseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#EE6FF8"))

var (
	HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	BoldStyle = lipgloss.NewStyle().Bold(true).Foreground(Theme.Focused.NoteTitle.GetForeground())
	ErrStyle  = lipgloss.NewStyle().Foreground(Theme.Focused.ErrorMessage.GetForeground())
)

func PrintErrStr(errMsg string) {
	fmt.Println(ErrStyle.Render(errMsg))
}

func DestructiveTheme() *huh.Theme {
	t := Theme

	red := lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#ED567A"}

	t.Focused.Base = t.Focused.Base.BorderForeground(lipgloss.Color("238"))
	t.Focused.Title = t.Focused.Title.Foreground(red).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(red)

	return t
}
