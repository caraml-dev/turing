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
			wantErrSubstr: `parameter "project_id" in path has an error: value NOT_NUMBER: an invalid integer: strconv.ParseFloat: parsing "NOT_NUMBER": invalid syntax`,
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
			wantErrSubstr: `request body has an error: doesn't match the schema: Error at "/environment_name": property "environment_name" is missing`,
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
			wantErrSubstr: `request body has an error: doesn't match the schema: Error at "/config/ensembler/type": value is not one of the allowed values`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openapi, err := NewOpenAPIValidation(
				"../../api/openapi.yaml",
				OpenAPIValidationOptions{
					IgnoreAuthentication: true,
					IgnoreServers:        true,
				},
			)
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

func TestNewOpenAPIValidation(t *testing.T) {
	tests := []struct {
		name            string
		openapiYamlFile string
		options         OpenAPIValidationOptions
	}{
		{
			name:            "default options",
			openapiYamlFile: "../../api/openapi.yaml",
			options:         OpenAPIValidationOptions{},
		},
		{
			name:            "ignore authentication and servers",
			openapiYamlFile: "../../api/openapi.yaml",
			options: OpenAPIValidationOptions{
				IgnoreAuthentication: true,
				IgnoreServers:        true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oapi, err := NewOpenAPIValidation(tt.openapiYamlFile, tt.options)
			if err != nil {
				t.Error(err)
			}
			if tt.name == "default options" {
				r, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/1/routers", nil)
				if err != nil {
					t.Error(err)
				}
				if err = oapi.Validate(r); err == nil {
					t.Error("Validate() want err")
				}
			}
			if tt.name == "ignore authentication and servers" {
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
