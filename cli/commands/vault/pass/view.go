package pass

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"slices"
	"strconv"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/globals"
	"yeetfile/cli/models"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

var title = utils.GenerateTitle("New Pass Vault Item")
var canceledEvent = internal.Event{
	Status: internal.StatusCanceled,
	Type:   internal.NewFolderRequest,
}

const (
	PasswordContinueAction int = iota
	PasswordGenerateAction
	PasswordVisibilityAction
)

func showNameForm(current string, edit bool) (string, error) {
	name := current
	confirm := true

	fields := []huh.Field{
		huh.NewNote().Title(title),
		huh.NewInput().Title("Item Name").
			Value(&name).
			Validate(func(s string) error {
				if len(s) == 0 {
					return errors.New("name cannot be blank")
				}
				return nil
			}),
	}

	if len(name) > 0 {
		fields = append(fields,
			huh.NewConfirm().
				Affirmative("Update").
				Negative("Cancel").
				Value(&confirm))
	}

	err := huh.NewForm(huh.NewGroup(fields...)).WithTheme(styles.Theme).Run()

	if edit && !confirm {
		return current, err
	}

	return name, err
}

func showUsernameForm(current string, edit bool) (string, error) {
	username := current
	confirm := true

	fields := []huh.Field{
		huh.NewNote().Title(title),
		huh.NewInput().Title("Username").
			Description("(Optional)").
			Value(&username),
	}

	if len(username) > 0 {
		fields = append(fields,
			huh.NewConfirm().
				Affirmative("Update").
				Negative("Cancel").
				Value(&confirm))
	}

	err := huh.NewForm(huh.NewGroup(fields...)).WithTheme(styles.Theme).Run()

	if edit && !confirm {
		return current, err
	}

	return username, err
}

func showURLsForm(current []*string) ([]*string, error) {
	var action int

	addURLOpt := 0
	removeURLsOpt := 1
	continueOpt := 2

	urls := current

	fields := []huh.Field{huh.NewNote().Title(title)}

	if len(urls) == 0 {
		fields = append(fields, huh.NewNote().Title("URLs").Description("None"))
	}

	for i, url := range urls {
		label := fmt.Sprintf("URL %d", i+1)
		fields = append(fields, huh.NewInput().Title(label).Value(url).Placeholder("https://..."))
	}

	actionsOpts := []huh.Option[int]{
		huh.NewOption[int]("Add URL", addURLOpt),
	}

	if len(urls) > 0 {
		actionsOpts = append(actionsOpts, huh.NewOption[int]("Remove URL", removeURLsOpt))
	}

	actionsOpts = append(actionsOpts, huh.NewOption[int]("Continue", continueOpt))
	actions := huh.NewSelect[int]().Options(actionsOpts...).Value(&action)
	fields = append(fields, actions)

	err := huh.NewForm(huh.NewGroup(fields...)).WithTheme(styles.Theme).Run()
	if err != nil {
		return nil, err
	} else if action == addURLOpt {
		newURL := ""
		urls = append(urls, &newURL)
		return showURLsForm(urls)
	} else if action == removeURLsOpt {
		if len(urls) == 1 {
			return showURLsForm(nil)
		}

		removal := 0
		removalOpts := []huh.Option[int]{huh.NewOption[int]("None", 0)}
		removalFormFields := []huh.Field{huh.NewNote().Title(title)}
		for i, url := range urls {
			label := fmt.Sprintf("URL %d", i+1)
			note := huh.NewNote().Title(label).Description(*url)
			removalFormFields = append(removalFormFields, note)
			removalOpts = append(removalOpts, huh.NewOption[int](label, i+1))
		}

		removalSelect := huh.NewSelect[int]().Options(removalOpts...).Value(&removal)
		removalFormFields = append(removalFormFields, removalSelect)

		err = huh.NewForm(huh.NewGroup(removalFormFields...)).WithTheme(styles.Theme).Run()
		if err != nil {
			return urls, err
		} else if removal > 0 {
			urls = slices.Delete(urls, removal-1, removal)
		}

		return showURLsForm(urls)
	}

	return urls, nil
}

func showPasswordForm(password string, mode huh.EchoMode) (string, bool, error) {
	var passwordAction int

	actionsOpts := []huh.Option[int]{
		huh.NewOption[int]("Continue", PasswordContinueAction),
		huh.NewOption[int]("Generate Password", PasswordGenerateAction),
		huh.NewOption[int]("Toggle Password Visibility", PasswordVisibilityAction),
	}

	actions := huh.NewSelect[int]().
		Options(actionsOpts...).
		Value(&passwordAction)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(title),
			huh.NewInput().Title("Password").
				EchoMode(mode).
				Value(&password),
			actions,
		)).WithTheme(styles.Theme).Run()

	if err == nil && passwordAction == PasswordVisibilityAction {
		if mode == huh.EchoModePassword {
			mode = huh.EchoModeNormal
		} else {
			mode = huh.EchoModePassword
		}

		return showPasswordForm(password, mode)
	}

	return password, passwordAction == PasswordGenerateAction, err
}

