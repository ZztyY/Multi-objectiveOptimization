package interfaces

import (
	"Multi-objectiveOptimization/interfaces/algorithm"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {

	group := r.Group("/moea")

	group.POST("/execute", algorithm.Execute)
}
