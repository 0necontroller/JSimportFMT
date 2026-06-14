package cmd

import (
	"bufio"
	"fmt"
	"os"
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
		
		if checkIgnore {
			matcher, err := scanner.GetMatcher(target, allowDirs)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(2)
			}
			
			fmt.Println("Default Ignored Directories:")
			for _, dir := range matcher.GetDefaultIgnored() {
				fmt.Printf("  - %s\n", dir)
			}
			
			lines := matcher.GetGitignoreLines()
			if len(lines) > 0 {
				fmt.Println("\nRules from .gitignore:")
				for _, line := range lines {
					fmt.Printf("  - %s\n", line)
				}
			} else {
				fmt.Println("\nNo rules found in .gitignore (or file not found).")
			}
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
			if err == scanner.ErrNotInGitRepo {
				os.Exit(2)
			}
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
	rootCmd.Flags().BoolVar(&dryRunMode, "dry-run", false, "Display unified diffs without modifying files")
	rootCmd.Flags().BoolVar(&separateTypes, "separate-types", false, "Sort and separate type imports from regular imports")
	rootCmd.Flags().StringSliceVar(&allowDirs, "allow", []string{}, "Explicitly allow ignored directories (e.g. dist, build)")
}
