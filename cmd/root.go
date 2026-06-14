package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0necontroller/jsimportfmt/internal/scanner"
	"github.com/0necontroller/jsimportfmt/internal/workerpool"
	"github.com/spf13/cobra"
)

var (
	writeMode     bool
	checkMode     bool
	dryRunMode    bool
	separateTypes bool
	checkIgnore   bool
	allowDirs     []string
	allowFile     string
)

var rootCmd = &cobra.Command{
	Use:   "jsimportfmt [target]",
	Short: "A high-performance CLI tool to format JavaScript and TypeScript imports",
	Long: `jsimportfmt recursively scans JavaScript and TypeScript projects,
locates import statements, and reorders them according to configurable rules.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}
		
		allowFilePath := allowFile
		if allowFilePath == "" {
			info, err := os.Stat(target)
			dir := target
			if err == nil && !info.IsDir() {
				dir = filepath.Dir(target)
			}
			allowFilePath = filepath.Join(dir, ".jifallow")
		}

		// Keep track of what we got from CLI vs what is in the file
		cliAllowDirs := make([]string, len(allowDirs))
		copy(cliAllowDirs, allowDirs)
		
		existingAllowed := make(map[string]bool)
		content, err := os.ReadFile(allowFilePath)
		if err == nil {
			for _, line := range strings.Split(string(content), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					existingAllowed[line] = true
					// If not already in allowDirs (from CLI), add it
					found := false
					for _, d := range allowDirs {
						if d == line {
							found = true
							break
						}
					}
					if !found {
						allowDirs = append(allowDirs, line)
					}
				}
			}
		}

		// If --allow was used, persist new ones to .jifallow
		if len(cliAllowDirs) > 0 {
			var newToAppend []string
			for _, d := range cliAllowDirs {
				if !existingAllowed[d] {
					newToAppend = append(newToAppend, d)
				}
			}

			if len(newToAppend) > 0 {
				f, err := os.OpenFile(allowFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					// If file is new/empty, add the header
					info, _ := f.Stat()
					if info.Size() == 0 {
						f.WriteString("# Directories listed here will be overridden (allowed) bypassing default ignores and .gitignore\n")
					}
					for _, d := range newToAppend {
						f.WriteString(d + "\n")
					}
					f.Close()
				}
			}
		}
		
		if checkIgnore {
			matcher, err := scanner.GetMatcher(target, allowDirs)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(2)
			}
			
			printColumns := func(title string, items []string) {
				fmt.Printf("\n%s\n", title)
				if len(items) == 0 {
					fmt.Println("  (none)")
					return
				}
				
				cols := 4
				for i := 0; i < len(items); i += cols {
					for j := 0; j < cols && i+j < len(items); j++ {
						// 22 characters wide per column is generally good enough
						fmt.Printf("  %-22s", items[i+j])
					}
					fmt.Println()
				}
			}

			printColumns("Default Ignored Directories:", matcher.GetDefaultIgnored())
			printColumns("Rules from .gitignore:", matcher.GetGitignoreLines())
			return nil
		}

		isInteractive := len(args) == 0 && !writeMode && !checkMode && !dryRunMode

		if isInteractive {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Target path: ")
			targetInput, _ := reader.ReadString('\n')
			targetInput = strings.TrimSpace(targetInput)
			if targetInput != "" {
				target = targetInput
			}

			fmt.Println("Mode:\n1. Check\n2. Write\n3. Dry Run")
			fmt.Print("> ")
			modeInput, _ := reader.ReadString('\n')
			modeInput = strings.TrimSpace(modeInput)
			switch modeInput {
			case "1":
				checkMode = true
			case "2":
				writeMode = true
			case "3":
				dryRunMode = true
			default:
				writeMode = true // Default to write
			}

			fmt.Print("Separate type imports? (y/n) ")
			sepInput, _ := reader.ReadString('\n')
			sepInput = strings.TrimSpace(strings.ToLower(sepInput))
			if sepInput == "y" || sepInput == "yes" {
				separateTypes = true
			}

			fmt.Printf("\nMode: %s\nSeparate types: %v\n", map[bool]string{checkMode: "Check", writeMode: "Write", dryRunMode: "Dry Run"}[true], separateTypes)
			fmt.Print("Continue? (Y/n) ")
			contInput, _ := reader.ReadString('\n')
			contInput = strings.TrimSpace(strings.ToLower(contInput))
			if contInput == "n" || contInput == "no" {
				return nil
			}
		} else if len(args) > 0 {
			target = args[0]
		}

		if !writeMode && !checkMode && !dryRunMode {
			checkMode = true // Default mode if not specified
		}

		files := make(chan string, 100)
		errChan := make(chan error, 1)

		go scanner.Scan(target, allowDirs, files, errChan)

		results := make(chan workerpool.Result, 100)
		config := workerpool.Config{
			WriteMode:     writeMode,
			CheckMode:     checkMode,
			DryRunMode:    dryRunMode,
			SeparateTypes: separateTypes,
		}

		go workerpool.Run(files, results, config)

		var scannedCount, changedCount, errorCount, skippedCount int

		for res := range results {
			scannedCount++
			if res.Err != nil {
				errorCount++
				fmt.Printf("Warning: failed to process %s: %v\n", res.Path, res.Err)
				continue
			}

			if res.Changed {
				changedCount++
				if dryRunMode {
					fmt.Print(res.DiffString)
				}
			}
		}

		select {
		case err := <-errChan:
			fmt.Printf("✖ error: %v\n", err)
			os.Exit(2)
		default:
		}

		fmt.Printf("\n✔ scanned %d files\n", scannedCount)
		if changedCount > 0 {
			if writeMode {
				fmt.Printf("✔ formatted %d files\n", changedCount)
			} else if checkMode {
				fmt.Printf("✖ %d files need formatting\n", changedCount)
				os.Exit(1)
			}
		} else {
			if checkMode {
				fmt.Println("✔ all files formatted")
			} else {
				fmt.Println("✔ 0 files formatted")
			}
		}

		if errorCount > 0 {
			fmt.Printf("✖ parsing error in %d files\n", errorCount)
			os.Exit(2)
		}
		if skippedCount > 0 {
			fmt.Printf("⚠ skipped %d files\n", skippedCount)
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&writeMode, "write", "w", false, "Rewrite files in place")
	rootCmd.Flags().BoolVarP(&checkMode, "check", "c", false, "Check if files need formatting without changing them")
	rootCmd.Flags().BoolVar(&checkIgnore, "check-ignore", false, "List default and .gitignore rules that apply to the target")
	rootCmd.Flags().StringVar(&allowFile, "allow-file", "", "Path to a .jifallow file containing a list of directories to always allow")
	rootCmd.Flags().BoolVar(&dryRunMode, "dry-run", false, "Display unified diffs without modifying files")
	rootCmd.Flags().BoolVar(&separateTypes, "separate-types", false, "Sort and separate type imports from regular imports")
	rootCmd.Flags().StringSliceVar(&allowDirs, "allow", []string{}, "Explicitly allow ignored directories (e.g. dist, build)")
}
