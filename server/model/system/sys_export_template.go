// 自动生成模板SysExportTemplate
package system

import (
	"github.com/LightningRAG/LightningRAG/server/global"
)

// SysExportTemplate export template
type SysExportTemplate struct {
	global.LRAG_MODEL
	DBName       string         `json:"dbName" form:"dbName" gorm:"column:db_name;comment:Database name;"`                        // Database name
	Name         string         `json:"name" form:"name" gorm:"column:name;comment:Template name;"`                               // Template name
	TableName    string         `json:"tableName" form:"tableName" gorm:"column:table_name;comment:Table name;"`                  // Table name
	TemplateID   string         `json:"templateID" form:"templateID" gorm:"column:template_id;comment:Template ID;"`              // Template ID
	TemplateInfo string         `json:"templateInfo" form:"templateInfo" gorm:"column:template_info;type:text;"`                  // Column labels JSON
	SQL          string         `json:"sql" form:"sql" gorm:"column:sql;type:text;comment:Custom export SQL;"`                    // Custom export SQL
	ImportSQL    string         `json:"importSql" form:"importSql" gorm:"column:import_sql;type:text;comment:Custom import SQL;"` // Custom import SQL
	Limit        *int           `json:"limit" form:"limit" gorm:"column:limit;comment:Export row limit"`
	Order        string         `json:"order" form:"order" gorm:"column:order;comment:Sort order"`
	Conditions   []Condition    `json:"conditions" form:"conditions" gorm:"foreignKey:TemplateID;references:TemplateID;comment:Filter conditions"`
	JoinTemplate []JoinTemplate `json:"joinTemplate" form:"joinTemplate" gorm:"foreignKey:TemplateID;references:TemplateID;comment:Joins"`
}

type JoinTemplate struct {
	global.LRAG_MODEL
	TemplateID string `json:"templateID" form:"templateID" gorm:"column:template_id;comment:Template ID"`
	JOINS      string `json:"joins" form:"joins" gorm:"column:joins;comment:JOIN clause"`
	Table      string `json:"table" form:"table" gorm:"column:table;comment:Joined table"`
	ON         string `json:"on" form:"on" gorm:"column:on;comment:Join ON condition"`
}

func (JoinTemplate) TableName() string {
	return "sys_export_template_join"
}

type Condition struct {
	global.LRAG_MODEL
	TemplateID string `json:"templateID" form:"templateID" gorm:"column:template_id;comment:Template ID"`
	From       string `json:"from" form:"from" gorm:"column:from;comment:Query parameter key"`
	Column     string `json:"column" form:"column" gorm:"column:column;comment:Filter column"`
	Operator   string `json:"operator" form:"operator" gorm:"column:operator;comment:Operator"`
}

func (Condition) TableName() string {
	return "sys_export_template_condition"
}
