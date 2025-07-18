package openapiv3

import (
	"fmt"

	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

func RegisterOpenApi3Spec(fizz *fizz.Fizz, openInfo *openapi.Info, servers []*openapi.Server, securityRequirement []*openapi.SecurityRequirement, securitySchemes map[string]*openapi.SecuritySchemeOrRef, path string, httpPort int) {

	if len(servers) == 0 {
		servers = []*openapi.Server{
			&openapi.Server{
				URL:         fmt.Sprintf("http://localhost:%v", httpPort),
				Description: "Generated server url",
			},
		}
	}
	fizz.Generator().SetServers(servers)
	fizz.Generator().SetSecurityRequirement(securityRequirement)
	fizz.Generator().SetSecuritySchemes(securitySchemes)

	// Create a new route that serve the OpenAPI spec.
	fizz.GET(path, nil, fizz.OpenAPI(openInfo, "json"))
}
