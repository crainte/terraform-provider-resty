package main

import (
	"github.com/crainte/terraform-provider-resty/resty"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: resty.Provider})
}
