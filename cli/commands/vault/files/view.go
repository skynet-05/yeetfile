package files

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/models"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
)

type Model struct {
	IncomingEvent internal.Event
	ViewRequest   internal.ViewRequest
	Context       *VaultContext

	init     bool
	quitting bool
	table    table.Model
	spinner  spinner.Model
	progress progress.Model
}

type Status struct {
	Loading    bool
	NewFolder  bool
	Processing bool
	Message    string
	Success    string
	Err        error
	Progress   int
	Total      int
}

const Help = `
Enter -> select/open | x -----> delete | s --> share | u ---> upload |
Backspace ----> back | n -> new folder | r -> rename | d -> download |`

var rows []table.Row
var items []models.VaultItem
var status Status

var folderViews = []string{""}
var folderPath = "/"

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.table.SetRows(rows)
	if m.IncomingEvent.Status == internal.StatusOk {
		switch m.IncomingEvent.Type {
		case internal.UploadFileRequest:
			m.upload(m.IncomingEvent)
		case internal.DeleteFileRequest:
			m.delete(m.IncomingEvent)
		case internal.NewFolderRequest:
			m.createFolder(m.IncomingEvent)
		case internal.RenameRequest:
			m.rename(m.IncomingEvent)
		case internal.ShareRequest:
			m.share(m.IncomingEvent)
		}

		m.IncomingEvent = internal.Event{}
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if status.Processing {
			var spinnerCmd tea.Cmd
			m.spinner, spinnerCmd = m.spinner.Update(msg)
			cmds = append(cmds, spinnerCmd)
			cmds = append(cmds, m.spinner.Tick)
		}
	case tea.KeyMsg:
		status.Err = nil
		status.Success = ""
		switch msg.String() {
		case "backspace":
			if len(folderViews) > 1 {
				status.Loading = true
				folderID := folderViews[len(folderViews)-2]
				newModel, err := NewModel(folderID)
				if err == nil {
					splitPath := strings.Split(folderPath, "/")
					folderPath = strings.Join(splitPath[:len(splitPath)-2], "/")
					if len(folderPath) == 0 {
						folderPath = "/"
					} else {
						folderPath += "/"
					}
					folderViews = folderViews[:len(folderViews)-1]
				}
				status.Loading = false
				return newModel, nil
			}
		case "enter":
			if len(items) == 0 {
				return m, nil
			}

			item := items[m.table.Cursor()]
			// Open folder or prompt for download
			if item.IsFolder {
				status.Loading = true
				newModel, err := NewModel(item.RefID)
				if err == nil {
					folderViews = append(folderViews, item.RefID)
					folderPath += item.Name + "/"
				}
				status.Loading = false

				return newModel, nil
			}
		case "q", "ctrl+c": // Exit
			m.quitting = true
			return m, tea.Quit
		case "n": // New folder
			return m.NewFolderRequest()
		case "d": // Download
			if len(items) == 0 {
				return m, nil
			}

			item := items[m.table.Cursor()]
			if item.IsFolder {
				status.Err = errors.New("folder download is not currently supported")
				return m, nil
			}

			m.download(item)
			return m, m.spinner.Tick
		case "x", "r", "s": // Modify file (delete, rename, share, etc)
			if len(items) == 0 {
				return m, nil
			}

			item := items[m.table.Cursor()]
			if !item.CanModify {
				status.Err = errors.New("you are not allowed to modify this file")
				return m, nil
			} else if !item.IsOwner && msg.String() == "s" {
				status.Err = errors.New("you cannot share content you do not own")
				return m, nil
			}

			switch msg.String() {
			case "r":
				return m.NewRenameRequest(item)
			case "x":
				return m.NewDeleteRequest(item)
			case "s":
				return m.NewShareRequest(item)
			}
		case "u": // Upload file
			return m.NewUploadRequest()
		}
	}

	if !status.Processing && !status.NewFolder {
		m.table, cmd = m.table.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting || m.ViewRequest.View > internal.NullView {
		return ""
	}

	vaultView := styles.BaseStyle.Render(m.table.View())

	if status.Err != nil {
		errMsg := status.Err.Error()
		vaultView += "\n✗ Error: " + errMsg
	} else if len(status.Success) > 0 {
		vaultView += "\n" + status.Success
	} else if status.Loading {
		vaultView += "Loading..."
	} else if status.Processing && status.Total == 0 {
		vaultView += "\n" + m.spinner.View() + " " + status.Message
	} else if status.Processing && status.Total > 0 {
		percent := float64(status.Progress) / float64(status.Total)
		progressStr := "\n " + status.Message + " " +
			m.progress.ViewAs(percent) + "\n"
		vaultView += progressStr
	} else {
		progressStr := "\n Storage: " +
			m.progress.ViewAs(.56) + " full │ 3.5KB / 100GB \n"
		vaultView += progressStr
	}

	return styles.BoldStyle.Render("YeetFile Vault > home"+folderPath) + "\n" +
		vaultView + "\n" +
		styles.HelpStyle.Render(Help)
}

