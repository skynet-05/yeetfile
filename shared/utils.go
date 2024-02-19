package shared

import (
	"bufio"
	"fmt"
)

func ReadableFileSize(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGT"[exp])
}

// IsPlaintext checks the scanner of either a raw string or file contents
// and determines if the file is contains non-ascii characters
func IsPlaintext(scanner *bufio.Scanner) bool {
	for scanner.Scan() {
		for _, r := range scanner.Text() {
			if r > 127 {
				return false // Non-ASCII character found
			}
		}
	}

	return true
}
