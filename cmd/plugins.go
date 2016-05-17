package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/blang/semver"
	"github.com/kelproject/kel-go"
	"github.com/spf13/cobra"
)

// Plugin represents a Kel client plugin.
type Plugin struct {
	Name    string        `json:"name,omitempty"`
	Version string        `json:"version,omitempty"`
	Command PluginCommand `json:"command,omitempty"`
}

// PluginCommand represents the client command plugin
type PluginCommand struct {
	BinaryURL string `json:"binary_url,omitempty"`
	Use       string `json:"use,omitempty"`
	Short     string `json:"short,omitempty"`
}

// LoadPlugins will load configured plugins for the activated site.
func LoadPlugins() {
	siteConfig := GetActivatedSiteConfig()
	if siteConfig != nil {
		for pluginName, pluginVersionRange := range siteConfig.Plugins {
			var plugin *Plugin
			vRange, err := semver.ParseRange(pluginVersionRange)
			if err != nil {
				fatal(fmt.Sprintf("site plugin %q version range %q is invalid.", pluginName, pluginVersionRange))
			}
			for _, plugin = range config.Plugins {
				v, err := semver.Make(plugin.Version)
				if err != nil {
					fatal(fmt.Sprintf("plugin %q version %q is invalid.", plugin.Name, plugin.Version))
				}
				if plugin.Name == pluginName && vRange(v) {
					break
				}
				plugin = nil
			}
			if plugin == nil {
				fatal(fmt.Sprintf("plugin matching %s %s was not found.", pluginName, pluginVersionRange))
			}
			RootCmd.AddCommand(plugin.AsCmd())
		}
	}
}

// SyncSitePlugins will sync the local state of plugins match what the site
// is providing.
func SyncSitePlugins(site *kel.Site) {
	siteConfig := GetActivatedSiteConfig()

	fmt.Printf("Fetching plugins... ")
	time.Sleep(2 * time.Second)
	plugin := &Plugin{
		Name:    "kel-build",
		Version: "0.1.0",
		Command: PluginCommand{
			Use:       "build",
			Short:     "Build me",
			BinaryURL: "http://localhost:8080/kel-build",
		},
	}
	fmt.Println(green("done"))

	if _, ok := config.Plugins[plugin.String()]; !ok {
		fmt.Printf("Installing plugin %q... ", plugin.Name)
		if err := plugin.Install(); err != nil {
			fmt.Printf("%s\n", red("error"))
			fatal(err.Error())
		}
		siteConfig.AddPlugin(plugin)
		config.AddPlugin(plugin)
		fmt.Printf("%s (version: %s)\n", green("installed"), whiteBold(plugin.Version))
	}

	config.Save()
}

// Install will download and install the plugin binary.
func (plugin *Plugin) Install() error {
	pluginDir := path.Dir(plugin.BinaryPath())
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		if err := os.MkdirAll(pluginDir, os.FileMode(0755)); err != nil {
			if err != nil {
				return fmt.Errorf("unable to create directory %q: %s", pluginDir, err.Error())
			}
		}
	}
	f, err := os.Create(plugin.BinaryPath())
	if err != nil {
		return err
	}
	defer f.Close()
	resp, err := http.Get(plugin.Command.BinaryURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	if err := f.Chmod(os.FileMode(0755)); err != nil {
		return err
	}
	return nil
}

// BinaryPath will return the full filesystem path to the binary for this
// plugin.
func (plugin *Plugin) BinaryPath() string {
	return path.Join(
		getConfigDir(),
		"plugins",
		fmt.Sprintf("%s-v%s", plugin.Name, plugin.Version),
	)
}

// AsCmd returns a cobra.Command based on dynamic plugin values.
func (plugin *Plugin) AsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   plugin.Command.Use,
		Short: plugin.Command.Short,
		Run: func(cmd *cobra.Command, args []string) {
			args = append([]string{path.Base(plugin.BinaryPath())}, args...)
			env := os.Environ()
			if err := syscall.Exec(plugin.BinaryPath(), args, env); err != nil {
				fatal(fmt.Sprintf("failed executing plugin %q binary %s: %s", plugin.Name, plugin.BinaryPath(), err.Error()))
			}
		},
	}
}

func (plugin *Plugin) String() string {
	return fmt.Sprintf("%s==%s", plugin.Name, plugin.Version)
}
