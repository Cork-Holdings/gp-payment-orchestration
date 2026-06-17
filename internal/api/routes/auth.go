package routes

import (
	"context"
	"net/http"
	"os"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/proto/authpb"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func RegisterAuthRoutes(e *gin.Engine, app *global.App) {
	// Establish gRPC connection to Auth Service
	addr := os.Getenv("AUTH_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50052"
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to auth service: " + err.Error())
	}

	client := authpb.NewAuthServiceClient(conn)

	e.POST("/auth/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := client.Login(context.Background(), &authpb.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	e.POST("/auth/refresh", func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := client.Refresh(context.Background(), &authpb.RefreshRequest{
			RefreshToken: req.RefreshToken,
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	e.POST("/auth/keys", func(c *gin.Context) {
		res, err := client.GetKeys(context.Background(), &authpb.KeysRequest{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	e.GET("/auth/roles", func(c *gin.Context) {
		res, err := client.GetRoles(context.Background(), &authpb.RolesRequest{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	e.GET("/auth/permissions", func(c *gin.Context) {
		role := c.Query("role")
		if role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role query parameter is required"})
			return
		}

		res, err := client.GetPermissions(context.Background(), &authpb.PermissionsRequest{
			Role: role,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})
}
