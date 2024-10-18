package files

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	"sort"
	"strings"
	"time"
	"yeetfile/cli/models"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

const folderIndicator = ">"
const fileIndicator = "·"

// CreateItemRows creates a slice of table.Row elements containing the ordered
// elements to display in the vault table
func CreateItemRows(items []models.VaultItem, isPassVault bool) []table.Row {
	result := []table.Row{}

	sort.Slice(items, func(i, j int) bool {
		// Sort folders before non-folders
		if items[i].IsFolder != items[j].IsFolder {
			return items[i].IsFolder
		}

		return items[i].Modified.After(items[j].Modified)
	})

	spacing := utils.GenerateListIdxSpacing(len(items))

	var genFn func(item models.VaultItem, total, idx int, spacing string) []string
	if isPassVault {
		genFn = generatePassRow
	} else {
		genFn = generateFileRow
	}

	for idx, item := range items {
		rowStr := genFn(item, len(items), idx, spacing)
		result = append(result, rowStr)
	}

	return result
}

func generatePassRow(
	item models.VaultItem,
	total,
	idx int,
	spacing string,
) []string {
	name := item.Name

	var (
		prefix   string
		suffix   string
		username string
		url      string
	)

	if item.IsFolder {
		prefix = folderIndicator
		suffix = "/"
	} else {
		prefix = fileIndicator
		username = item.PassEntry.Username
		if len(item.PassEntry.URLs) > 0 {
			url = item.PassEntry.URLs[0]
			url = strings.ReplaceAll(url, "http://", "")
			url = strings.ReplaceAll(url, "https://", "")
			if len(item.PassEntry.URLs) > 1 {
				url += fmt.Sprintf(" (+ %d)", len(item.PassEntry.URLs)-1)
			}
		}
	}

	spacing = utils.GetListIdxSpacing(spacing, idx+1, total)
	shareIndicator := genShareIndicator(item)

	formattedName := fmt.Sprintf("%d%s| %s %s%s", idx+1, spacing, prefix, name, suffix)
	rowStr := []string{formattedName, url, username, shareIndicator}

	return rowStr
}

func generateFileRow(
	item models.VaultItem,
	total,
	idx int,
	spacing string,
) []string {
	size := shared.ReadableFileSize(item.Size - int64(constants.TotalOverhead))
	name := item.Name
	prefix := fileIndicator
	suffix := ""
	if item.IsFolder {
		prefix = folderIndicator
		suffix = "/"
		size = ""
	}

	spacing = utils.GetListIdxSpacing(spacing, idx+1, total)
	shareIndicator := genShareIndicator(item)

	formattedName := fmt.Sprintf("%d%s| %s %s%s", idx+1, spacing, prefix, name, suffix)
	modified := item.Modified.Format(time.DateOnly)
	rowStr := []string{formattedName, size, modified, shareIndicator}
	return rowStr
}

func genShareIndicator(item models.VaultItem) string {
	var shareIndicator string
	if len(item.SharedBy) > 0 {
		shareIndicator = fmt.Sprintf("<- %s", item.SharedBy)
	} else if item.SharedWith > 0 {
		shareIndicator = fmt.Sprintf("%d ->", item.SharedWith)
	} else {
		shareIndicator = "-"
	}

	return shareIndicator
}
