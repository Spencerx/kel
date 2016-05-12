package cmd

import (
	"fmt"
	"os"

	"github.com/mgutz/ansi"
)

var (
	red       = ansi.ColorFunc("red")
	green     = ansi.ColorFunc("green")
	whiteBold = ansi.ColorFunc("white+bold")
)

func success(s string) {
	fmt.Fprintf(os.Stdout, "%s %s\n", green("Success:"), s)
}

func failure(s string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", red("Error:"), s)
}

func fatal(s string) {
	failure(s)
	os.Exit(1)
}
