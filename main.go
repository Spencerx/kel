package main

import (
	"fmt"
	"os"

	"github.com/kelproject/kel/cmd"
)

func main() {
	cmd.LoadConfig()
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
