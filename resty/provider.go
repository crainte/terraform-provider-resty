package resty

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"headers": &schema.Schema{
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Optional:    true,
				Description: "A map of headers to be used with every request",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"resty": dataSourceREST(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"resty": resourceREST(),
		},
		ConfigureFunc: configureProvider,
	}
}

type ParentClient struct {
	headers map[string]string
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	headers := make(map[string]string)

	if ok := d.Get("headers"); ok != nil {
		for k, v := range ok.(map[string]interface{}) {
			headers[k] = v.(string)
		}
	}

	return &ParentClient{
		headers: headers,
	}, nil
}
