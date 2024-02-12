package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/gateway/request"
	"github.com/koor-tech/genesis/internal/cluster"
	"github.com/koor-tech/genesis/pkg/database"
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

	clusterSvc := cluster.NewKoorCluster(database.NewDB())
	err := clusterSvc.NewCluster(context.Background(), createClusterRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	clusterState, err := clusterSvc.BuildCluster(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(201, gin.H{"cluster": clusterState})
}

func GetCluster(c *gin.Context) {
	clusterID := uuid.MustParse(c.Param("id"))
	clusterSvc := cluster.NewKoorCluster(database.NewDB())
	koorCluster, err := clusterSvc.Cluster(context.Background(), clusterID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err})
		return
	}
	c.JSON(201, gin.H{"cluster": koorCluster})
}