func (m Model) upload(event internal.Event) {
	fullPath := strings.Split(event.Value, string(os.PathSeparator))
	fileName := fullPath[len(fullPath)-1]
	status.Processing = true
	status.Message = fmt.Sprintf("Uploading %s...", fileName)

	go func() {
		err := m.Context.UploadFile(event.Value, func(current int, total int) {
			status.Progress = current
			status.Total = total
		})
		m.finishUpdates(err, true)
	}()
}

func (m Model) delete(event internal.Event) {
	status.Processing = true
	status.Message = fmt.Sprintf("Deleting '%s'...", event.Item.Name)

	go func() {
		err := m.Context.Delete(event.Item)
		m.finishUpdates(err, true)
	}()
}

func (m Model) rename(event internal.Event) {
	status.Processing = true
	status.Message = fmt.Sprintf(
		"Renaming '%s' to '%s'...", event.Item.Name, event.Value)

	go func() {
		err := m.Context.Rename(event.Value, event.Item)
		m.finishUpdates(err, true)
	}()
}

func (m Model) share(event internal.Event) {
	m.Context.Update(event.Item)
	m.finishUpdates(nil, true)
}

func (m Model) download(item models.VaultItem) {
	downloadStr := fmt.Sprintf("Downloading '%s'...", item.Name)
	status.Processing = true
	status.Message = downloadStr

	go func() {
		filename, err := m.Context.Download(item, func(c int, max int) {
			progressPercent := int((float32(c) / float32(max)) * 100)
			status.Message = fmt.Sprintf(
				"%s (%d%%)",
				downloadStr, progressPercent)
		})

		m.finishUpdates(err, false)
		if err == nil {
			fileStr := fmt.Sprintf(".%c%s", os.PathSeparator, filename)
			status.Success = fmt.Sprintf(
				"File downloaded: %s",
				fileStr)
		}
	}()
}

func (m Model) createFolder(event internal.Event) {
	status.Processing = true
	status.Message = fmt.Sprintf("Creating folder '%s'...", event.Value)

	go func() {
		err := m.Context.CreateFolder(event.Value)

		status = Status{}
		if err != nil {
			status.Err = err
		} else {
			items = m.Context.Content
			rows = CreateItemRows(items)
		}
	}()
}

func (m Model) finishUpdates(err error, updateItems bool) {
	status = Status{}
	if err != nil {
		status.Err = err
	} else if updateItems {
		items = m.Context.Content
		rows = CreateItemRows(items)
	}
}

func (m Model) NewDeleteRequest(item models.VaultItem) (tea.Model, tea.Cmd) {
	m.ViewRequest = internal.ViewRequest{
		View: internal.ConfirmationView,
		Type: internal.DeleteFileRequest,
		Item: item,
	}

	return m, tea.Quit
}

func (m Model) NewShareRequest(item models.VaultItem) (tea.Model, tea.Cmd) {
	m.ViewRequest = internal.ViewRequest{
		View: internal.ShareView,
		Type: internal.ShareRequest,
		Item: item,
	}

	return m, tea.Quit
}

func (m Model) NewRenameRequest(item models.VaultItem) (tea.Model, tea.Cmd) {
	m.ViewRequest = internal.ViewRequest{
		View: internal.RenameView,
		Type: internal.RenameRequest,
		Item: item,
	}

	return m, tea.Quit
}

