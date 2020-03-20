package resty

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"resty": dataSourceREST(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"resty": resourceREST(),
		},
	}
}
