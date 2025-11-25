package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

// RouterFactory bundles handlers required to build the HTTP router.
type RouterFactory struct {
	IdentityHandler *IdentityHandler
	WSServer        *socketio.Server
	TokenValidator  TokenValidator
	CatalogHandler  *CatalogHandler
}

// Build wires all HTTP routes for REST and WebSocket.
func (f *RouterFactory) Build() *gin.Engine {
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	if f.WSServer != nil {
		router.GET("/socket.io/*any", gin.WrapH(f.WSServer))
	}

	api := router.Group("/api/v1")
	if f.CatalogHandler != nil {
		cat := api.Group("/categories")
		{
			cat.GET("", f.CatalogHandler.ListCategories)
			adminCats := cat.Group("")
			if f.TokenValidator != nil {
				adminCats.Use(AuthMiddleware(f.TokenValidator), RoleMiddleware("admin"))
			}
			adminCats.POST("", f.CatalogHandler.CreateCategory)
			adminCats.PUT("/:id", f.CatalogHandler.UpdateCategory)
			adminCats.DELETE("/:id", f.CatalogHandler.DeleteCategory)
		}

		prod := api.Group("/products")
		{
			prod.GET("", f.CatalogHandler.ListProducts)
			prod.GET("/:id", f.CatalogHandler.GetProduct)

			adminProd := prod.Group("")
			if f.TokenValidator != nil {
				adminProd.Use(AuthMiddleware(f.TokenValidator), RoleMiddleware("admin"))
			}
			adminProd.POST("", f.CatalogHandler.CreateProduct)
			adminProd.PUT("/:id", f.CatalogHandler.UpdateProduct)
			adminProd.DELETE("/:id", f.CatalogHandler.DeleteProduct)
		}

		api.GET("/search", f.CatalogHandler.Search)
	}
	if f.IdentityHandler != nil {
		identityGroup := api.Group("/identity")
		identityGroup.POST("/users/client", f.IdentityHandler.RegisterClient)
		identityGroup.POST("/users", f.IdentityHandler.RegisterUser)
		identityGroup.POST("/verify", f.IdentityHandler.VerifyUser)
		identityGroup.POST("/login", f.IdentityHandler.Login)

		protected := identityGroup.Group("")
		if f.TokenValidator != nil {
			protected.Use(AuthMiddleware(f.TokenValidator))
		}

		protected.PUT("/users/:id", f.IdentityHandler.UpdateUser)

		adminProtected := protected.Group("")
		if f.TokenValidator != nil {
			adminProtected.Use(RoleMiddleware("admin"))
		}
		adminProtected.POST("/users/:id/block", f.IdentityHandler.BlockUser)
		adminProtected.PUT("/users/:id/role", f.IdentityHandler.UpdateUserRole)
	}

	// Swagger docs placeholder.
	router.GET("/docs/swagger/openapi.yaml", ServeSwaggerSpec)

	return router
}
