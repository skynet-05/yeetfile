package styles

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var Theme = YeetFileTheme()

var (
	white       = lipgloss.Color("#ffffff")
	gray        = lipgloss.Color("#a6adc8")
	accent      = lipgloss.Color("#5a56e0")
	accentLight = lipgloss.Color("#8a86f0")
	shared      = lipgloss.Color("#3EB974")
	destructive = lipgloss.Color("#a83c3c")
)

func YeetFileTheme() *huh.Theme {
	t := huh.ThemeBase()

	//t.Focused.Base = t.Focused.Base.BorderForeground(subtext1)
	t.Focused.Title = t.Focused.Title.Foreground(white)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(white).Bold(true)
	t.Focused.Directory = t.Focused.Directory.Foreground(accentLight)
	t.Focused.Description = t.Focused.Description.Foreground(gray)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(destructive)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(destructive)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(accentLight).Bold(true)
	//t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(destructive)
	//t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(pink)
	t.Focused.Option = t.Focused.Option.PaddingLeft(1).PaddingRight(1).Foreground(gray)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(accentLight)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(white).PaddingLeft(1).PaddingRight(1)
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(accentLight)
	//t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(text)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.PaddingLeft(1).PaddingRight(1)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(white).Background(accent)
	//t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(text).Background(base)

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(white)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(gray)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(accentLight)

	t.Help = help.New().Styles

	// Blurred styles.
	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	return t
}

var (
	TableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(Theme.Focused.NoteTitle.GetForeground())
	HelpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	DirStyle     = lipgloss.NewStyle().Foreground(accentLight)
	SharedStyle  = lipgloss.NewStyle().Foreground(shared)
	BoldStyle    = lipgloss.NewStyle().Bold(true).Foreground(Theme.Focused.NoteTitle.GetForeground())
	TitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(Theme.Focused.FocusedButton.GetForeground())
	ErrStyle     = lipgloss.NewStyle().Foreground(Theme.Focused.ErrorMessage.GetForeground())
	SuccessStyle = lipgloss.NewStyle().Foreground(shared)
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
