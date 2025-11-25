package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServeSwaggerSpec exposes the OpenAPI spec for tooling/UI hosting.
func ServeSwaggerSpec(c *gin.Context) {
	// TODO: stream the spec from docs/swagger/openapi.yaml or embed at build time.
	c.JSON(http.StatusNotImplemented, gin.H{"error": "swagger spec handler not implemented yet"})
}
