package filepicker

import (
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/styles"
)

const Help = "q -> cancel | Enter -> select file | Backspace -> parent dir"

type Model struct {
	Event      internal.Event
	filepicker filepicker.Model
	quitting   bool
	err        error
}

func (m Model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			m.Event = internal.Event{
				Status: internal.StatusCanceled,
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Check if the user has selected a file
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.quitting = true
		m.Event = internal.Event{
			Value:  path,
			Status: internal.StatusOk,
			Type:   internal.UploadFileRequest,
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	return styles.BoldStyle.Render("Select a file to upload:") + "\n" +
		m.filepicker.Styles.Directory.Render(
			m.filepicker.CurrentDirectory,
		) + "\n" +
		styles.BaseStyle.Render(m.filepicker.View()) + "\n" +
		styles.HelpStyle.Render(Help)
}

func NewModel() *Model {
	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.Getwd()
	fp.AutoHeight = false
	fp.Height = 20
	fp.ShowPermissions = true

	m := Model{
		filepicker: fp,
	}

	return &m
}

func RunModel() (internal.Event, error) {
	m := NewModel()
	p := tea.NewProgram(m)

	model, err := p.Run()
	return model.(Model).Event, err
}
