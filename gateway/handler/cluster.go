package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/koor-tech/genesis/gateway/request"
	clusterService "github.com/koor-tech/genesis/internal/cluster"
	"github.com/koor-tech/genesis/pkg/models"
	"net/http"
)

var (
	ErrorBadRequest = "some parameters missed"
)

func CreateCluster(c *gin.Context) {
	var createClusterRequest request.CreateClusterRequest
	if err := c.ShouldBindJSON(&createClusterRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrorBadRequest})
		return
	}

	clusterSvc := clusterService.NewKoorCluster(createClusterRequest.Provider, models.NewClient(createClusterRequest.ClientName))
	err := clusterSvc.BuildCluster(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrorBadRequest})
		return
	}
	c.JSON(201, gin.H{"cluster": clusterSvc.Cluster()})
}
