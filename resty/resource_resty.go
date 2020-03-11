package resty

import (
    "fmt"
    "log"
    "time"
    "bytes"
    "net/http"
    "encoding/json"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceREST() *schema.Resource {
    return &schema.Resource{
        Create: restyRequest
        Read:   restyRequest,
        Update: restyRequest,
        Delete: restyRequest,
        Exists: restyExists,

        Schema: map[string]*schema.Schema{
            "url": &schema.Schema(
                Type: schema.TypeString,
                Description: "The request URL",
                Required: true,
                ForceNew: true,
            ),
            "method": &schema.Schema(
                Type: schema.TypeString,
                Description: "The http request verb",
                Default: "GET",
                Optional: true,
            ),
            "headers": &schema.Schema(
                Type: schema.TypeMap,
                Description: "Extra headers for the request",
                Optional: true,
                Sensitive: true
            ),
            "data": &schema.Schema(
                Type: schema.TypeString,
                Description: "Data sent during the request",
                Optional: true,
                Sensitive: true
            ),

            "insecure": &schema.Schema(
                Type: schema.TypeBool,
                Description: "Validate Certificate",
                Default: true,
                Optional: true
            ),
            "force_new": &schema.Schema{
                Type:        schema.TypeList,
                Elem:        &schema.Schema{Type: schema.TypeString},
                Description: "Create a new instance if any of these items changes",
                Optional:    true,
                ForceNew:    true,
            },
            "id_field": &schema.Schema(
                Type: schema.TypeString,
                Description: "Default ID field",
                Default: "id",
                Optional: true
            ),
            "timeout": &schema.Schema(
                Type: schema.TypeInt,
                Description: "HTTP Timeout",
                Default: 10,
                Optional: true
            ),
            "username": &schema.Schema(
                Type: schema.TypeString,
                Description: "Basic Auth Username",
                Optional: true,
                Sensitive: true
            ),
            "password": &schema.Schema(
                Type: schema.TypeString,
                Description: "Basic Auth Password",
                Optional: true,
                Sensitive: true
            ),
            "debug": &schema.Schema(
                Type: schema.TypeBool,
                Description: "Print Debug Information",
                Default: false,
                Optional: true
            ),

            "search_key": &schema.Schema(
                Type: schema.TypeString,
                Description: "Search the results for a key",
                Optional: true
            ),
            "search_value": &schema.Schema(
                Type: schema.TypeString,
                Description: "Collect the results from key[value]",
                Optional: true
            ),

            "response": &schema.Schema(
                Type: schema.TypeString,
                Description: "Response from the request",
                Computed: true
            ),

        }
    }
}

func restyRequest(d *schema.ResourceData, meta interface{}) error {

    url := d.Get("url").(string)
    method := d.Get("method").(string)
    headers := d.Get("headers").(map[string]interface{})
    data := d.Get("data").(string)
    username := d.Get("username").(string)
    password := d.Get("password").(string)
    debug := d.Get("debug").(bool)
    id_field := d.Get("id_field").(string)

    transport := &http.Transport{
        TLSClientConfig: &tls.Config(InsecureSkipVerify: d.Get("insecure").(bool),
        Proxy: http.ProxyFromEnvironment,
    }

    client := &http.Client{
        Timeout: time.Second * time.Duration(d.Get("timeout").(int)),
        Transport: transport,
    }

    buffer := bytes.NewBuffer([]byte(data))
    if data == "" {
        req, err := http.NewRequest(method, url, nil)
    } else {
        req, err := http.NewRequest(method, url, buffer)
        req.Header.Set("Content-Type", "application/json")
    }

    if err != nil {
        return fmt.Errorf("Error building request: %s", err)
    }

    if len(headers) > 0 {
        for k, v := range client.headers {
            req.Header.Set(k, v)
        }
    }

    if username != "" && password != "" {
        req.SetBasicAuth(username, password)
    }

    if debug {
        log.Printf("RESTY: Request headers:\n")
        for k, v := range req.Header {
            for _, h := range v {
                log.Printf("api_client.go:   %v: %v", k, h)
            }
        }

        log.Printf("RESTY: Request Body:\n")
        body := "<none>"
        if req.Body != nil {
            body = string(data)
        }
        log.Printf("%s\n", body)
    }

    resp, err := client.http_client.Do(req)
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
        log.Printf("RESTY: Response Body:\n%s\n", response_body)
    }

    response_headers := make(map[string]string)
    for k, v := range resp.Header {
        // Concatenate according to RFC2616
        // cf. https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2
        response_headers[k] = strings.Join(v, ", ")
    }

    d.Set("response", string(response_body))
    if err = d.Set("response_headers", response_headers); err != nil {
        return fmt.Errorf("Error setting HTTP Response Headers: %s", err)
    }

    if response_body != "" {
        err := json.Unmarshal([]byte(response_body), map[string]interface{})
        if err != nil {
            log.Printf("RESTY: Non-Fatal error parsing body as JSON")
            d.SetId(time.Now().UTC().String())
        } else {
            id, err := GetStringAtKey(response_body, id_field, debug)
            if err == nil {
                d.SetId(id)
            } else {
                d.SetId(time.Now().UTC().String())
            }
        }
    }

    return nil
}

func restyExists(d *schema.ResourceData, meta interface{}) (bool, error) {}
