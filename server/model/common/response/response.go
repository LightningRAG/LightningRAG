package response

import (
	"net/http"

	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

const (
	ERROR   = 7
	SUCCESS = 0
)

func Result(code int, data interface{}, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		code,
		data,
		msg,
	})
}

func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, i18n.Msg(c, "response.operation_success"), c)
}

func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, message, c)
}

func OkWithData(data interface{}, c *gin.Context) {
	Result(SUCCESS, data, i18n.Msg(c, "response.success"), c)
}

func OkWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

func Fail(c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, i18n.Msg(c, "response.operation_failed"), c)
}

func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, message, c)
}

// FailWithError returns a locale-aware message (response.failed_detail) with err.Error() as detail.
func FailWithError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if i18n.IsProviderAPIKeyError(err) {
		Result(ERROR, map[string]interface{}{}, i18n.Msg(c, "rag.error.model_api_key_rejected"), c)
		return
	}
	Result(ERROR, map[string]interface{}{}, i18n.Msgf(c, "response.failed_detail", err.Error()), c)
}

func NoAuth(message string, c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		7,
		nil,
		message,
	})
}

func FailWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(ERROR, data, message, c)
}
