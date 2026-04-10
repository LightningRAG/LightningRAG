package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/i18n"
	"github.com/LightningRAG/LightningRAG/server/model/common/response"
	"github.com/LightningRAG/LightningRAG/server/model/system/request"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type DBApi struct{}

// InitDB
// @Tags     InitDB
// @Summary  初始化用户数据库
// @Produce  application/json
// @Param    data  body      request.InitDB                  true  "初始化数据库参数"
// @Success  200   {object}  response.Response{data=string}  "初始化用户数据库"
// @Router   /init/initdb [post]
func (i *DBApi) InitDB(c *gin.Context) {
	if global.LRAG_DB != nil {
		global.LRAG_LOG.Error("已存在数据库配置!")
		response.FailWithMessage(i18n.Msg(c, "sys.initdb.config_exists"), c)
		return
	}
	var dbInfo request.InitDB
	if err := c.ShouldBindJSON(&dbInfo); err != nil {
		global.LRAG_LOG.Error("参数校验不通过!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "sys.initdb.param_invalid"), c)
		return
	}
	if err := initDBService.InitDB(dbInfo); err != nil {
		global.LRAG_LOG.Error("自动创建数据库失败!", zap.Error(err))
		response.FailWithMessage(i18n.Msg(c, "sys.initdb.auto_create_failed"), c)
		return
	}
	response.OkWithMessage(i18n.Msg(c, "sys.initdb.auto_create_success"), c)
}

// CheckDB
// @Tags     CheckDB
// @Summary  初始化用户数据库
// @Produce  application/json
// @Success  200  {object}  response.Response{data=map[string]interface{},msg=string}  "初始化用户数据库"
// @Router   /init/checkdb [post]
func (i *DBApi) CheckDB(c *gin.Context) {
	var (
		message  = "前往初始化数据库"
		needInit = true
	)

	if global.LRAG_DB != nil {
		message = "数据库无需初始化"
		needInit = false
	}
	//global.LRAG_LOG.Info(message)
	response.OkWithDetailed(gin.H{"needInit": needInit}, message, c)
}
