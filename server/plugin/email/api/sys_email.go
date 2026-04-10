package api

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	email_response "github.com/LightningRAG/LightningRAG/server/plugin/email/model/response"
	"github.com/LightningRAG/LightningRAG/server/plugin/email/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EmailApi struct{}

// EmailTest
// @Tags      System
// @Summary   发送测试邮件
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200  {string}  string  "{"success":true,"data":{},"msg":"发送成功"}"
// @Router    /email/emailTest [post]
func (s *EmailApi) EmailTest(c *gin.Context) {
	err := service.ServiceGroupApp.EmailTest()
	if err != nil {
		global.LRAG_LOG.Error("发送失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "plugin.email.send_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "plugin.email.send_success"), c)
}

// SendEmail
// @Tags      System
// @Summary   发送邮件
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  body      email_response.Email  true  "发送邮件必须的参数"
// @Success   200   {string}  string                "{"success":true,"data":{},"msg":"发送成功"}"
// @Router    /email/sendEmail [post]
func (s *EmailApi) SendEmail(c *gin.Context) {
	var email email_response.Email
	err := c.ShouldBindJSON(&email)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	err = service.ServiceGroupApp.SendEmail(email.To, email.Subject, email.Body)
	if err != nil {
		global.LRAG_LOG.Error("发送失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "plugin.email.send_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "plugin.email.send_success"), c)
}