func showGeneratorForm() (string, error) {
	var generatorType int
	passwordGenerator := 0
	passphraseGenerator := 1
	formTitle := utils.GenerateTitle("Generate Password")

	generatorOpts := []huh.Option[int]{
		huh.NewOption[int]("Password (random characters)", passwordGenerator),
		huh.NewOption[int]("Passphrase (random words)", passphraseGenerator),
	}

	generatorSelect := huh.NewSelect[int]().
		Title("Password Type").
		Options(generatorOpts...).
		Value(&generatorType)

	err := huh.NewForm(huh.NewGroup(
		huh.NewNote().Title(formTitle),
		generatorSelect)).WithTheme(styles.Theme).Run()

	if err != nil {
		return "", err
	}

	if generatorType == passwordGenerator {
		return showPasswordGeneratorForm(defaultPasswordOpts)
	} else {
		return showPassphraseGeneratorForm(defaultPassphraseOpts)
	}
}

func showPasswordGeneratorForm(opts passwordOpts) (string, error) {
	var confirmed bool
	var genAction int

	formTitle := utils.GenerateTitle("Generate Password")
	genPassword, err := generatePassword(opts)
	if err != nil {
		return "", err
	}

	size := strconv.Itoa(opts.size)
	symbols := opts.symbols

	desc := opts.generateDescription()
	err = huh.NewForm(huh.NewGroup(
		huh.NewNote().Title(formTitle),
		huh.NewNote().Title("Generated Password").
			Description(shared.EscapeString(genPassword)),
		huh.NewNote().Title("Parameters").Description(utils.GenerateDescription(desc, 30)),
		huh.NewSelect[int]().Options(
			huh.NewOption[int]("Regenerate", 0),
			huh.NewOption[int]("Edit Parameters", 1),
			huh.NewOption[int]("Confirm Password", 2)).Value(&genAction),
	)).WithTheme(styles.Theme).Run()

	if err != nil {
		return "", err
	} else if genAction == 0 {
		return showPasswordGeneratorForm(opts)
	} else if genAction == 2 {
		return genPassword, nil
	}

	var selectedOpts []string
	selectUseUpper := "use upper"
	selectUseLower := "use lower"
	selectUseNumbers := "use numbers"
	selectUseSymbols := "use symbols"

	err = huh.NewForm(huh.NewGroup(
		huh.NewNote().Title(formTitle),
		//huh.NewNote().Title("Generated Password").Description(shared.EscapeString(genPassword)),
		huh.NewInput().Title("# of characters").
			Description("(max 99)").
			Value(&size).Validate(
			func(s string) error {
				val, convErr := strconv.Atoi(s)
				if convErr != nil {
					return convErr
				} else if val <= 0 {
					return errors.New("value must be between 1-99")
				}

				return nil
			}).CharLimit(2),
		huh.NewMultiSelect[string]().Options(
			huh.NewOption("A-Z", selectUseUpper).Selected(opts.useUpper),
			huh.NewOption("a-z", selectUseLower).Selected(opts.useLower),
			huh.NewOption("0-9", selectUseNumbers).Selected(opts.useNumbers),
			huh.NewOption("Symbols", selectUseSymbols).Selected(opts.useSymbols),
		).Value(&selectedOpts),
		huh.NewInput().Title("Symbols").Value(&symbols).CharLimit(30),
		huh.NewConfirm().
			Affirmative("Confirm").
			Negative("Cancel").
			Value(&confirmed),
	)).WithTheme(styles.Theme).Run()

	if err == nil && confirmed {
		intSize, _ := strconv.Atoi(size)
		opts.size = intSize
		opts.useUpper = shared.ArrayContains(selectedOpts, selectUseUpper)
		opts.useLower = shared.ArrayContains(selectedOpts, selectUseLower)
		opts.useNumbers = shared.ArrayContains(selectedOpts, selectUseNumbers)
		opts.useSymbols = shared.ArrayContains(selectedOpts, selectUseSymbols)
		opts.symbols = symbols

		if !opts.useUpper && !opts.useLower && !opts.useNumbers && !opts.useSymbols {
			opts.useLower = true
		}
	}

	if err == nil {
		return showPasswordGeneratorForm(opts)
	}

	return genPassword, err
}

