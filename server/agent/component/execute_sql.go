package component

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"
)

func init() {
	Register("ExecuteSQL", NewExecuteSQL)
}

// ExecuteSQL SQL 执行组件，支持 MySQL/PostgreSQL/MariaDB/MSSQL
type ExecuteSQL struct {
	id     string
	canvas Canvas
	params map[string]any
	output map[string]any
	err    string
	mu     sync.RWMutex
}

// NewExecuteSQL 创建 ExecuteSQL 组件
func NewExecuteSQL(canvas Canvas, id string, params map[string]any) (Component, error) {
	return &ExecuteSQL{
		id:     id,
		canvas: canvas,
		params: params,
		output: make(map[string]any),
	}, nil
}

// ComponentName 返回组件名
func (e *ExecuteSQL) ComponentName() string {
	return "ExecuteSQL"
}

// Invoke 执行 SQL 查询
func (e *ExecuteSQL) Invoke(inputs map[string]any) error {
	e.mu.Lock()
	e.err = ""
	e.mu.Unlock()

	sqlStr := e.canvas.ResolveString(getStrParam(e.params, "sql"))
	if sqlStr == "" {
		sqlStr = getStrParam(e.params, "sql")
	}
	if strings.TrimSpace(sqlStr) == "" {
		e.mu.Lock()
		e.err = "ExecuteSQL sql 为空"
		e.mu.Unlock()
		return fmt.Errorf("ExecuteSQL sql 为空")
	}

	dbType := strings.ToLower(getStrParam(e.params, "db_type"))
	if dbType == "" {
		dbType = "mysql"
	}
	host := e.canvas.ResolveString(getStrParam(e.params, "host"))
	if host == "" {
		host = getStrParam(e.params, "host")
	}
	port := getIntParam(e.params, "port", 3306)
	username := e.canvas.ResolveString(getStrParam(e.params, "username"))
	if username == "" {
		username = getStrParam(e.params, "username")
	}
	password := e.canvas.ResolveString(getStrParam(e.params, "password"))
	if password == "" {
		password = getStrParam(e.params, "password")
	}
	database := e.canvas.ResolveString(getStrParam(e.params, "database"))
	if database == "" {
		database = getStrParam(e.params, "database")
	}
	maxRecords := getIntParam(e.params, "max_records", 1024)
	if maxRecords <= 0 {
		maxRecords = 1024
	}

	dsn, err := e.buildDSN(dbType, host, port, username, password, database)
	if err != nil {
		e.mu.Lock()
		e.err = err.Error()
		e.mu.Unlock()
		return err
	}

	db, err := sql.Open(e.driverName(dbType), dsn)
	if err != nil {
		e.mu.Lock()
		e.err = err.Error()
		e.mu.Unlock()
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		e.mu.Lock()
		e.err = "数据库连接失败: " + err.Error()
		e.mu.Unlock()
		return err
	}

	// 限制返回行数
	limitSQL := sqlStr
	if !strings.Contains(strings.ToUpper(sqlStr), "LIMIT") && dbType != "mssql" {
		limitSQL = sqlStr + fmt.Sprintf(" LIMIT %d", maxRecords)
	}
	if dbType == "mssql" && !strings.Contains(strings.ToUpper(sqlStr), "TOP") {
		// MSSQL 使用 TOP，需在 SELECT 后插入
		limitSQL = e.injectMSSQLLimit(sqlStr, maxRecords)
	}

	rows, err := db.Query(limitSQL)
	if err != nil {
		e.mu.Lock()
		e.err = err.Error()
		e.mu.Unlock()
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		e.mu.Lock()
		e.err = err.Error()
		e.mu.Unlock()
		return err
	}

	var result []map[string]any
	count := 0
	for rows.Next() && count < maxRecords {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			e.mu.Lock()
			e.err = err.Error()
			e.mu.Unlock()
			return err
		}
		row := make(map[string]any)
		for i, col := range columns {
			v := values[i]
			if v == nil {
				row[col] = nil
			} else {
				switch val := v.(type) {
				case []byte:
					row[col] = string(val)
				default:
					row[col] = val
				}
			}
		}
		result = append(result, row)
		count++
	}

	// formalized_content: 表格式字符串，便于 Message 展示
	formalized := e.formatAsTable(columns, result)
	jsonBytes, _ := json.Marshal(result)

	e.mu.Lock()
	e.output["json"] = result
	e.output["formalized_content"] = formalized
	e.output["json_string"] = string(jsonBytes)
	e.mu.Unlock()
	return nil
}

func (e *ExecuteSQL) buildDSN(dbType, host string, port int, username, password, database string) (string, error) {
	switch dbType {
	case "mysql", "mariadb":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
			username, password, host, port, database), nil
	case "postgres", "postgresql":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			username, password, host, port, database), nil
	case "mssql", "sqlserver":
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			username, password, host, port, database), nil
	default:
		return "", fmt.Errorf("不支持的数据库类型: %s，支持: mysql, mariadb, postgres, postgresql, mssql, sqlserver", dbType)
	}
}

func (e *ExecuteSQL) driverName(dbType string) string {
	switch dbType {
	case "mysql", "mariadb":
		return "mysql"
	case "postgres", "postgresql":
		return "pgx"
	case "mssql", "sqlserver":
		return "sqlserver"
	default:
		return "mysql"
	}
}

func (e *ExecuteSQL) injectMSSQLLimit(sqlStr string, limit int) string {
	upper := strings.ToUpper(strings.TrimSpace(sqlStr))
	if strings.HasPrefix(upper, "SELECT ") {
		return "SELECT TOP " + fmt.Sprintf("%d", limit) + " " + sqlStr[7:]
	}
	return sqlStr
}

func (e *ExecuteSQL) formatAsTable(columns []string, rows []map[string]any) string {
	if len(rows) == 0 {
		return "无查询结果"
	}
	var sb strings.Builder
	// 表头
	for i, col := range columns {
		if i > 0 {
			sb.WriteString(" | ")
		}
		sb.WriteString(col)
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", len(columns)*12))
	sb.WriteString("\n")
	for _, row := range rows {
		for i, col := range columns {
			if i > 0 {
				sb.WriteString(" | ")
			}
			v := row[col]
			if v == nil {
				sb.WriteString("")
			} else {
				sb.WriteString(fmt.Sprint(v))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// Output 获取输出
func (e *ExecuteSQL) Output(key string) any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.output[key]
}

// OutputAll 获取所有输出
func (e *ExecuteSQL) OutputAll() map[string]any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range e.output {
		out[k] = v
	}
	return out
}

// SetOutput 设置输出
func (e *ExecuteSQL) SetOutput(key string, value any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.output[key] = value
}

// Error 返回错误
func (e *ExecuteSQL) Error() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.err
}

// Reset 重置
func (e *ExecuteSQL) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.output = make(map[string]any)
	e.err = ""
}
