package share

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"strings"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/crypto"
	"yeetfile/cli/models"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

type model struct {
	item        models.VaultItem
	users       []shared.ShareInfo
	input       string
	errMsg      string
	decryptFunc crypto.CryptFunc
	decryptKey  []byte
}

var actionMap map[Action]func(model) (internal.Event, error)

func (m model) add() (internal.Event, error) {
	recipient := m.input
	var confirmed bool
	var perm Perm
	fields := []huh.Field{
		huh.NewInput().
			Title("Share With User").
			Description("Enter user's email or account ID below").
			Placeholder("user@example.com | 1234123412341234").
			Value(&recipient),
		huh.NewSelect[Perm]().
			Value(&perm).
			Options(
				huh.NewOption(ReadPerm, Read),
				huh.NewOption(WritePerm, Write),
			).
			Title("Permissions"),
		huh.NewConfirm().
			Affirmative("Share").
			Negative("Cancel").
			Value(&confirmed),
	}

	if len(m.errMsg) > 0 {
		fields = append(fields, huh.NewNote().
			Title(styles.ErrStyle.Render("Error:")).
			Description(styles.ErrStyle.Render(m.errMsg)))
	}
	_ = huh.NewForm(huh.NewGroup(fields...)).WithTheme(styles.Theme).Run()

	m.input = recipient
	if confirmed {
		addedUser, err := shareItem(
			m.item,
			m.decryptFunc,
			m.decryptKey,
			recipient,
			perm)
		if err != nil {
			m.errMsg = err.Error()
			return m.add()
		}

		m.users = append(m.users, addedUser)
	}

	return RunModel(m.item, m.users, m.decryptFunc, m.decryptKey)
}

func (m model) edit() (internal.Event, error) {
	var confirmed bool

	confirm := huh.NewConfirm().
		Affirmative("Confirm").
		Negative("Cancel").
		Value(&confirmed)
	if len(m.errMsg) > 0 {
		confirm.Description(styles.ErrStyle.Render(m.errMsg))
	}

	formTitle := fmt.Sprintf("Edit Permissions for '%s'", m.item.Name)
	fields := []huh.Field{huh.NewNote().Title(formTitle)}
	var shareFields []huh.Field

	editList := make([]shared.ShareInfo, len(m.users))
	copy(editList, m.users)

	for i := range editList {
		share := &editList[i]
		title := fmt.Sprintf("%d. %s", i+1, share.Recipient)

		shareField := huh.NewSelect[bool]().Options([]huh.Option[bool]{
			huh.NewOption(ReadPerm, false),
			huh.NewOption(WritePerm, true),
		}...).Title(title).Value(&share.CanModify)

		shareFields = append(shareFields, shareField)
	}

	fields = append(fields, shareFields...)
	fields = append(fields, confirm)
	_ = huh.NewForm(huh.NewGroup(fields...)).WithTheme(styles.Theme).Run()

	var err error
	var updated []shared.ShareInfo
	_ = spinner.New().Title("Updating access...").
		Action(func() {
			updated, err = editPermissions(m.item, editList)
		}).Run()
	if err != nil {
		m.errMsg = err.Error()
		for i := range m.users {
			share := &m.users[i]
			for _, update := range updated {
				if share.ID == update.ID {
					share.CanModify = update.CanModify
					break
				}
			}
		}
		return m.edit()
	} else {
		m.users = updated
	}

	return RunModel(m.item, m.users, m.decryptFunc, m.decryptKey)
}

