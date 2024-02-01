package request

type CreateClusterRequest struct {
	Provider   string `form:"provider" json:"provider" binding:"required"`
	ClientName string `form:"client" json:"client" binding:"required"`
}
