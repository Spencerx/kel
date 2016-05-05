package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// Config is the global configuration for the Kel command-line client.
type Config struct {
	DefaultCluster *URI                     `json:"cluster,omitempty"`
	SitePaths      map[string]*URI          `json:"site_paths"`
	Tokens         map[string]*oauth2.Token `json:"tokens"`
}

var config *Config

func init() {
	config = &Config{
		SitePaths: make(map[string]*URI),
		Tokens:    make(map[string]*oauth2.Token),
	}
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(
		configGetCmd,
		configSetCmd,
	)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get configuration value",
	Run: func(cmd *cobra.Command, args []string) {
		usage := func(msg string) {
			fmt.Fprintf(os.Stderr, "Usage: kel config get\n")
			fatal(msg)
		}
		if len(args) < 1 {
			usage("too few arguments.")
		}
		if len(args) > 1 {
			usage("too many arguments.")
		}
		switch args[0] {
		case "cluster":
			fmt.Println(config.DefaultCluster)
			break
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration value",
	Run: func(cmd *cobra.Command, args []string) {
		usage := func(msg string) {
			fmt.Fprintf(os.Stderr, "Usage: kel config set <uri>|<name>\n")
			fatal(msg)
		}
		if len(args) < 2 {
			usage("too few arguments.")
		}
		if len(args) > 2 {
			usage("too many arguments.")
		}
		switch args[0] {
		case "cluster":
			uri, err := ParseURI(args[1])
			if err != nil {
				fatal(fmt.Sprintf("failed to parse URI (error: %v)", err))
			}
			config.DefaultCluster = uri
			config.Save()
			break
		}
	},
}

func getConfigDir() string {
	var homeDir string
	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE")
		}
	} else {
		homeDir = os.Getenv("HOME")
	}
	return path.Join(homeDir, ".kel")
}

func getConfigPath() string {
	return path.Join(getConfigDir(), "config.json")
}

// LoadConfig loads the global Kel configuration
func LoadConfig() {
	configDir := getConfigDir()
	configPath := getConfigPath()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.Mkdir(configDir, 0755); err != nil {
			fatal(fmt.Sprintf("failed to create %s (%v)", configDir, err.Error()))
		}
		config.Save()
	} else {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			config.Save()
		}
	}
	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		fatal(fmt.Sprintf("failed to read configuration (%v)", err.Error()))
	}
	if err := json.Unmarshal(buf, config); err != nil {
		fatal(fmt.Sprintf("failed to load configuration (%v)", err.Error()))
	}
}

// Save will persist configuration to disk.
func (config *Config) Save() {
	configPath := getConfigPath()
	buf, err := json.Marshal(&config)
	if err != nil {
		fatal(fmt.Sprintf("failed to encode configuration (%v)", err.Error()))
	}
	if err := ioutil.WriteFile(configPath, buf, 0644); err != nil {
		fatal(fmt.Sprintf("failed to create config.json (%v)", err.Error()))
	}
}