func (m model) remove() (internal.Event, error) {
	var confirmed bool
	var toRemove []shared.ShareInfo
	var options []huh.Option[shared.ShareInfo]
	for _, user := range m.users {
		options = append(options, huh.NewOption(user.Recipient, user))
	}

	confirm := huh.NewConfirm().
		Affirmative("Confirm").
		Negative("Cancel").
		Value(&confirmed)
	if len(m.errMsg) > 0 {
		confirm.Description(styles.ErrStyle.Render(m.errMsg))
	}

	_ = huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[shared.ShareInfo]().
			Title("Remove access for the following user(s):").
			Description("Press 'x' to select").
			Options(options...).Value(&toRemove), confirm)).Run()

	if confirmed {
		var err error
		var removed []shared.ShareInfo
		title := fmt.Sprintf("Removing access for %d user(s)...", len(toRemove))
		_ = spinner.New().Title(title).
			Action(func() {
				removed, err = removeAccess(m.item, toRemove)
			}).Run()

		m.users = shared.RemoveOverlap(m.users, removed)
		if err != nil {
			m.errMsg = err.Error()
			return m.remove()
		}
	}

	return RunModel(m.item, m.users, m.decryptFunc, m.decryptKey)
}

func (m model) cancel() (internal.Event, error) {
	m.item.SharedWith = len(m.users)
	return internal.Event{
		Status: internal.StatusOk,
		Type:   internal.ShareRequest,
		Item:   m.item,
	}, nil
}

func generateFormFields(
	label string,
	shares []shared.ShareInfo,
	action *Action,
) []huh.Field {
	var fields []huh.Field
	if len(shares) == 0 {
		fields = append(fields, huh.NewNote().
			Title(label).
			Description("No shared users!"))

		fields = append(fields, huh.NewSelect[Action]().
			Value(action).
			Options(
				huh.NewOption("Add User", Add),
				huh.NewOption("Return to Vault", Cancel),
			).
			Title("Select an action to perform"))
	} else {
		divider := huh.NewNote().Title("------------------------------")
		fields = append(fields, huh.NewNote().
			Title(label).
			Description(fmt.Sprintf("%d shared users", len(shares))))
		fields = append(fields, divider)
		for i, share := range shares {
			var opt string
			if share.CanModify {
				opt = WritePerm
			} else {
				opt = ReadPerm
			}

			idx := fmt.Sprintf("%d. ", i+1)
			title := fmt.Sprintf("%s%s", idx, share.Recipient)
			desc := strings.Repeat(" ", len(idx)) + opt
			field := huh.NewNote().Title(title).Description(desc)
			fields = append(fields, field)
		}
		fields = append(fields, divider)

		fields = append(fields, huh.NewSelect[Action]().
			Value(action).
			Options(
				huh.NewOption("Add User", Add),
				huh.NewOption("Edit Permissions", Edit),
				huh.NewOption("Remove Access", Remove),
				huh.NewOption("Return to vault", Cancel),
			).
			Title("Select an action to perform"))
	}

	return fields
}

func RunModel(
	item models.VaultItem,
	users []shared.ShareInfo,
	decryptFunc crypto.CryptFunc,
	decryptKey []byte,
) (internal.Event, error) {
	var sharedItemUsers []shared.ShareInfo
	if users == nil {
		_ = spinner.New().Title("Fetching shared info...").
			Action(func() {
				sharedItemUsers, _ = fetchSharedInfo(item)

			}).Run()
	} else {
		sharedItemUsers = users
	}

	m := model{
		item:        item,
		users:       sharedItemUsers,
		decryptFunc: decryptFunc,
		decryptKey:  decryptKey,
	}

	var title string
	var label string
	if item.IsFolder {
		title = "Share Folder"
		label = fmt.Sprintf("> Folder: %s", item.Name)
	} else {
		title = "Share File"
		label = fmt.Sprintf("> File: %s", item.Name)
	}

	header := huh.NewNote().Title(utils.GenerateTitle(title))

	var action Action
	fields := generateFormFields(label, sharedItemUsers, &action)
	fields = append([]huh.Field{header}, fields...)

	_ = huh.NewForm(huh.NewGroup(fields...)).WithTheme(styles.Theme).Run()
	return actionMap[action](m)
}

func init() {
	actionMap = map[Action]func(model) (internal.Event, error){
		Add:    model.add,
		Remove: model.remove,
		Edit:   model.edit,
		Cancel: model.cancel,
	}
}
