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
			fmt.Fprintf(os.Stderr, "Usage: kel resource-groups create [name]\n")
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
			fmt.Fprintf(os.Stderr, "Usage: kel resource-groups list\n")
			fatal(msg)
		}
		if len(args) > 0 {
			usage("too many arguments")
		}
		uri, err := LookupURI()
		if err != nil {
			usage(err.Error())
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
