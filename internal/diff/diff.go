package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func GenerateUnifiedDiff(filename, oldText, newText string) string {
	dmp := diffmatchpatch.New()
	
	// Use line mode for better patch formatting
	a, b, c := dmp.DiffLinesToChars(oldText, newText)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, c)
	
	if len(diffs) == 1 && diffs[0].Type == diffmatchpatch.DiffEqual {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s (original)\n", filename))
	sb.WriteString(fmt.Sprintf("+++ %s (formatted)\n", filename))

	// We'll iterate over diffs and manually build the hunk strings
	// Custom Context formatter
	contextLines := 3
	type hunk struct {
		start1, len1 int
		start2, len2 int
		lines []string
	}
	
	var hunks []hunk
	var currentHunk *hunk
	
	line1 := 1
	line2 := 1
	
	for i := 0; i < len(diffs); i++ {
		d := diffs[i]
		lines := strings.Split(d.Text, "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		
		if d.Type == diffmatchpatch.DiffEqual {
			if currentHunk != nil {
				if len(lines) <= contextLines*2 {
					for _, l := range lines {
						currentHunk.lines = append(currentHunk.lines, " "+l)
						currentHunk.len1++
						currentHunk.len2++
					}
				} else {
					for j := 0; j < contextLines; j++ {
						currentHunk.lines = append(currentHunk.lines, " "+lines[j])
						currentHunk.len1++
						currentHunk.len2++
					}
					hunks = append(hunks, *currentHunk)
					currentHunk = nil
				}
			}
			line1 += len(lines)
			line2 += len(lines)
		} else {
			if currentHunk == nil {
				// Find start context
				ctxStart := line1 - contextLines
				
				var ctxLines []string
				if i > 0 && diffs[i-1].Type == diffmatchpatch.DiffEqual {
					prevLines := strings.Split(diffs[i-1].Text, "\n")
					if len(prevLines) > 0 && prevLines[len(prevLines)-1] == "" {
						prevLines = prevLines[:len(prevLines)-1]
					}
					
					start := len(prevLines) - contextLines
					if start < 0 { start = 0 }
					ctxStart = line1 - (len(prevLines) - start)
					ctxLines = prevLines[start:]
				} else {
					ctxStart = line1
				}
				
				currentHunk = &hunk{
					start1: ctxStart, len1: len(ctxLines),
					start2: ctxStart, len2: len(ctxLines), // Approximate for now
				}
				for _, l := range ctxLines {
					currentHunk.lines = append(currentHunk.lines, " "+l)
				}
			}
			
			for _, l := range lines {
				if d.Type == diffmatchpatch.DiffInsert {
					currentHunk.lines = append(currentHunk.lines, color.GreenString("+"+l))
					currentHunk.len2++
					line2++
				} else {
					currentHunk.lines = append(currentHunk.lines, color.RedString("-"+l))
					currentHunk.len1++
					line1++
				}
			}
		}
	}
	
	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}
	
	for _, h := range hunks {
		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", h.start1, h.len1, h.start2, h.len2))
		for _, l := range h.lines {
			sb.WriteString(l + "\n")
		}
	}

	return sb.String()
}
