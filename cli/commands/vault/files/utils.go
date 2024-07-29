package files

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	"sort"
	"time"
	"yeetfile/cli/models"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

const folderIndicator = ">"
const fileIndicator = "·"

// CreateItemRows creates a slice of table.Row elements containing the ordered
// elements to display in the vault table
func CreateItemRows(items []models.VaultItem) []table.Row {
	result := []table.Row{}

	sort.Slice(items, func(i, j int) bool {
		// Sort folders before non-folders
		if items[i].IsFolder != items[j].IsFolder {
			return items[i].IsFolder
		}

		return items[i].Modified.After(items[j].Modified)
	})

	spacing := utils.GenerateListIdxSpacing(len(items))

	for idx, item := range items {
		size := shared.ReadableFileSize(item.Size)
		name := item.Name
		prefix := fileIndicator
		suffix := ""
		if item.IsFolder {
			prefix = folderIndicator
			suffix = "/"
			size = ""
		}

		spacing = utils.GetListIdxSpacing(spacing, idx, len(items))

		var shared string
		if len(item.SharedBy) > 0 {
			shared = fmt.Sprintf("<- %s", item.SharedBy)
		} else if item.SharedWith > 0 {
			shared = fmt.Sprintf("%d ->", item.SharedWith)
		} else {
			shared = "-"
		}

		formattedName := fmt.Sprintf("%d%s| %s %s%s", idx, spacing, prefix, name, suffix)
		modified := item.Modified.Format(time.DateTime)
		rowStr := []string{formattedName, size, modified, shared}
		result = append(result, rowStr)
	}

	return result
}
