package resty

import (
        "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
        "github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() *schema.Provider {
        return &schema.Provider{
                Schema: map[string]*schema.Schema{},
                ResourcesMap: map[string]*schema.Resource{
                    "resty": resourceREST(),
                },
                DataSourcesMap: map[string]*schema.Resource{},
        }
}
