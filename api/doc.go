// Package api embeds the OpenAPI description of the HDF5 Agent HTTP API.
package api

import _ "embed"

// OpenAPIYAML is the OpenAPI 3 document for /api/v1.
//
//go:embed openapi.yaml
var OpenAPIYAML []byte
