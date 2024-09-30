package filepicker

import (
	"fmt"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/utils"
)

var docStyle = lipgloss.NewStyle().Margin(0, 0)
var fpStyle = filepicker.New().Styles
var seenBackspaceUp bool

type item struct {
	name  string
	size  string
	perm  string
	date  string
	isDir bool
}

type Model struct {
	Event    internal.Event
	list     list.Model
	quitting bool
	err      error

	selected    item
	currentDir  string
	gotoTopNext bool
}

var cursorPos map[string]int

func (i item) Title() string {
	if i.isDir {
		return i.name + "/"
	}

	return i.name
}

func (i item) Description() string {
	itemType := fpStyle.DisabledFile.Render("File")
	if i.isDir {
		itemType = fpStyle.Directory.Render("Directory")
		return fmt.Sprintf("└─ %s", itemType)
	}

	return fmt.Sprintf("└─ %s", i.size)
}

func (i item) FilterValue() string { return i.name }

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() != "g" || m.list.FilterState() == list.Filtering {
			m.gotoTopNext = false
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if m.list.FilterState() != list.Filtering {
				m.quitting = true
				m.Event = internal.Event{
					Status: internal.StatusCanceled,
				}
				return m, tea.Quit
			}
		case "escape":
			m.list.ResetFilter()
			m.list.ShowTitle()
			return m, nil
		case "J", "K":
			if m.list.FilterState() == list.Filtering {
				break
			}

			switch msg.String() {
			case "J":
				newIdx := min(len(m.list.Items())-1, m.list.Index()+5)
				m.list.Select(newIdx)
			case "K":
				newIdx := max(0, m.list.Index()-5)
				m.list.Select(newIdx)
			}
		case "backspace":
			if m.list.FilterState() == list.Unfiltered {
				m.currentDir = goUpDir(m.currentDir)
				items, err := getItemsFromDir(m.currentDir)
				if err != nil {
					log.Println(err)
				}

				m.list.SetItems(items)
				m.list.Select(0)

				idx, ok := cursorPos[m.currentDir]
				if ok {
					m.list.Select(idx)
				}

				return m.showNewDirStatus(), nil
			}
		case "enter":
			if m.list.FilterState() == list.Filtering {
				return m, func() tea.Msg {
					return tea.KeyMsg{Type: tea.KeyDown}
				}
			}

			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.selected = i
			}

			if m.list.FilterState() == list.FilterApplied {
				m.list.ResetFilter()
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	cursorPos[m.currentDir] = m.list.Index()

	// Check if the user has selected a file
	if m.selected != (item{}) {
		if m.selected.isDir {
			m.currentDir = appendDir(m.currentDir, m.selected.name)
			items, err := getItemsFromDir(m.currentDir)
			if err != nil {
				log.Println(err)
			}

			m.list.SetItems(items)
			m.list.Select(0)
			m.selected = item{}

			idx, ok := cursorPos[m.currentDir]
			if ok {
				m.list.Select(idx)
			}

			return m.showNewDirStatus(), nil
		} else {
			m.quitting = true
			m.Event = internal.Event{
				Value:  getItemPath(m.currentDir, m.selected.name),
				Status: internal.StatusOk,
				Type:   internal.UploadFileRequest,
			}
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	return fpStyle.Directory.Render(m.currentDir) + "\n" +
		docStyle.Render(m.list.View())
}

func (m Model) showNewDirStatus() Model {
	var backspaceNote string
	if !seenBackspaceUp {
		backspaceNote = "\n[Backspace -> Navigate Up]"
		seenBackspaceUp = true
	}

	singularName := fmt.Sprintf("\r %s%s", m.currentDir, backspaceNote)
	pluralName := fmt.Sprintf("\r %s%s", m.currentDir, backspaceNote)
	m.list.SetStatusBarItemName(singularName, pluralName)
	return m
}

func NewModel() *Model {
	m := Model{}

	currentDir, _ := os.Getwd()
	items, err := getItemsFromDir(currentDir)
	if err != nil {
		m.err = err
	}

	listDelegate := list.NewDefaultDelegate()
	listDelegate.SetSpacing(0)
	m.currentDir = currentDir
	m.list = list.New(items, listDelegate, 0, 0)
	m.list.Styles.Title = lipgloss.NewStyle()
	m.list.Title = utils.GenerateTitle("Upload")
	m.list.SetStatusBarItemName("ball", "balls")

	return &m
}

func RunModel() (internal.Event, error) {
	m := NewModel().showNewDirStatus()
	p := tea.NewProgram(m)

	model, err := p.Run()
	return model.(Model).Event, err
}

func init() {
	cursorPos = make(map[string]int)
}
