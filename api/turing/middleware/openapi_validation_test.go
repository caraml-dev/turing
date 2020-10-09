package middleware

import (
	"net/http"
	"strings"
	"testing"
)

func TestOpenAPIValidationValidate(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		url           string
		body          string
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name:   "valid request",
			method: "POST",
			url:    "/projects/1/routers",
			body: `
{
  "environment_name": "myenv",
  "name": "myrouter",
  "config": {
    "routes": [
      {
        "id": "myroute",
        "type": "PROXY",
        "endpoint": "http://example.com",
        "timeout": "5s"
      }
    ],
    "default_route_id": "myroute",
	"log_config": {
      "result_logger_type": "nop"
    },
	"experiment_engine": {
		"type": "nop"
	},
    "resource_request": {
      "min_replica": 0,
      "max_replica": 0
    },
    "timeout": "5s"
  }
}
`,
			wantErr: false,
		},
		{
			name:          "bad path parameter",
			method:        "GET",
			url:           "/projects/NOT_NUMBER/routers",
			wantErr:       true,
			wantErrSubstr: "Parameter 'project_id' in path has an error: value NOT_NUMBER: an invalid integer",
		},
		{
			name:   "Missing property in body",
			method: "POST",
			url:    "/projects/1/routers",
			body: `
{
  "name": "myrouter",
  "config": {
    "routes": [
      {
        "id": "myroute",
        "type": "PROXY",
        "endpoint": "http://example.com",
        "timeout": "5s"
      }
    ],
    "default_route_id": "myroute",
	"log_config": {
      "result_logger_type": "nop"
    },
	"experiment_engine": {
		"type": "nop"
	},
    "resource_request": {
      "min_replica": 0,
      "max_replica": 0
    },
    "timeout": "5s"
  }
}
`,
			wantErr:       true,
			wantErrSubstr: `Error at "/environment_name":Property 'environment_name' is missing`,
		},
		{
			name:   "invalid enum",
			method: "POST",
			url:    "/projects/1/routers",
			body: `
{
  "environment_name": "myenv",
  "name": "myrouter",
  "config": {
    "routes": [
      {
        "id": "myroute",
        "type": "PROXY",
        "endpoint": "http://example.com",
        "timeout": "5s"
      }
    ],
    "default_route_id": "myroute",
    "experiment_engine": {
      "type": "nop"
    },
    "log_config": {
      "result_logger_type": "nop"
    },
    "resource_request": {
      "min_replica": 0,
      "max_replica": 0
    },
    "timeout": "5s",
    "ensembler": {
      "type": "invalidtype"
    }
  }
}
`,
			wantErr:       true,
			wantErrSubstr: `Error at "/config/ensembler/type":JSON value is not one of the allowed values`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openapi, err := NewOpenAPIV2Validation("../../swagger.yaml", OpenAPIValidationOptions{
				IgnoreAuthentication: true,
				IgnoreServers:        true,
			})
			if err != nil {
				t.Error(err)
			}
			req, err := http.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			if err != nil {
				t.Error(err)
			}
			req.Header.Set("Content-Type", "application/json")

			err = openapi.Validate(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErrSubstr != "" && !strings.Contains(err.Error(), tt.wantErrSubstr) {
				t.Errorf("Validate() error = %v, wantErrSubstr= %v", err.Error(), tt.wantErrSubstr)
			}
		})
	}
}

func TestNewOpenAPIV2Validation(t *testing.T) {
	tests := []struct {
		name            string
		swaggerYamlFile string
		options         OpenAPIValidationOptions
	}{
		{
			name:            "default options",
			swaggerYamlFile: "../../swagger.yaml",
			options:         OpenAPIValidationOptions{},
		},
		{
			name:            "ignore authentication and servers",
			swaggerYamlFile: "../../swagger.yaml",
			options: OpenAPIValidationOptions{
				IgnoreAuthentication: true,
				IgnoreServers:        true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oapi, err := NewOpenAPIV2Validation(tt.swaggerYamlFile, tt.options)
			if err != nil {
				t.Error(err)
			}
			if tt.name == "default options" {
				if len(oapi.swagger.Servers) < 1 {
					t.Errorf("server len got: %d, want: >= 1", len(oapi.swagger.Servers))
				}
				r, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/1/routers", nil)
				if err != nil {
					t.Error(err)
				}
				if err = oapi.Validate(r); err == nil {
					t.Error("Validate() want err")
				}
			}
			if tt.name == "ignore authentication and servers" {
				if len(oapi.swagger.Servers) > 0 {
					t.Error("server len want 0")
				}
				r, err := http.NewRequest("GET", "/projects/1/routers", nil)
				if err != nil {
					t.Error(err)
				}
				if err = oapi.Validate(r); err != nil {
					t.Error("Validate() do not want err")
				}
			}
		})
	}
}
