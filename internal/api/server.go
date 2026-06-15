package api

import (
	"os"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/api/routes"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/gin-gonic/gin"
)

func Server(app *global.App) error {
	e := gin.Default()
	// e.Use(cors.New(cors.Config{
	// 	AllowOrigins:     strings.Split(os.Getenv("ADMIN_URLS", ",")),
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// 	MaxAge:           10 * time.Second,
	// }))

	e.Use(gin.Logger())
	e.Use(gin.Recovery())

	e.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// routes.RegisterRoutes(e, app)
	routes.RegisterRoutes(e)

	return e.Run(":" + os.Getenv("LISTEN_ADDR"))
}
