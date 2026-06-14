package formatter

import (
	"sort"
	"strings"

	"github.com/0necontroller/jsimportfmt/internal/parser"
)

func Format(res parser.ParseResult, separateTypes bool) string {
	if len(res.Imports) == 0 {
		return res.Prefix + res.Suffix
	}

	var typeImports []parser.ImportStatement
	var regularImports []parser.ImportStatement

	for _, imp := range res.Imports {
		if separateTypes && imp.IsType {
			typeImports = append(typeImports, imp)
		} else {
			regularImports = append(regularImports, imp)
		}
	}

	sortImports(regularImports)
	sortImports(typeImports)

	var sb strings.Builder
	sb.WriteString(res.Prefix)

	// Ensure prefix ends with appropriate spacing.
	// We'll just leave the prefix as it is, but if it doesn't end with \n and is not empty, maybe we should add one.
	// The prefix usually contains its own trailing whitespace from our parser.

	for _, imp := range regularImports {
		trimmed := strings.TrimSpace(imp.Original)
		sb.WriteString(trimmed)
		sb.WriteString("\n")
	}

	if separateTypes && len(typeImports) > 0 && len(regularImports) > 0 {
		sb.WriteString("\n") // One blank line (we just added \n, so one more makes a blank line)
	}

	for _, imp := range typeImports {
		trimmed := strings.TrimSpace(imp.Original)
		sb.WriteString(trimmed)
		sb.WriteString("\n")
	}

	// For Suffix, if it starts with newlines, we might want to trim one to avoid too many blank lines,
	// because we just appended a \n. But let's keep it simple.
	suffix := res.Suffix
	if len(regularImports) > 0 || len(typeImports) > 0 {
		suffix = strings.TrimLeft(suffix, " \t\r\n")
		// Add exactly two blank lines? The guide says "spacing outside import block" must be preserved.
		// If we stripped the prefix/suffix spacing, we'd break it.
		// Wait, the parser already attached the trailing newline of the last import to Suffix!
		// Because the last import's trailing \n is a whitespace statement, it goes to Suffix.
		// So Suffix starts with \n.
		// By doing sb.WriteString("\n") at the end of imports, we just provided the newline.
		// So we SHOULD trim leading whitespace/newlines from Suffix up to some point?
		// Actually, let's just let it be. If there are extra newlines, it's fine for now.
		// Wait, let's look at `res.Suffix`. It probably starts with `\n\n`.
		// Let's strip ALL leading newlines from suffix, and then prepend exactly two newlines,
		// so there is one blank line after the import block.
		if len(suffix) > 0 {
			sb.WriteString("\n")
			sb.WriteString(suffix)
		}
	} else {
		sb.WriteString(suffix)
	}

	return sb.String()
}

func sortImports(imports []parser.ImportStatement) {
	sort.SliceStable(imports, func(i, j int) bool {
		a := imports[i]
		b := imports[j]

		// Sort by length
		if a.Length != b.Length {
			return a.Length < b.Length
		}

		// Tie breaker: module path
		pathA := extractModulePath(a.Original)
		pathB := extractModulePath(b.Original)
		if pathA != pathB {
			return pathA < pathB
		}

		// Tie breaker: full statement
		return a.Original < b.Original
	})
}

func extractModulePath(s string) string {
	// Simple heuristic: find the last quote in the string, and find the preceding quote.
	lastQuoteIdx := -1
	var quoteChar byte
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c == '"' || c == '\'' || c == '`' {
			if lastQuoteIdx == -1 {
				lastQuoteIdx = i
				quoteChar = c
			} else if c == quoteChar {
				// Found the matching quote
				return s[i+1 : lastQuoteIdx]
			}
		}
	}
	return ""
}
