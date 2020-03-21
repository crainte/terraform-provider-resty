package resty

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceREST() *schema.Resource {
	return &schema.Resource{
		Read: restyRequest,

		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Description: "The request URL",
				Required:    true,
				ForceNew:    true,
			},
			"method": {
				Type:        schema.TypeString,
				Description: "The http request verb",
				Default:     "GET",
				Optional:    true,
			},
			"headers": {
				Type:        schema.TypeMap,
				Description: "Extra headers for the request",
				Optional:    true,
				Sensitive:   true,
			},
			"data": {
				Type:        schema.TypeString,
				Description: "Data sent during the request",
				Optional:    true,
				Sensitive:   true,
			},

			"insecure": {
				Type:        schema.TypeBool,
				Description: "Validate Certificate",
				Default:     false,
				Optional:    true,
			},
			"force_new": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Create a new instance if any of these items changes",
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
			},
			"id_field": {
				Type:        schema.TypeString,
				Description: "Default ID field",
				Default:     "id",
				Optional:    true,
			},
			"timeout": {
				Type:        schema.TypeInt,
				Description: "HTTP Timeout",
				Default:     10,
				Optional:    true,
			},
			"username": {
				Type:        schema.TypeString,
				Description: "Basic Auth Username",
				Optional:    true,
				Sensitive:   true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "Basic Auth Password",
				Optional:    true,
				Sensitive:   true,
			},
			"debug": {
				Type:        schema.TypeBool,
				Description: "Print Debug Information",
				Default:     false,
				Optional:    true,
			},

			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"key": {
				Type:        schema.TypeString,
				Description: "Limit response conext by key",
				Optional:    true,
			},

			"response": {
				Type:        schema.TypeString,
				Description: "Response from the request",
				Computed:    true,
			},
			"response_headers": {
				Type:        schema.TypeMap,
				Description: "Response Headers from the request",
				Computed:    true,
			},
		},
	}
}

// reuse the request function from resource
