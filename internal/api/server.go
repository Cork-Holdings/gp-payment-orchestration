package api

import (
	"os"
	"strings"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/api/routes"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/m_api"
	"github.com/gin-gonic/gin"
)

func configureTrustedProxies(e *gin.Engine) {
	raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if raw == "" {
		_ = e.SetTrustedProxies(nil)
		return
	}

	proxies := make([]string, 0)
	for _, proxy := range strings.Split(raw, ",") {
		proxy = strings.TrimSpace(proxy)
		if proxy != "" {
			proxies = append(proxies, proxy)
		}
	}

	if len(proxies) == 0 {
		_ = e.SetTrustedProxies(nil)
		return
	}

	_ = e.SetTrustedProxies(proxies)
}

func Server(app *global.App) error {
	e := gin.Default()
	configureTrustedProxies(e)
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

	m_api.RegisterMerchantRoutes(e, app)

	return e.Run(":" + os.Getenv("LISTEN_ADDR"))
}
