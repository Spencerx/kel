package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	AuthNone    = "none"
	AuthCluster = "cluster"
)

// Config is the global configuration for the Kel command-line client.
type Config struct {
	DefaultCluster *URI                     `json:"cluster,omitempty"`
	Auth           string                   `json:"auth,omitempty"`
	Sites          map[string]*SiteConfig   `json:"sites"`
	Tokens         map[string]*oauth2.Token `json:"tokens"`
	Plugins        map[string]*Plugin       `json:"plugins"`
}

type SiteConfig struct {
	URI     *URI              `json:"uri,omitempty"`
	Plugins map[string]string `json:"plugins"`
}

var config *Config

func init() {
	config = &Config{
		Auth:    AuthCluster,
		Sites:   make(map[string]*SiteConfig),
		Tokens:  make(map[string]*oauth2.Token),
		Plugins: make(map[string]*Plugin),
	}
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(
		configGetCmd,
		configSetCmd,
	)
}

type TokenSaver interface {
	Save(*oauth2.Token) error
}

type cachedTokenSource struct {
	pts oauth2.TokenSource // called when t is expired.
	ts  TokenSaver
}

func newCachedTokenSource(src oauth2.TokenSource, ts TokenSaver) oauth2.TokenSource {
	return &cachedTokenSource{
		pts: src,
		ts:  ts,
	}
}

func (s *cachedTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.pts.Token()
	if err != nil {
		return nil, err
	}
	if err := s.ts.Save(t); err != nil {
		return nil, err
	}
	return t, nil
}

type configTokenSaver struct {
	provider string
	mtx      sync.Mutex
}

func (cts *configTokenSaver) Save(token *oauth2.Token) error {
	cts.mtx.Lock()
	defer cts.mtx.Unlock()
	config.Tokens[cts.provider] = token
	config.Save()
	return nil
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
		case "auth":
			fmt.Println(config.Auth)
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
			config.DefaultCluster = &uri
			config.Save()
			break
		case "auth":
			switch args[1] {
			case AuthNone, AuthCluster:
				config.Auth = args[1]
				config.Save()
				break
			default:
				fatal("invalid authentication type")
			}
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

// AddPlugin will add the given plugin to the site config.
func (config *Config) AddPlugin(plugin *Plugin) {
	if config.Plugins == nil {
		config.Plugins = make(map[string]*Plugin)
	}
	if _, ok := config.Plugins[plugin.String()]; !ok {
		config.Plugins[plugin.String()] = plugin
	}
}

// Save will persist configuration to disk.
func (config *Config) Save() {
	configPath := getConfigPath()
	buf, err := json.Marshal(&config)
	if err != nil {
		fatal(fmt.Sprintf("failed to encode configuration (%v)", err.Error()))
	}
	var out bytes.Buffer
	json.Indent(&out, buf, "", "  ")
	if err := ioutil.WriteFile(configPath, out.Bytes(), 0644); err != nil {
		fatal(fmt.Sprintf("failed to create config.json (%v)", err.Error()))
	}
}

// AddPlugin will add the given plugin to the site config.
func (siteConfig *SiteConfig) AddPlugin(plugin *Plugin) {
	if siteConfig.Plugins == nil {
		siteConfig.Plugins = make(map[string]string)
	}
	siteConfig.Plugins[plugin.Name] = fmt.Sprintf("=%s", plugin.Version)
}
