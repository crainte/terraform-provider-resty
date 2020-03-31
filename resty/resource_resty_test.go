package resty

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type testHttpMock struct {
	server *httptest.Server
}

const testResourceConfig = `
resource "resty" "test" {
  url    = "%s/test"
  method = "GET"
}
`

func TestResourceGet(t *testing.T) {
	mock := initMockHttpServer()

	defer mock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfig, mock.server.URL),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["resty.test"]
					if !ok {
						return fmt.Errorf("missing resty resource 'test'")
					}

					r := s.RootModule().Resources["resty.test"].Primary.Attributes

					if r["response"] != "{}" {
						return fmt.Errorf(
							`'response' output is %s; want '{}'`,
							r["response"],
						)
					}

					if r["response_headers.X-Is-Teapot"] != "Yes" {
						return fmt.Errorf(
							`'X-Is-Teapot' response header is %s; want 'Yes'`,
							r["response_headers"],
						)
					}

					return nil
				},
			},
		},
	})
}

const testResourceConfig_404 = `
resource "resty" "test" {
  url    = "%s/nope"
  method = "GET"
}
`

func TestResourceGet_404(t *testing.T) {
	mock := initMockHttpServer()

	defer mock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testResourceConfig_404, mock.server.URL),
				ExpectError: regexp.MustCompile("HTTP request error. Response code: 404"),
			},
		},
	})
}

const testResourceConfigHeaders = `
resource "resty" "test" {
  url     = "%s/headers"
  method  = "GET"
  headers = {
    "Authorization" = "ZGVhZDpiZWVmCg=="
  }
}
`

func TestResourceGet_withHeaders(t *testing.T) {
	mock := initMockHttpServer()

	defer mock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfigHeaders, mock.server.URL),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["resty.test"]
					if !ok {
						return fmt.Errorf("missing resty resource 'test'")
					}

					r := s.RootModule().Resources["resty.test"].Primary.Attributes

					if r["response"] != "{}" {
						return fmt.Errorf(
							`'response' output is %s; want '{}'`,
							r["response"],
						)
					}

					if ok := r["headers.Authorization"] == "ZGVhZDpiZWVmCg=="; !ok {
						return fmt.Errorf(
							"'headers.Authorization' not stored correctly in state",
						)
					}

					return nil
				},
			},
		},
	})
}

const testResourceConfigHeaders_401 = `
resource "resty" "test" {
  url     = "%s/headers"
  method  = "GET"
  headers = {
    "Authorization" = "nope"
  }
}
`

func TestResourceGet_withBadHeaders(t *testing.T) {
	mock := initMockHttpServer()

	defer mock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testResourceConfigHeaders_401, mock.server.URL),
				ExpectError: regexp.MustCompile("HTTP request error. Response code: 401"),
			},
		},
	})
}

const testResourceConfigFilter = `
resource "resty" "test" {
  url     = "%s/filter"
  method  = "GET"
  headers = {
    "Authorization" = "ZGVhZDpiZWVmCg=="
  }
  key = "content"
  filter {
    name  = "interesting"
    value = "value"
  }
}
`

func TestResourceGet_withFilter(t *testing.T) {
	mock := initMockHttpServer()

	defer mock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResourceConfigFilter, mock.server.URL),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["resty.test"]
					if !ok {
						return fmt.Errorf("missing resty resource 'test'")
					}

					r := s.RootModule().Resources["resty.test"].Primary.Attributes

					m := make(map[string]interface{})
					if err := json.Unmarshal([]byte(r["response"]), &m); err != nil {
						return fmt.Errorf("Unable to parse JSON")
					}

					if m["working"].(string) != "yes" {
						return fmt.Errorf(
							`'response' output is %s; want '{"working": "yes"}'`,
							r["response"],
						)
					}

					return nil
				},
			},
		},
	})
}

const testResourceConfigMissingFilter = `
resource "resty" "test" {
  url     = "%s/filter"
  method  = "GET"
  headers = {
    "Authorization" = "ZGVhZDpiZWVmCg=="
  }
  key = "content"
  filter {
    name  = "interesting"
    value = "missing"
  }
}
`

func TestResourceGet_withMissingFilter(t *testing.T) {
	mock := initMockHttpServer()

	defer mock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testResourceConfigMissingFilter, mock.server.URL),
				ExpectError: regexp.MustCompile("Response no filter match for: interesting = missing"),
			},
		},
	})
}

func initMockHttpServer() *testHttpMock {
	Server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "application/json")
			w.Header().Add("X-Is-Teapot", "Yes")
			if r.URL.Path == "/test" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{}"))
			} else if r.URL.Path == "/headers" {
				if r.Header.Get("Authorization") == "ZGVhZDpiZWVmCg==" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("{}"))
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
			} else if r.URL.Path == "/filter" {
				if r.Header.Get("Authorization") == "ZGVhZDpiZWVmCg==" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("{\"content\": [{\"interesting\": \"no\", \"you\": \"failed\"},{\"interesting\": \"value\", \"working\": \"yes\"}]}"))
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)

	return &testHttpMock{
		server: Server,
	}
}
