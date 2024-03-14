package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/gateway/request"
	"github.com/koor-tech/genesis/internal/cluster"
	"github.com/koor-tech/genesis/pkg/models"
	"go.uber.org/fx"
)

var (
	ErrorBadRequest = "some parameters missed"
)

type Cluster struct {
	clusterSvc *cluster.Service
}

type Params struct {
	fx.In

	ClusterSvc *cluster.Service
}

func NewCluster(p Params) *Cluster {
	return &Cluster{
		clusterSvc: p.ClusterSvc,
	}
}

func (h *Cluster) CreateCluster(c *gin.Context) {
	var createClusterRequest request.CreateClusterRequest
	if err := c.ShouldBindJSON(&createClusterRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrorBadRequest})
		return
	}

	customer := models.Customer{
		ID:    uuid.New(),
		Name:  createClusterRequest.ClientName,
		Email: createClusterRequest.ClientEmail,
	}
	clusterState, err := h.clusterSvc.BuildCluster(c, &customer, uuid.MustParse("80be226b-8355-4dea-b41a-6e17ea37559a"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"cluster": clusterState})
}

func (h *Cluster) GetCluster(c *gin.Context) {
	clusterID := uuid.MustParse(c.Param("id"))

	koorCluster, err := h.clusterSvc.GetCluster(context.Background(), clusterID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"cluster": koorCluster})
}

func (h *Cluster) DeleteCluster(c *gin.Context) {
	clusterID := uuid.MustParse(c.Param("id"))

	if err := h.clusterSvc.DeleteCluster(context.Background(), clusterID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cluster": nil})
}

func (h *Cluster) ResumeCluster(c *gin.Context) {
	clusterID := uuid.MustParse(c.Param("id"))
	if err := h.clusterSvc.ResumeCluster(context.Background(), clusterID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cluster": nil})
}
