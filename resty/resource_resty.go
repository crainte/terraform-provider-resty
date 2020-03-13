package resty

import (
    "fmt"
    "log"
    "time"
    "bytes"
    "strings"
    "net/http"
    "io/ioutil"
    "crypto/tls"
    "encoding/json"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceREST() *schema.Resource {
    return &schema.Resource{
        Create: restyRequest,
        Read:   restyRequest,
        Update: restyRequest,
        Delete: restyRequest,
        Exists: restyExists,

        Schema: map[string]*schema.Schema{
            "url": {
                Type: schema.TypeString,
                Description: "The request URL",
                Required: true,
                ForceNew: true,
            },
            "method": {
                Type: schema.TypeString,
                Description: "The http request verb",
                Default: "GET",
                Optional: true,
            },
            "headers": {
                Type: schema.TypeMap,
                Description: "Extra headers for the request",
                Optional: true,
                Sensitive: true,
            },
            "data": {
                Type: schema.TypeString,
                Description: "Data sent during the request",
                Optional: true,
                Sensitive: true,
            },

            "insecure": {
                Type: schema.TypeBool,
                Description: "Validate Certificate",
                Default: false,
                Optional: true,
            },
            "force_new": {
                Type:        schema.TypeList,
                Elem:        &schema.Schema{Type: schema.TypeString},
                Description: "Create a new instance if any of these items changes",
                Optional:    true,
                ForceNew:    true,
            },
            "id_field": {
                Type: schema.TypeString,
                Description: "Default ID field",
                Default: "id",
                Optional: true,
            },
            "timeout": {
                Type: schema.TypeInt,
                Description: "HTTP Timeout",
                Default: 10,
                Optional: true,
            },
            "username": {
                Type: schema.TypeString,
                Description: "Basic Auth Username",
                Optional: true,
                Sensitive: true,
            },
            "password": {
                Type: schema.TypeString,
                Description: "Basic Auth Password",
                Optional: true,
                Sensitive: true,
            },
            "debug": {
                Type: schema.TypeBool,
                Description: "Print Debug Information",
                Default: false,
                Optional: true,
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
                Type: schema.TypeString,
                Description: "Limit response conext by key",
                Optional: true,
            },

            "response": {
                Type: schema.TypeString,
                Description: "Response from the request",
                Computed: true,
            },
            "response_headers": {
                Type: schema.TypeMap,
                Description: "Response Headers from the request",
                Computed: true,
            },

        },
    }
}

func restyRequest(d *schema.ResourceData, meta interface{}) error {

    var req *http.Request
    var err error
    var id string
    var output = make(map[string]interface{})
    var response = make(map[string]interface{})
    var response_headers = make(map[string]interface{})

    url := d.Get("url").(string)
    method := d.Get("method").(string)
    headers := d.Get("headers").(map[string]interface{})
    data := d.Get("data").(string)
    username := d.Get("username").(string)
    password := d.Get("password").(string)
    debug := d.Get("debug").(bool)
    id_field := d.Get("id_field").(string)
    key := d.Get("key").(string)
    timeout := d.Get("timeout").(int)
    insecure := d.Get("insecure").(bool)
    filters := d.Get("filter").([]interface{})

    d.Set("id_field", id_field)
    d.Set("insecure", insecure)
    d.Set("timeout", timeout)

    transport := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
        Proxy: http.ProxyFromEnvironment,
    }

    client := &http.Client{
        Timeout: time.Second * time.Duration(timeout),
        Transport: transport,
    }

    buffer := bytes.NewBuffer([]byte(data))
    if data == "" {
        req, err = http.NewRequest(method, url, nil)
    } else {
        req, err = http.NewRequest(method, url, buffer)
        req.Header.Set("Content-Type", "application/json")
    }

    if err != nil {
        return fmt.Errorf("Error building request: %s", err)
    }

    if len(headers) > 0 {
        for k, v := range headers {
            req.Header.Set(k, v.(string))
        }
    }

    if username != "" && password != "" {
        req.SetBasicAuth(username, password)
    }

    if debug {
        log.Printf("[RESTY] Request headers:\n")
        for k, v := range req.Header {
            for _, h := range v {
                log.Printf("[RESTY]     %v: %v", k, h)
            }
        }

        log.Printf("[RESTY] Request Body:\n")
        body := "<none>"
        if req.Body != nil {
            body = string(data)
        }
        log.Printf("%s\n", body)
    }

    log.Printf("[RESTY] Making request")

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("Error making a request: %s", err)
    }

    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("HTTP request error. Response code: %d", resp.StatusCode)
    }

    response_body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("Error while reading response body. %s", err)
    }

    if debug {
        log.Printf("[RESTY] Response Body:\n%s\n", string(response_body))
    }

    for k, v := range resp.Header {
        // Concatenate according to RFC2616
        // cf. https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2
        response_headers[k] = strings.Join(v, ", ")
    }

    d.Set("response_headers", response_headers)

    if string(response_body) != "" {
        err := json.Unmarshal([]byte(response_body), &response)
        if err != nil {
            log.Printf("[RESTY] Non-Fatal error parsing body as JSON")
            d.SetId(time.Now().UTC().String())
            d.Set("response", string(response_body))
        } else {
            if key != "" {
                if _, ok := response[key]; ok {
                    log.Printf("[RESTY] Key exists")
                    if tmp, ok := response[key].(map[string]interface{}); ok {
                        log.Printf("[RESTY] It points to a map")
                        output = tmp
                    } else if tmp, ok := response[key].([]interface{}); ok {
                        log.Printf("[RESTY] It points to a list")

                        var name string
                        var value string

                        if filters != nil {
                            log.Printf("[RESTY] Filter response requested")

                            // todo: clean this doesn't need to be a loop
                            for _, cf := range filters {
                                customFilterMap := cf.(map[string]interface{})
                                name = customFilterMap["name"].(string)
                                value = customFilterMap["value"].(string)
                            }
                            Done:
                                for _, parent := range tmp {
                                    for entry, child := range parent.(map[string]interface{}) {
                                        if entry == name && child == value {
                                            log.Printf("[RESTY] Found the item: %s", parent)
                                            output = parent.(map[string]interface{})
                                            break Done
                                        }
                                    }
                                }
                        } else {
                            output = response[key].(map[string]interface{})
                        }
                    }
                } else {
                    return fmt.Errorf("Response does not contain key: %s", key)
                }
            } else {
                log.Printf("[RESTY] No key requested")
                output = response
            }

            id, err = GetStringAtKey(output, id_field, debug)
            out, _ := json.Marshal(output)
            d.Set("response", string(out))

            if id != "" {
                d.SetId(id)
            } else {
                d.SetId(time.Now().UTC().String())
            }
        }
    }

    return nil
}

func restyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
    return true, nil
}
