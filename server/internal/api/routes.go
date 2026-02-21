package api

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *WalletHandler, apiKeys map[string]bool) {
	router.Use(RequestLogger(), CORSMiddleware())
	router.GET("/health", handler.Health)

	v1 := router.Group("/api/v1")
	v1.Use(APIKeyAuth(apiKeys))
	{
		v1.POST("/wallet/create", handler.CreateWallet)
		v1.POST("/wallet/sign", handler.Sign)
		v1.GET("/wallet/:address", handler.GetWallet)
	}
}
