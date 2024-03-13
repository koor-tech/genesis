package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/gateway/request"
	"github.com/koor-tech/genesis/internal/cluster"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/rabbitmq"
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

	clusterSvc := cluster.NewService(database.NewDB(), rabbitmq.NewClient())

	customer := models.Customer{
		ID:    uuid.New(),
		Name:  createClusterRequest.ClientName,
		Email: createClusterRequest.ClientEmail,
	}
	clusterState, err := clusterSvc.BuildCluster(c, &customer, uuid.MustParse("80be226b-8355-4dea-b41a-6e17ea37559a"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"cluster": clusterState})
}

func GetCluster(c *gin.Context) {
	clusterID := uuid.MustParse(c.Param("id"))
	clusterSvc := cluster.NewService(database.NewDB(), rabbitmq.NewClient())
	koorCluster, err := clusterSvc.GetCluster(context.Background(), clusterID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"cluster": koorCluster})
}

func DeleteCluster(c *gin.Context) {
	clusterID := uuid.MustParse(c.Param("id"))
	clusterSvc := cluster.NewService(database.NewDB(), rabbitmq.NewClient())
	if err := clusterSvc.DeleteCluster(context.Background(), clusterID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cluster": nil})
}

func ResumeCluster(c *gin.Context) {
	clusterID := uuid.MustParse(c.Param("id"))
	clusterSvc := cluster.NewService(database.NewDB(), rabbitmq.NewClient())
	if err := clusterSvc.ResumeCluster(context.Background(), clusterID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cluster": nil})
}
