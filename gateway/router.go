package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/koor-tech/genesis/gateway/handler"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/status", handler.GetStatus)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/cluster", handler.CreateCluster)
		v1.GET("/clusters/:id", handler.GetCluster)
	}
	return r
}