func (m Model) NewFolderRequest() (tea.Model, tea.Cmd) {
	m.ViewRequest = internal.ViewRequest{
		View: internal.NewFolderView,
		Type: internal.NewFolderRequest,
	}

	return m, tea.Quit
}

func (m Model) NewUploadRequest() (tea.Model, tea.Cmd) {
	m.ViewRequest = internal.ViewRequest{
		View: internal.FilePickerView,
		Type: internal.UploadFileRequest,
	}

	return m, tea.Quit
}

func NewModel(folderID string) (Model, error) {
	ctx, err := FetchVaultContext(folderID)
	if err == nil && len(ctx.Content) == 0 {
		items, err = ctx.parseContent()
	} else if err == nil {
		items = ctx.Content
	}

	status.Err = err
	rows = CreateItemRows(items)

	maxNameLen := 15
	maxDateLen := 20
	maxSharedLen := 6
	for _, row := range rows {
		maxNameLen = max(len(row[0]), maxNameLen)
		maxDateLen = max(len(row[2]), maxDateLen)
		maxSharedLen = max(len(row[3]), maxSharedLen)
	}

	columns := []table.Column{
		{Title: "Name", Width: maxNameLen},
		{Title: "Size", Width: 10},
		{Title: "Modified", Width: maxDateLen},
		{Title: "Shared", Width: maxSharedLen},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Cell = s.Cell.Foreground(lipgloss.Color("255"))
	s.Header = s.Header.
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("255")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Background(lipgloss.Color("#5A56E0"))
	s.Cell = s.Cell.Foreground(lipgloss.Color("#ffffff"))
	t.SetStyles(s)

	m := Model{
		Context:  ctx,
		table:    t,
		spinner:  spinner.New(),
		progress: progress.New(progress.WithDefaultScaledGradient()),
	}

	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	m.spinner.Spinner = spinner.Points

	m.init = true
	return m, err
}

func ShowVaultPasswordPromptModel(errorMsgs ...string) ([]byte, error) {
	var password string
	desc := "Enter your vault session password below to continue"
	if len(errorMsgs) > 0 {
		desc = errorMsgs[0]
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(utils.GenerateTitle(
				"Vault Session Password")).
				Description(desc),
			huh.NewInput().Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&password),
			huh.NewConfirm().Affirmative("Submit").Negative(""),
		),
	).WithTheme(styles.Theme).Run()

	return []byte(password), err
}

func RunFilesModel(m Model, event internal.Event) (Model, error) {
	if keyPair.PublicKey == nil || keyPair.PrivateKey == nil {
		var keyErr error
		keyPair, keyErr = unlockVaultKeys()
		if keyErr != nil {
			errMsg := fmt.Sprintf(
				"Error unlocking vault keys: %v\n",
				keyErr)
			styles.PrintErrStr(errMsg)
			os.Exit(1)
		}
	}

	if m.init == false {
		m, _ = NewModel("")
	}

	m.IncomingEvent = event
	m.ViewRequest = internal.ViewRequest{}

	p := tea.NewProgram(m)
	model, err := p.Run()
	return model.(Model), err
}

func unlockVaultKeys() (crypto.KeyPair, error) {
	var kp crypto.KeyPair

	cliKey := crypto.ReadCLIKey()
	encPrivateKey, publicKey, err := globals.Config.GetKeys()
	if err != nil {
		errMsg := fmt.Sprintf("Error reading key files: %v\n", err)
		styles.PrintErrStr(errMsg)
	}
	if privateKey, err := crypto.DecryptChunk(cliKey, encPrivateKey); err == nil {
		kp = crypto.IngestKeys(privateKey, publicKey)
	} else {
		var cliKeyFunc func(errMsgs ...string) error
		cliKeyFunc = func(errMsgs ...string) error {
			cliPassword, err := ShowVaultPasswordPromptModel(errMsgs...)
			if err != nil {
				return err
			}
			key := crypto.DerivePBKDFKey(cliPassword, cliKey)
			if privateKey, err := crypto.DecryptChunk(key, encPrivateKey); err == nil {
				kp = crypto.IngestKeys(privateKey, publicKey)
				return nil
			} else {
				return cliKeyFunc("Incorrect password")
			}
		}

		err = cliKeyFunc()
		if err != nil {
			return crypto.KeyPair{}, err
		}
	}

	return kp, nil
}
