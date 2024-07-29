package auth

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"yeetfile/cli/commands/auth/login"
	"yeetfile/cli/commands/auth/signup"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
)

const LoginAction = "Log In"
const SignUpAction = "Sign Up"

type Model struct {
	form *huh.Form
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// Quit when the form is done.
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.form.View()
}

func ShowAuthModel() {
	var action string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(utils.GenerateTitle("Authentication")),
			huh.NewSelect[string]().
				Options(huh.NewOptions(
					SignUpAction,
					LoginAction,
					"Cancel")...,
				).Value(&action),
		),
	).WithTheme(styles.Theme).WithShowHelp(true).Run()
	utils.HandleCLIError("", err)

	if action == SignUpAction {
		signup.ShowSignupModel()
		login.ShowLoginModel()
	} else if action == LoginAction {
		login.ShowLoginModel()
	}
}
