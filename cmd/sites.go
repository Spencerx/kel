package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kelproject/kel-go"
	"github.com/spf13/cobra"
)

var (
	flagResourceGroupName string
	flagForce             bool
)

func init() {
	RootCmd.AddCommand(sitesCmd)
	sitesCmd.AddCommand(
		sitesCreateCmd,
		sitesListCmd,
	)
	sitesCmd.PersistentFlags().StringVarP(&flagResourceGroupName, "resource-group", "", "", "Name of resource group")

	RootCmd.AddCommand(activateCmd)
	activateCmd.Flags().BoolVarP(&flagForce, "force", "", false, "Force activation of site")
	RootCmd.AddCommand(deactivateCmd)
}

var sitesCmd = &cobra.Command{
	Use:   "sites",
	Short: "Manage sites",
}

var sitesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a site",
	Run: func(cmd *cobra.Command, args []string) {
		usage := func(msg string) {
			fmt.Fprintf(os.Stderr, "Usage: kel sites create <uri>|<name>\n")
			fatal(msg)
		}
		if len(args) != 1 {
			usage("too few arguments.")
		}
		var uri *URI
		var err error
		uri, err = ParseURI(args[0])
		if err != nil {
			if config.DefaultCluster == nil {
				usage("first argument must be a URI or a default cluster must be set.")
			}
			uri = &URI{}
			*uri = *config.DefaultCluster
			switch strings.Count(args[0], "/") {
			case 0:
				uri.Site = args[0]
				break
			case 1:
				parts := strings.Split(args[0], "/")
				uri.ResourceGroup = parts[0]
				uri.Site = parts[1]
				break
			default:
				usage("invalid resource group / site pair")
			}
		}
		if uri.ResourceGroup == "" || uri.Site == "" {
			usage("must specify resource group and site in URI.")
		}
		kc := setupKelClient(uri)
		var resourceGroup kel.ResourceGroup
		if err := kc.ResourceGroups.Get(uri.ResourceGroup, &resourceGroup).Do(); err != nil {
			if err == kel.ErrNotFound {
				fatal(fmt.Sprintf("resource group %q does not exist.", uri.ResourceGroup))
			}
			fatal(fmt.Sprintf("failed to get resource group (error: %v)", err))
		}
		site := kel.Site{
			ResourceGroup: &resourceGroup,
			Name:          uri.Site,
		}
		if err = kc.Sites.Create(&site).Do(); err != nil {
			fatal(fmt.Sprintf("failed to create site (error: %v)", err))
		}
		success(fmt.Sprintf("created %q site.", fmt.Sprintf("%s/%s", site.ResourceGroup.Name, site.Name)))
	},
}

var sitesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sites",
	Run: func(cmd *cobra.Command, args []string) {
		usage := func(msg string) {
			fmt.Fprintf(os.Stderr, "Usage: kel sites list <uri>|<name>\n")
			fatal(msg)
		}
	},
}

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate a site",
	Run: func(cmd *cobra.Command, args []string) {
		usage := func(msg string) {
			fmt.Fprintf(os.Stderr, "Usage: kel activate [--force] <site-url>\n")
			fatal(msg)
		}
		if len(args) != 1 {
			usage("not enough arguments")
		}
		cwd, err := os.Getwd()
		if err != nil {
			fatal(fmt.Sprintf("failed to get current working directory (%s)", err.Error()))
		}
		uri, err := ParseURI(args[0])
		if err != nil {
			fatal(fmt.Sprintf("failed to parse site URL (%v)", err))
		}
		if uri.Site == "" {
			fatal(fmt.Sprintf("no site provided in URI (given: %s)", uri))
		}
		if _, ok := config.SitePaths[cwd]; ok && !flagForce {
			msg := "this directory is already activated"
			if !uri.Equals(config.SitePaths[cwd]) {
				msg += fmt.Sprintf(" for %s", config.SitePaths[cwd])
			} else {
				msg += " for the given site"
			}
			fatal(msg + ". Use --force to override.")
		}
		config.SitePaths[cwd] = uri
		config.Save()
		kc := setupKelClient(uri)
		var resourceGroup kel.ResourceGroup
		if err := kc.ResourceGroups.Get(uri.ResourceGroup, &resourceGroup).Do(); err != nil {
			if err == kel.ErrNotFound {
				fatal(fmt.Sprintf("resource group %q does not exist.", uri.ResourceGroup))
			}
			fatal(fmt.Sprintf("failed to get resource group (error: %v)", err))
		}
		var site kel.Site
		if err := kc.Sites.Get(uri.Site, &site).Do(); err != nil {
			if err == kel.ErrNotFound {
				fatal(fmt.Sprintf("site %q does not exist.", uri.Site))
			}
			fatal(fmt.Sprintf("failed to get site (error: %v)", err))
		}
		success(fmt.Sprintf("%s/%s has been activated.", uri.ResourceGroup, uri.Site))
	},
}

var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate a site",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fatal(fmt.Sprintf("failed to get current working directory (%s)", err.Error()))
		}
		if _, ok := config.SitePaths[cwd]; !ok {
			fatal("nothing to delete")
		}
		delete(config.SitePaths, cwd)
		config.Save()
	},
}
