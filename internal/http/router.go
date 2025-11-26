package http

import (
	"net/http"

	"catalog-api/internal/ws"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouterFactory agrupa los handlers necesarios para construir el router HTTP.
type RouterFactory struct {
	IdentityHandler *IdentityHandler
	WSHub           *ws.Hub
	TokenValidator  TokenValidator
	CatalogHandler  *CatalogHandler
}

// Build cablea todas las rutas HTTP para REST y WebSocket.
func (f *RouterFactory) Build() *gin.Engine {
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	if f.WSHub != nil {
		router.GET("/ws", func(c *gin.Context) {
			if f.TokenValidator == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token validator"})
				return
			}
			token := websocketToken(c)
			if token == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
				return
			}
			if _, err := f.TokenValidator.Validate(token); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			f.WSHub.ServeHTTP(c.Writer, c.Request)
		})
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
			prod.GET("/:id/history", f.CatalogHandler.GetProductHistory)

			adminProd := prod.Group("")
			if f.TokenValidator != nil {
				adminProd.Use(AuthMiddleware(f.TokenValidator), RoleMiddleware("admin"))
			}
			adminProd.POST("", f.CatalogHandler.CreateProduct)
			adminProd.PUT("/:id", f.CatalogHandler.UpdateProduct)
			adminProd.DELETE("/:id", f.CatalogHandler.DeleteProduct)
			adminProd.POST("/:id/categories/:categoryId", f.CatalogHandler.AddProductCategory)
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

		protected.PUT("/users/me", f.IdentityHandler.UpdateUser)

		adminProtected := protected.Group("")
		if f.TokenValidator != nil {
			adminProtected.Use(RoleMiddleware("admin"))
		}
		adminProtected.POST("/users/:id/block", f.IdentityHandler.BlockUser)
		adminProtected.PUT("/users/:id/role", f.IdentityHandler.UpdateUserRole)
	}

	api.GET("/events", EventsCatalog)

	// Sirve el spec de swagger desde archivo local para evitar builds desactualizados.
	router.GET("/swagger/doc.json", func(c *gin.Context) {
		c.File("docs/swagger/swagger.json")
	})
	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	// Expone el diagrama ER para consulta fuera del wildcard de swagger.
	router.StaticFile("/db-schema.puml", "docs/db-schema.puml")
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL, ginSwagger.InstanceName("swagger")))
	router.StaticFile("/events-ui", "web/events.html")

	return router
}

func websocketToken(c *gin.Context) string {
	if token := c.Query("token"); token != "" {
		return token
	}
	if token := bearerTokenFromHeader(c.Request); token != "" {
		return token
	}
	if token, err := c.Cookie("token"); err == nil && token != "" {
		return token
	}
	return ""
}
