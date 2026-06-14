package workerpool

import (
	"os"
	"runtime"
	"sync"

	"github.com/0necontroller/jsimportfmt/internal/diff"
	"github.com/0necontroller/jsimportfmt/internal/formatter"
	"github.com/0necontroller/jsimportfmt/internal/parser"
)

type Config struct {
	WriteMode     bool
	CheckMode     bool
	DryRunMode    bool
	SeparateTypes bool
}

type Result struct {
	Path       string
	Err        error
	Changed    bool
	DiffString string
}

func Run(files <-chan string, results chan<- Result, config Config) {
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range files {
				results <- processFile(path, config)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()
}

func processFile(path string, config Config) Result {
	res := Result{Path: path}

	content, err := os.ReadFile(path)
	if err != nil {
		res.Err = err
		return res
	}

	parseResult, err := parser.Parse(content)
	if err != nil {
		res.Err = err
		return res
	}

	formatted := formatter.Format(parseResult, config.SeparateTypes)
	oldText := string(content)

	if formatted != oldText {
		res.Changed = true

		if config.DryRunMode {
			res.DiffString = diff.GenerateUnifiedDiff(path, oldText, formatted)
		} else if config.WriteMode {
			info, err := os.Stat(path)
			if err == nil {
				err = os.WriteFile(path, []byte(formatted), info.Mode())
			} else {
				err = os.WriteFile(path, []byte(formatted), 0644)
			}
			if err != nil {
				res.Err = err
			}
		}
	}

	return res
}
