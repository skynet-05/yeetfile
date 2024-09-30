package files

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	huhSpinner "github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
	"unicode"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/models"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

type Model struct {
	IncomingEvent internal.Event
	ViewRequest   internal.ViewRequest
	Context       *VaultContext

	filteredItems []models.VaultItem

	filtering bool
	filterStr string
	init      bool
	quitting  bool
	table     table.Model
	spinner   spinner.Model
	progress  progress.Model
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

type Storage struct {
	available int64
	used      int64
}

const Help = `
Enter -> select/open | x -----> delete | s --> share | u ---> upload |
Backspace ----> back | n -> new folder | r -> rename | d -> download |`

const FilterHelp = `
Enter -> select/open | escape -> exit filter`

var status Status
var storage Storage

var rows []table.Row
var items []models.VaultItem
var folderViews = []string{""}
var folderPath = "/"

var height = 12

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if m.IncomingEvent.Status == internal.StatusOk {
		m.filtering = false
		m.filterStr = ""
		rows = CreateItemRows(items)
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

	m.table.SetRows(rows)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		height = msg.Height - 20
		m.table.SetHeight(height)
		return m, nil

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
		if m.filtering {
			return m.filterItems(msg.String())
		}

		switch msg.String() {
		case "/":
			m.filtering = true
			return m, nil
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
				newModel.table.SetHeight(height)
				return newModel, nil
			}
		case "q", "ctrl+c", "esc": // Exit
			m.quitting = true
			return m, tea.Quit
		case "n": // New folder
			return m.NewFolderRequest()
		case "enter", "d", "x", "r", "s":
			if len(items) == 0 {
				return m, nil
			}

			var item models.VaultItem
			if len(m.filterStr) > 0 {
				item = m.filteredItems[m.table.Cursor()]
			} else {
				item = items[m.table.Cursor()]
			}

			switch msg.String() {
			case "enter": // Open file
				// Open folder or view file
				if item.IsFolder {
					status.Loading = true
					newModel, err := NewModel(item.RefID)
					if err == nil {
						folderViews = append(folderViews, item.RefID)
						folderPath += item.Name + "/"
					}
					status.Loading = false

					return newModel, nil
				} else {
					// Enter file view
					return m.NewFileViewRequest(item)
				}
			case "d": // Download file
				if item.IsFolder {
					status.Err = errors.New("folder download is not currently supported")
					return m, nil
				}

				m.download(item)
				return m, m.spinner.Tick
			case "x", "r", "s": // Modify file
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
			}
		case "u": // Upload file
			return m.NewUploadRequest()
		}
	}

	if !status.Processing {
		m.table, cmd = m.table.Update(msg)
	}

	if m.table.Cursor() < 0 {
		m.table.SetCursor(0)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) filterItems(c string) (tea.Model, tea.Cmd) {
	needsFilter := false
	switch c {
	case "backspace":
		if len(m.filterStr) > 0 {
			m.filterStr = m.filterStr[:len(m.filterStr)-1]
			needsFilter = true
		} else {
			m.filtering = false
		}
	case "ctrl+c", "esc":
		rows = CreateItemRows(items)
		m.filtering = false
	case "enter":
		m.filtering = false
	case "tab", "left", "right":
		// Ignore
	case "up", "down":
		cursor := m.table.Cursor()
		if c == "up" {
			m.table.SetCursor(cursor - 1)
		} else {
			m.table.SetCursor(cursor + 1)
		}
	default:
		if len(m.filterStr) > 0 && len(m.filteredItems) == 0 {
			// Don't continue filtering if there's nothing to filter
			break
		}
		m.filterStr += c
		needsFilter = true
	}

	if needsFilter {
		hasUpper := false
		for _, c := range m.filterStr {
			if unicode.IsUpper(c) {
				hasUpper = true
				break
			}
		}

		var filteredItems []models.VaultItem
		for _, item := range items {
			var name string
			if hasUpper {
				name = item.Name
			} else {
				name = strings.ToLower(item.Name)
			}

			if strings.Contains(name, m.filterStr) {
				filteredItems = append(filteredItems, item)
			}
		}

		rows = CreateItemRows(filteredItems)
		m.filteredItems = filteredItems
		m.table.SetCursor(0)
	}

	return m.Update(nil)
}

