package runner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// confirmWrite shows the proposed file content and asks the user to confirm.
// Returns true if the user confirmed, false if rejected or input ended.
func confirmWrite(path, proposed string, in io.Reader, out io.Writer) (bool, error) {
	existing, err := os.ReadFile(path) //nolint:gosec
	isNew := err != nil

	fmt.Fprintln(out)
	if isNew {
		fmt.Fprintf(out, "━━━ NEW FILE: %s ━━━\n", path)
	} else {
		fmt.Fprintf(out, "━━━ MODIFIED: %s ━━━\n", path)
		printDiff(string(existing), proposed, out)
		fmt.Fprintln(out, "━━━━━━━━━━━━━━━━━━")
		fmt.Fprintf(out, "\nWrite %s? [y/N] ", path)
		answer, err := readline(in)
		if err != nil {
			return false, err
		}
		return strings.ToLower(strings.TrimSpace(answer)) == "y", nil
	}

	// New file: show the full content (truncated if large).
	lines := strings.Split(proposed, "\n")
	const maxPreview = 60
	for i, line := range lines {
		if i >= maxPreview {
			fmt.Fprintf(out, "  ... (%d more lines)\n", len(lines)-maxPreview)
			break
		}
		fmt.Fprintf(out, "  %s\n", line)
	}
	fmt.Fprintln(out, "━━━━━━━━━━━━━━━━━━")
	fmt.Fprintf(out, "\nWrite %s? [y/N] ", path)
	answer, err := readline(in)
	if err != nil {
		return false, err
	}
	return strings.ToLower(strings.TrimSpace(answer)) == "y", nil
}

// printDiff prints a simple +/- line diff between old and new content.
func printDiff(old, new string, out io.Writer) {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	// Use a basic LCS-free approach: show changed regions.
	// For a first implementation we use a simple side-by-side scan.
	maxOld := len(oldLines)
	maxNew := len(newLines)
	i, j := 0, 0
	changed := 0
	const maxDiffLines = 80

	for (i < maxOld || j < maxNew) && changed < maxDiffLines {
		switch {
		case i >= maxOld:
			fmt.Fprintf(out, "\033[32m+ %s\033[0m\n", newLines[j])
			j++
			changed++
		case j >= maxNew:
			fmt.Fprintf(out, "\033[31m- %s\033[0m\n", oldLines[i])
			i++
			changed++
		case oldLines[i] == newLines[j]:
			// Unchanged: show a few lines of context but skip long runs.
			fmt.Fprintf(out, "  %s\n", oldLines[i])
			i++
			j++
		default:
			fmt.Fprintf(out, "\033[31m- %s\033[0m\n", oldLines[i])
			fmt.Fprintf(out, "\033[32m+ %s\033[0m\n", newLines[j])
			i++
			j++
			changed++
		}
	}
	if changed >= maxDiffLines {
		fmt.Fprintln(out, "  ... (diff truncated)")
	}
}

func readline(in io.Reader) (string, error) {
	scanner := bufio.NewScanner(in)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}
