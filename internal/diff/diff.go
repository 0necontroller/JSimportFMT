package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func GenerateUnifiedDiff(filename, oldText, newText string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldText, newText, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	if len(diffs) == 1 && diffs[0].Type == diffmatchpatch.DiffEqual {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s (original)\n", filename))
	sb.WriteString(fmt.Sprintf("+++ %s (formatted)\n", filename))

	for _, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			sb.WriteString(color.GreenString("+" + d.Text))
		case diffmatchpatch.DiffDelete:
			sb.WriteString(color.RedString("-" + d.Text))
		case diffmatchpatch.DiffEqual:
			sb.WriteString(d.Text)
		}
	}

	return sb.String()
}
