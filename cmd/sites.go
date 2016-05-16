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
			fmt.Fprintf(os.Stderr, "Usage: kel sites create [name]\n")
			fatal(msg)
		}
		uri, err := LookupURI()
		if err != nil {
			fatal(err.Error())
		}
		if len(args) == 1 {
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
		} else if len(args) > 1 {
			usage("too many arguments")
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
			fmt.Fprintf(os.Stderr, "Usage: kel sites list <resource-group>\n")
			fatal(msg)
		}
		uri, err := LookupURI()
		if err != nil {
			usage(err.Error())
		}
		if len(args) == 1 {
			uri.ResourceGroup = args[0]
		} else if len(args) > 1 {
			usage("too many arguments")
		}
		if uri.ResourceGroup == "" {
			usage("missing resource group (specify with optional argument or in URI)")
		}
		kc := setupKelClient(uri)
		var resourceGroup kel.ResourceGroup
		if err := kc.ResourceGroups.Get(uri.ResourceGroup, &resourceGroup).Do(); err != nil {
			if err == kel.ErrNotFound {
				fatal(fmt.Sprintf("resource group %q does not exist.", uri.ResourceGroup))
			}
			fatal(fmt.Sprintf("failed to get resource group (error: %v)", err))
		}
		var sites []*kel.Site
		if err := kc.Sites.List(&resourceGroup, &sites).Do(); err != nil {
			fatal(fmt.Sprintf("failed to list sites (error: %v)", err))
		}
		for i := range sites {
			fmt.Println(sites[i].Name)
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
		uri, err := LookupURI()
		if err != nil {
			fatal(err.Error())
		}
		if len(args) == 1 {
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
		} else if len(args) > 1 {
			usage("too many arguments")
		}
		if uri.ResourceGroup == "" || uri.Site == "" {
			usage("must specify resource group and site in URI.")
		}
		cwd, err := os.Getwd()
		if err != nil {
			fatal(fmt.Sprintf("failed to get current working directory (%s)", err.Error()))
		}
		if _, ok := config.Sites[cwd]; ok && !flagForce {
			msg := "this directory is already activated"
			if !uri.Equals(config.Sites[cwd].URI) {
				msg += fmt.Sprintf(" for %s", config.Sites[cwd].URI)
			} else {
				msg += " for the given site"
			}
			fatal(msg + ". Use --force to override.")
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
		}
		if err := kc.Sites.Get(uri.Site, &site).Do(); err != nil {
			if err == kel.ErrNotFound {
				fatal(fmt.Sprintf("site %q does not exist.", uri.Site))
			}
			fatal(fmt.Sprintf("failed to get site (error: %v)", err))
		}
		config.Sites[cwd] = &SiteConfig{URI: uri}
		config.Save()
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
		if _, ok := config.Sites[cwd]; !ok {
			fatal("nothing to delete")
		}
		delete(config.Sites, cwd)
		config.Save()
	},
}
