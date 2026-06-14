# jsimportfmt

A high-performance CLI tool written in Go to automatically format and sort JavaScript and TypeScript imports by length. Optimized for very large repositories and CI/CD pipelines.

## Features

- **Blazing Fast**: Written in Go using concurrent worker pools.
- **Smart Parsing**: Uses a lightweight state-machine parser instead of a slow AST, ensuring comments, spacing, and formatting within imports are perfectly preserved.
- **Git & Gitignore Aware**: Automatically respects `.gitignore` rules and ensures execution within a Git repository.
- **Type Separation**: Built-in support to separate type imports (`--separate-types`).
- **Interactive Mode**: Launch a guided session by simply running `jsimportfmt`.
- **Multiple Modes**: Supports Write (`--write`), Check (`--check`), and Dry-Run (`--dry-run`) modes.

## Installation

```bash
go install github.com/0necontroller/jsimportfmt@latest
```

## Build Instructions

```bash
git clone https://github.com/0necontroller/jsimportfmt.git
cd jsimportfmt
go build -o jsimportfmt .
```

## Usage Examples

Run interactively:
```bash
jsimportfmt
```

Format a specific directory:
```bash
jsimportfmt src --write
```

Format and separate type imports:
```bash
jsimportfmt src --write --separate-types
```

Check if files need formatting (useful for CI):
```bash
jsimportfmt src --check
```

See what would change without modifying files:
```bash
jsimportfmt src --dry-run
```

Allow formatting in typically ignored directories:
```bash
jsimportfmt . --allow build --allow dist
```

## CI Usage

You can use `jsimportfmt` in your CI pipeline to enforce import formatting. Use the `--check` flag:

```yaml
steps:
  - name: Check Import Formatting
    run: jsimportfmt . --check
```

### Exit Codes
- `0`: Success (no formatting needed, or all files successfully formatted in write mode)
- `1`: Formatting needed (returned in `--check` mode)
- `2`: Fatal error (e.g., target is not inside a Git repository, or file parsing failed)

## Performance Notes
`jsimportfmt` streams file discovery and feeds a worker pool matched to your system's CPU count (`runtime.NumCPU()`). It reads, parses, and formats files incrementally, ensuring low memory overhead even for monorepos with tens of thousands of files.
