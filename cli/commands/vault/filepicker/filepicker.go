package filepicker

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"yeetfile/shared"
)

type DirEntries []item

func (a DirEntries) Len() int      { return len(a) }
func (a DirEntries) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DirEntries) Less(i, j int) bool {
	if a[i].isDir != a[j].isDir {
		return a[i].isDir
	}
	return a[i].name < a[j].name
}

func goUpDir(dir string) string {
	parentDir := filepath.Dir(dir)
	return parentDir
}

func appendDir(currentDir, newDir string) string {
	if strings.HasSuffix(currentDir, string(os.PathSeparator)) {
		return currentDir + newDir
	}

	return fmt.Sprintf("%s%c%s", currentDir, os.PathSeparator, newDir)
}

func getItemPath(dir, filename string) string {
	if strings.HasSuffix(dir, string(os.PathSeparator)) {
		return dir + filename
	}
	return fmt.Sprintf("%s%c%s", dir, os.PathSeparator, filename)
}

func getItemsFromDir(dir string) ([]list.Item, error) {
	var items []item
	var result []list.Item
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		info, _ := entry.Info()
		newItem := item{
			name:  entry.Name(),
			size:  shared.ReadableFileSize(info.Size()),
			perm:  shared.EscapeString(info.Mode().String()),
			isDir: entry.IsDir(),
		}

		items = append(items, newItem)
	}

	sort.Sort(DirEntries(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result, nil
}
