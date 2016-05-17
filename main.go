package main

import (
	"fmt"
	"os"

	"github.com/kelproject/kel/cmd"
)

func main() {
	cmd.LoadConfig()
	cmd.LoadPlugins()
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
