package main

import (
    "github.com/hashicorp/terraform-plugin-sdk/plugin"
    "github.com/crainte/terraform-provider-resty/resty"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        ProviderFunc: resty.Provider})
}
