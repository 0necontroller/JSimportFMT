package parser

import (
	"strings"
)

type ImportStatement struct {
	Original string
	IsType   bool
	Length   int
}

type ParseResult struct {
	Prefix  string
	Imports []ImportStatement
	Suffix  string
}

// Parse extracts the leading import block from the source.
func Parse(source []byte) (ParseResult, error) {
	str := string(source)

	// We'll extract raw statements first.
	// A statement is a chunk of text.
	var statements []string

	inString := false
	var stringChar byte
	inLineComment := false
	inBlockComment := false
	bracketLevel := 0
	parenLevel := 0
	escapeNext := false

	var currentStmt strings.Builder

	for i := 0; i < len(str); i++ {
		c := str[i]
		currentStmt.WriteByte(c)

		if escapeNext {
			escapeNext = false
			continue
		}

		if c == '\\' && inString {
			escapeNext = true
			continue
		}

		if inLineComment {
			if c == '\n' {
				inLineComment = false
				if bracketLevel == 0 && parenLevel == 0 {
					statements = append(statements, currentStmt.String())
					currentStmt.Reset()
				}
			}
			continue
		}

		if inBlockComment {
			if c == '*' && i+1 < len(str) && str[i+1] == '/' {
				currentStmt.WriteByte('/')
				i++
				inBlockComment = false
			}
			continue
		}

		if inString {
			if c == stringChar {
				inString = false
			}
			continue
		}

		// Not in any string or comment
		switch c {
		case '/':
			if i+1 < len(str) {
				if str[i+1] == '/' {
					inLineComment = true
					currentStmt.WriteByte('/')
					i++
				} else if str[i+1] == '*' {
					inBlockComment = true
					currentStmt.WriteByte('*')
					i++
				}
			}
		case '"', '\'', '`':
			inString = true
			stringChar = c
		case '{':
			bracketLevel++
		case '}':
			bracketLevel--
		case '(':
			parenLevel++
		case ')':
			parenLevel--
		case ';':
			if bracketLevel == 0 && parenLevel == 0 {
				statements = append(statements, currentStmt.String())
				currentStmt.Reset()
			}
		case '\n':
			if bracketLevel == 0 && parenLevel == 0 {
				// We consider a newline to terminate a statement if we are at top-level.
				// However, if the statement is `import` but hasn't seen a string yet, maybe we shouldn't?
				// For now, let's just break on newline.
				s := currentStmt.String()
				// If it's just whitespaces, we might keep appending?
				// Let's just append it. We'll group them later.
				statements = append(statements, s)
				currentStmt.Reset()
			}
		}
	}

	if currentStmt.Len() > 0 {
		statements = append(statements, currentStmt.String())
	}

	// Group statements into complete imports
	var mergedStatements []string
	for i := 0; i < len(statements); i++ {
		stmt := statements[i]

		// Skip empty or whitespace-only statements unless they are the last ones?
		// We'll just group them.

		// If it looks like an import but isn't complete, merge forward.
		for isImportStatement(stmt) && !isCompleteImport(stmt) && i+1 < len(statements) {
			i++
			stmt += statements[i]
		}
		mergedStatements = append(mergedStatements, stmt)
	}

	var res ParseResult
	inImportBlock := false
	foundImportBlock := false

	for _, stmt := range mergedStatements {
		if !foundImportBlock {
			// Looking for the first import
			if isImportStatement(stmt) {
				foundImportBlock = true
				inImportBlock = true
				res.Imports = append(res.Imports, createImportStatement(stmt))
			} else {
				res.Prefix += stmt
			}
		} else if inImportBlock {
			// We are inside the import block
			if isImportStatement(stmt) {
				res.Imports = append(res.Imports, createImportStatement(stmt))
			} else if strings.TrimSpace(stmt) == "" || strings.HasPrefix(strings.TrimSpace(stmt), "//") || strings.HasPrefix(strings.TrimSpace(stmt), "/*") {
				// It's just a comment or whitespace. We should attach it to the next import, OR if the import block ends, to the suffix.
				// But to keep it simple, we could attach it to the NEXT import if there is one.
				// If there's no next import, it belongs to the suffix.
				// Since we don't know the future, let's look ahead.
				// But wait, the guide says "Extract import statements verbatim".
				// A comment BETWEEN imports belongs to the imports.
				// Let's treat comments between imports as part of the import block.
				// We can just append it as a "dummy" import or prepend it to the next import.
				// Prepending to the next import is better.
				// For now, let's say: if it's whitespace/comment, we append it to the LAST import.
				// Wait, if we sort them, the comment will move with the last import.
				// Is it better to attach it to the next import? Yes.
				// Let's create an "Accumulator" for leading whitespace/comments.
				// We will change the loop to handle this.
			} else {
				// Found non-import, end of block
				inImportBlock = false
				res.Suffix += stmt
			}
		} else {
			// After the import block
			res.Suffix += stmt
		}
	}

	// Wait, we need a better way to handle comments/whitespaces between imports.
	// Let's refine the loop.
	res = ParseResult{}
	foundImportBlock = false
	inImportBlock = false
	var pendingWhitespace string

	for _, stmt := range mergedStatements {
		trimmed := strings.TrimSpace(stmt)
		isWhitespaceOrComment := trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*")

		if !foundImportBlock {
			if isImportStatement(stmt) {
				foundImportBlock = true
				inImportBlock = true

				// Split pendingWhitespace. If there's a double newline, everything up to the last double newline goes to prefix.
				lastDoubleNL := strings.LastIndex(pendingWhitespace, "\n\n")
				if lastDoubleNL == -1 {
					lastDoubleNL = strings.LastIndex(pendingWhitespace, "\r\n\r\n")
					if lastDoubleNL != -1 {
						lastDoubleNL += 1 // to account for the \r
					}
				}

				if lastDoubleNL != -1 {
					splitIdx := lastDoubleNL + 2
					res.Prefix += pendingWhitespace[:splitIdx]
					pendingWhitespace = pendingWhitespace[splitIdx:]
				} else if strings.TrimSpace(pendingWhitespace) == "" {
					res.Prefix += pendingWhitespace
					pendingWhitespace = ""
				} else {
					// Also if there's no double newline but the file starts with a comment and NO blank lines, it's tricky.
					// Often we just attach it to the import.
				}

				res.Imports = append(res.Imports, createImportStatement(pendingWhitespace+stmt))
				pendingWhitespace = ""
			} else {
				if isWhitespaceOrComment {
					// Could be attached to the first import!
					pendingWhitespace += stmt
				} else {
					res.Prefix += pendingWhitespace + stmt
					pendingWhitespace = ""
				}
			}
		} else if inImportBlock {
			if isImportStatement(stmt) {
				res.Imports = append(res.Imports, createImportStatement(pendingWhitespace+stmt))
				pendingWhitespace = ""
			} else if isWhitespaceOrComment {
				pendingWhitespace += stmt
			} else {
				inImportBlock = false
				res.Suffix += pendingWhitespace + stmt
				pendingWhitespace = ""
			}
		} else {
			res.Suffix += stmt
		}
	}

	if pendingWhitespace != "" {
		if !foundImportBlock {
			res.Prefix += pendingWhitespace
		} else if inImportBlock {
			// trailing whitespace after last import belongs to suffix
			res.Suffix += pendingWhitespace
		} else {
			res.Suffix += pendingWhitespace
		}
	}

	return res, nil
}

