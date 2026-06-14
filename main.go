package main

import (
	"fmt"
	"os"

	"github.com/0necontroller/jsimportfmt/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
