package response

import (
	"Multi-objectiveOptimization/interfaces/errorcode"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    errorcode.CODE_SUCCESS,
		"message": "success",
		"data":    data,
	})
}

func SendError(c *gin.Context, code int, message interface{}, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": message,
		"data":    data,
	})
}

func SendParamsError(c *gin.Context, data interface{}) {
	SendError(c, errorcode.CODE_PARAMS_INVALID, "缺少参数", data)
}

func Redirect(c *gin.Context, url string) {
	c.Redirect(http.StatusFound, url)
}
