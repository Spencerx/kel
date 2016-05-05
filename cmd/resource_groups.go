package cmd

import (
	"fmt"
	"os"

	"github.com/kelproject/kel-go"
	"github.com/spf13/cobra"
)

var (
	flagToken string
)

func init() {
	RootCmd.AddCommand(resourceGroupsCmd)
	resourceGroupsCmd.AddCommand(
		resourceGroupCreateCmd,
		resourceGroupListCmd,
	)

	resourceGroupCreateCmd.Flags().StringVarP(&flagToken, "token", "", "", "Token to create resource group")
}

var resourceGroupsCmd = &cobra.Command{
	Use:   "resource-groups",
	Short: "Manage resource groups",
}

var resourceGroupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a resource group",
	Run: func(cmd *cobra.Command, args []string) {
		usage := func(msg string) {
			fmt.Fprintf(os.Stderr, "Usage: kel resource-groups create <uri>|<name>\n")
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
			uri.ResourceGroup = args[0]
		}
		if uri.ResourceGroup == "" {
			usage("must specify resource group in URI.")
		}
		kc := setupKelClient(uri)
		resourceGroup := kel.ResourceGroup{
			Name: uri.ResourceGroup,
		}
		if flagToken != "" {
			err = kc.ResourceGroups.CreateWithToken(&resourceGroup, flagToken).Do()
		} else {
			err = kc.ResourceGroups.Create(&resourceGroup).Do()
		}
		if err != nil {
			fatal(fmt.Sprintf("failed to create resource group (error: %v)", err))
		}
		success(fmt.Sprintf("created %q resource group.", resourceGroup.Name))
	},
}

var resourceGroupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List resource groups",
	Run: func(cmd *cobra.Command, args []string) {
		usage := func(msg string) {
			fmt.Fprintf(os.Stderr, "Usage: kel resource-groups list [uri]\n")
			fatal(msg)
		}
		var uri *URI
		if len(args) == 0 {
			uri = config.DefaultCluster
		} else if len(args) == 1 {
			var err error
			uri, err = ParseURI(args[0])
			if err != nil {
				usage(fmt.Sprintf("first argument must be a URI (error: %v)", err))
			}
		} else {
			usage("too many arguments.")
		}
		kc := setupKelClient(uri)
		var resourceGroups []*kel.ResourceGroup
		if err := kc.ResourceGroups.List(&resourceGroups).Do(); err != nil {
			fatal(fmt.Sprintf("failed to list resource groups (error: %v)", err))
		}
		for i := range resourceGroups {
			fmt.Println(resourceGroups[i].Name)
		}
	},
}