func showPassphraseGeneratorForm(opts passphraseOpts) (string, error) {
	var genAction int

	formTitle := utils.GenerateTitle("Generate Passphrase")
	genPassphrase, err := generatePassphrase(opts)
	passphraseLen := strconv.Itoa(len(genPassphrase))
	if err != nil {
		return "", err
	}

	desc := opts.generateDescription()
	err = huh.NewForm(huh.NewGroup(
		huh.NewNote().Title(formTitle),
		huh.NewNote().Title("Generated Passphrase").
			Description(shared.EscapeString(genPassphrase)),
		huh.NewNote().Title("Passphrase Length").Description(passphraseLen),
		huh.NewNote().Title("Parameters").Description(utils.GenerateDescription(desc, 30)),
		huh.NewSelect[int]().Options(
			huh.NewOption[int]("Regenerate", 0),
			huh.NewOption[int]("Edit Parameters", 1),
			huh.NewOption[int]("Confirm Password", 2)).Value(&genAction),
	)).WithTheme(styles.Theme).Run()

	if err != nil {
		return "", err
	} else if genAction == 0 {
		return showPassphraseGeneratorForm(opts)
	} else if genAction == 2 {
		return genPassphrase, nil
	}

	numWords := strconv.Itoa(opts.numWords)
	separator := opts.separator

	var selectedOpts []string
	selectUseShortWords := "use short"
	selectCapitalize := "capitalize"
	selectUseNumber := "use number"

	confirmed := true
	err = huh.NewForm(huh.NewGroup(
		huh.NewNote().Title(formTitle),
		huh.NewInput().Title("# of words").
			Description("(max 50)").
			Value(&numWords).Validate(
			func(s string) error {
				val, convErr := strconv.Atoi(s)
				if convErr != nil {
					return convErr
				} else if val <= 0 || val > 50 {
					return errors.New("value must be between 1-50")
				}

				return nil
			}).CharLimit(2),
		huh.NewMultiSelect[string]().Options(
			huh.NewOption("Use shorter words", selectUseShortWords).Selected(opts.shortWords),
			huh.NewOption("Capitalize", selectCapitalize).Selected(opts.capitalize),
			huh.NewOption("Include number", selectUseNumber).Selected(opts.useNumber),
		).Value(&selectedOpts),
		huh.NewInput().Title("Separator").Value(&separator).CharLimit(30),
		huh.NewConfirm().
			Affirmative("Confirm").
			Negative("Cancel").
			Value(&confirmed),
	)).WithTheme(styles.Theme).Run()

	if err == nil && confirmed {
		intWords, _ := strconv.Atoi(numWords)
		opts.numWords = intWords
		opts.shortWords = shared.ArrayContains(selectedOpts, selectUseShortWords)
		opts.capitalize = shared.ArrayContains(selectedOpts, selectCapitalize)
		opts.useNumber = shared.ArrayContains(selectedOpts, selectUseNumber)
		opts.separator = separator

		if opts.shortWords {
			opts.wordlist = globals.ShortWordlist
		} else {
			opts.wordlist = globals.LongWordlist
		}
	}

	if err == nil {
		return showPassphraseGeneratorForm(opts)
	}

	return "", err
}

func showNotesForm(current string, edit bool) (string, error) {
	notes := current
	confirm := true

	fields := []huh.Field{
		huh.NewNote().Title(title),
		huh.NewText().Title("Notes").
			Description("(Optional)").
			Value(&notes),
	}

	if len(notes) > 0 {
		fields = append(fields,
			huh.NewConfirm().
				Affirmative("Update").
				Negative("Cancel").
				Value(&confirm))
	}

	err := huh.NewForm(huh.NewGroup(fields...)).WithTheme(styles.Theme).Run()

	if edit && !confirm {
		return current, err
	}

	return notes, err
}

func showPassModel(name string, entry shared.PassEntry, mode huh.EchoMode) error {
	var urlPtrs []*string
	for _, urlStr := range entry.URLs {
		urlPtrs = append(urlPtrs, &urlStr)
	}

	desc := generatePassEntryDescription(
		name,
		entry.Username,
		urlPtrs,
		entry.Password,
		mode,
		entry.Notes)

	titleNote := huh.NewNote().Title(title).Description(desc)

	action := 1
	actions := huh.NewSelect[int]().Options(
		huh.NewOption[int]("Toggle Password Visibility", 1),
		huh.NewOption[int]("Close", 0)).
		Value(&action)

	err := huh.NewForm(huh.NewGroup(titleNote, actions)).WithTheme(styles.Theme).Run()
	if err != nil {
		return err
	}

	switch action {
	case 1:
		if mode == huh.EchoModePassword {
			return showPassModel(name, entry, huh.EchoModeNormal)
		} else {
			return showPassModel(name, entry, huh.EchoModePassword)
		}
	}

	return nil
}