func (m Model) View() string {
	if m.quitting || m.ViewRequest.View > internal.NullView {
		return ""
	}
	m.table.SetHeight(height)
	return m.tableViewer()
}

func (m Model) tableViewer() string {
	vaultView := styles.TableStyle.Render(m.table.View())

	if status.Err != nil {
		errMsg := status.Err.Error()
		vaultView += "\n✗ Error: " + errMsg
	} else if m.filtering || len(m.filterStr) > 0 {
		vaultView += "\nFilter: " + m.filterStr
	} else if len(status.Success) > 0 {
		vaultView += "\n " + status.Success
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
		percentage := float64(storage.used) / float64(storage.available)
		storageUsed := shared.ReadableFileSize(storage.used)
		storageAvailable := shared.ReadableFileSize(storage.available)
		progressStr := fmt.Sprintf(
			"\n Storage: %s full | %s / %s \n",
			m.progress.ViewAs(percentage),
			storageUsed,
			storageAvailable)
		vaultView += progressStr
	}

	var helpStr string
	if m.filtering {
		helpStr = FilterHelp
	} else {
		helpStr = Help
	}

	return utils.GenerateTitle("Vault > home"+folderPath) + "\n" +
		vaultView + "\n" +
		styles.HelpStyle.Render(helpStr)
}

func (m Model) upload(event internal.Event) {
	fullPath := strings.Split(event.Value, string(os.PathSeparator))
	fileName := fullPath[len(fullPath)-1]
	status.Processing = true
	status.Message = fmt.Sprintf("Uploading %s...", fileName)

	go func() {
		size, err := m.Context.UploadFile(event.Value, func(current int, total int) {
			status.Progress = current
			status.Total = total
		})
		m.finishUpdates(err, true)
		if err == nil {
			msg := fmt.Sprintf("Successfully uploaded %s!", fileName)
			status.Success = styles.SuccessStyle.Render(msg)
			storage.used += size
		}
	}()
}

func (m Model) delete(event internal.Event) {
	status.Processing = true
	status.Message = fmt.Sprintf("Deleting %s...", event.Item.Name)

	go func() {
		err := m.Context.Delete(event.Item)
		m.finishUpdates(err, true)
		if err == nil {
			storage.used -= event.Item.Size - int64(constants.TotalOverhead)
			if storage.used < 0 {
				storage.used = 0
			}

			msg := fmt.Sprintf("Deleted %s!", event.Item.Name)
			status.Success = styles.SuccessStyle.Render(msg)
		}
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

func (m Model) NewFileViewRequest(item models.VaultItem) (tea.Model, tea.Cmd) {
	m.ViewRequest = internal.ViewRequest{
		View:      internal.FileViewerView,
		Type:      internal.ViewFileRequest,
		Item:      item,
		CryptoCtx: m.Context.Crypto,
	}

	return m, tea.Quit
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

	minNameLen := 20
	maxNameLen := 30
	maxDateLen := 12
	maxSharedLen := 6
	for _, row := range rows {
		maxNameLen = max(len(row[0]), maxNameLen)
		maxDateLen = max(len(row[2]), maxDateLen)
		maxSharedLen = max(len(row[3]), maxSharedLen)
	}

	maxNameLen = max(maxNameLen, minNameLen)
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
		table.WithHeight(12),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(styles.Theme.Focused.FocusedButton.GetForeground()).
		Background(styles.Theme.Focused.FocusedButton.GetBackground()).
		Bold(false)
	t.SetStyles(s)

	m := Model{
		Context:  ctx,
		table:    t,
		spinner:  spinner.New(),
		progress: progress.New(progress.WithScaledGradient("#5A56E0", "#8A86F0")),
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
				"Error decrypting vault keys: %v\n",
				keyErr)
			styles.PrintErrStr(errMsg)
			os.Exit(1)
		}
	}

	if m.init == false {
		_ = huhSpinner.New().Title("Loading vault...").Action(func() {
			m, _ = NewModel("")
			usage, err := globals.API.GetAccountUsage()
			if err != nil {
				errMsg := fmt.Sprintf(
					"Error fetching account usage values: %v\n",
					err)
				styles.PrintErrStr(errMsg)
				os.Exit(1)
			}

			storage.available = usage.StorageAvailable
			storage.used = usage.StorageUsed
		}).Run()
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