func isImportStatement(s string) bool {
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "import") {
		if len(trimmed) >= 6 {
			if len(trimmed) == 6 {
				return true
			}
			nextChar := trimmed[6]
			if nextChar == ' ' || nextChar == '\t' || nextChar == '\n' || nextChar == '{' || nextChar == '"' || nextChar == '\'' || nextChar == '*' {
				return true
			}
		}
	}

	if strings.HasPrefix(trimmed, "const ") || strings.HasPrefix(trimmed, "let ") || strings.HasPrefix(trimmed, "var ") {
		if strings.Contains(trimmed, "require(") || strings.Contains(trimmed, "require (") {
			return true
		}
	}
	return false
}

func isCompleteImport(s string) bool {
	trimmed := strings.TrimSpace(s)
	if strings.HasSuffix(trimmed, ";") {
		return true
	}
	if strings.HasPrefix(trimmed, "import") {
		return strings.Contains(s, "\"") || strings.Contains(s, "'") || strings.Contains(s, "`")
	}
	if strings.HasPrefix(trimmed, "const") || strings.HasPrefix(trimmed, "let") || strings.HasPrefix(trimmed, "var") {
		return strings.Contains(s, ")")
	}
	return true
}

func createImportStatement(s string) ImportStatement {
	// Trim leading whitespace and comments to find the actual import code
	code := s
	for {
		trimmed := strings.TrimSpace(code)
		if strings.HasPrefix(trimmed, "//") {
			idx := strings.Index(trimmed, "\n")
			if idx == -1 {
				code = ""
				break
			}
			code = trimmed[idx+1:]
		} else if strings.HasPrefix(trimmed, "/*") {
			idx := strings.Index(trimmed, "*/")
			if idx == -1 {
				code = ""
				break
			}
			code = trimmed[idx+2:]
		} else {
			code = trimmed
			break
		}
	}

	isType := strings.HasPrefix(code, "import type")
	return ImportStatement{
		Original: s,
		IsType:   isType,
		Length:   len([]rune(code)), // Use rune count of the actual code for length
	}
}