func showFinalizePassModel(
	name,
	username string,
	urls []*string,
	password string,
	mode huh.EchoMode,
	notes string,
) (string, shared.PassEntry, error) {
	desc := generatePassEntryDescription(name, username, urls, password, mode, notes)

	titleNote := huh.NewNote().Title(title).Description(desc)

	action := 1
	actions := huh.NewSelect[int]().Options(
		huh.NewOption[int]("Toggle Password Visibility", 1),
		huh.NewOption[int]("Edit Name", 2),
		huh.NewOption[int]("Edit Username", 3),
		huh.NewOption[int]("Edit URLs", 4),
		huh.NewOption[int]("Edit Password", 5),
		huh.NewOption[int]("Edit Note", 6),
		huh.NewOption[int]("Confirm", 0)).
		Value(&action)

	err := huh.NewForm(huh.NewGroup(titleNote, actions)).WithTheme(styles.Theme).Run()
	if err != nil {
		return "", shared.PassEntry{}, err
	}

	switch action {
	case 1:
		if mode == huh.EchoModePassword {
			return showFinalizePassModel(name, username, urls, password, huh.EchoModeNormal, notes)
		} else {
			return showFinalizePassModel(name, username, urls, password, huh.EchoModePassword, notes)
		}
	case 2:
		name, err = showNameForm(name, true)
		if err != nil {
			return "", shared.PassEntry{}, err
		}
	case 3:
		username, err = showUsernameForm(username, true)
		if err != nil {
			return "", shared.PassEntry{}, err
		}
	case 4:
		urls, err = showURLsForm(urls)
		if err != nil {
			return "", shared.PassEntry{}, err
		}
	case 5:
		var generate bool
		password, generate, err = showPasswordForm(password, mode)
		for generate {
			password, err = showGeneratorForm()
			if err != nil {
				return "", shared.PassEntry{}, err
			}
			password, generate, err = showPasswordForm(password, mode)
		}
	case 6:
		notes, err = showNotesForm(notes, true)
		if err != nil {
			return "", shared.PassEntry{}, err
		}
	}

	if action > 1 {
		return showFinalizePassModel(name, username, urls, password, mode, notes)
	} else {
		var urlStrings []string
		for _, urlPtr := range urls {
			urlStrings = append(urlStrings, *urlPtr)
		}

		return name, shared.PassEntry{
			Username:        username,
			Password:        password,
			PasswordHistory: nil,
			URLs:            urlStrings,
			Notes:           notes,
		}, nil
	}
}

func RunNewPassEntryModel() (internal.Event, error) {
	name, err := showNameForm("", false)
	if err != nil {
		return canceledEvent, err
	}

	username, err := showUsernameForm("", false)
	if err != nil {
		return canceledEvent, err
	}

	urls, err := showURLsForm(nil)
	if err != nil {
		return canceledEvent, err
	}

	password, generate, err := showPasswordForm("", huh.EchoModePassword)
	if err != nil {
		return canceledEvent, err
	} else {
		for generate {
			password, err = showGeneratorForm()
			if err != nil {
				return canceledEvent, err
			}
			password, generate, err = showPasswordForm(password, huh.EchoModePassword)
		}
	}

	notes, err := showNotesForm("", false)
	if err != nil {
		return canceledEvent, err
	}

	finalName, passEntry, err := showFinalizePassModel(
		name,
		username,
		urls,
		password,
		huh.EchoModePassword,
		notes)

	newItem := models.VaultItem{Name: finalName, PassEntry: passEntry}
	return internal.Event{
		Status: internal.StatusOk,
		Type:   internal.NewPassRequest,
		Item:   newItem,
	}, err
}

func RunViewPassEntryModel(item models.VaultItem) error {
	return showPassModel(item.Name, item.PassEntry, huh.EchoModePassword)
}

func RunEditPassEntryModel(item models.VaultItem) (internal.Event, error) {
	var urlPtrs []*string
	for _, url := range item.PassEntry.URLs {
		urlPtrs = append(urlPtrs, &url)
	}

	name, passEntry, err := showFinalizePassModel(
		item.Name,
		item.PassEntry.Username,
		urlPtrs,
		item.PassEntry.Password,
		huh.EchoModePassword,
		item.PassEntry.Notes)

	item.Name = name
	item.PassEntry = passEntry

	return internal.Event{
		Status: internal.StatusOk,
		Type:   internal.EditPassRequest,
		Item:   item,
	}, err
}
