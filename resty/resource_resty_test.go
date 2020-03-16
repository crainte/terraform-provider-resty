package resty

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type testHttpMock struct {
    server *httptest.Server
}

const testResourceConfig = `
resource "resty" "test" {
  url = "%s/test"
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

const testResourceConfigHeaders = `
resource "resty" "test" {
  url = "%s/headers"
  method = "GET"
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

                    if ok := r["headers.Authorization"] == "ZGVhZDpiZWVmCg==" ; ! ok {
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
					w.WriteHeader(http.StatusForbidden)
				}
			} else if r.URL.Path == "/utf-8/meta_200.txt" {
				w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("1.0.0"))
			} else if r.URL.Path == "/utf-16/meta_200.txt" {
				w.Header().Set("Content-Type", "application/json; charset=UTF-16")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("\"1.0.0\""))
			} else if r.URL.Path == "/meta_404.txt" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)

	return &testHttpMock{
		server: Server,
	}
}
