package styles

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var Theme = YeetFileTheme()

var (
	Black       = lipgloss.Color("0")
	White       = lipgloss.Color("7")
	BrightWhite = lipgloss.Color("15")
	Gray        = lipgloss.Color("8")
	Magenta     = lipgloss.Color("5")
	Accent      = lipgloss.Color("4")
	AccentLight = lipgloss.Color("12")
	Green       = lipgloss.Color("2")
	Red         = lipgloss.Color("1")
)

func YeetFileTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Base = t.Focused.Base.Foreground(White)
	t.Focused.Title = t.Focused.Title.Foreground(White)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(White).Bold(true)
	t.Focused.Directory = t.Focused.Directory.Foreground(AccentLight)
	t.Focused.Description = t.Focused.Description.Foreground(Gray)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(Red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(Red)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(AccentLight).Bold(true)
	//t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(destructive)
	//t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(pink)
	t.Focused.Option = t.Focused.Option.PaddingLeft(1).PaddingRight(1).Foreground(Gray)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(AccentLight)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(White).PaddingLeft(1).PaddingRight(1)
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(AccentLight)
	//t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(text)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.PaddingLeft(1).PaddingRight(1)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(Black).Background(White)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(Black).Background(Gray)

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(White)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(Gray)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(AccentLight)
	t.Focused.TextInput.Text = t.Focused.TextInput.Text.Foreground(White)

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
	DirStyle     = lipgloss.NewStyle().Foreground(AccentLight)
	SharedStyle  = lipgloss.NewStyle().Foreground(Green)
	BoldStyle    = lipgloss.NewStyle().Bold(true).Foreground(Theme.Focused.NoteTitle.GetForeground())
	TitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(White)
	ErrStyle     = lipgloss.NewStyle().Foreground(Red)
	SuccessStyle = lipgloss.NewStyle().Foreground(Green)
)

func PrintErrStr(errMsg string) {
	fmt.Println(ErrStyle.Render(errMsg))
}

func DestructiveTheme() *huh.Theme {
	t := Theme

	t.Focused.Base = t.Focused.Base.BorderForeground(Red)
	t.Focused.Title = t.Focused.Title.Foreground(Red).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(Red)

	return t
}
