package monitor

import (
    "database/sql"
    "fmt"
    "net/http"
    "time"
    "log"
    "strings"
    "github.com/robfig/cron/v3"

    "github.com/gin-gonic/gin" 
    _ "github.com/go-sql-driver/mysql"
    "monitor_system/backend/config"
)

var db *sql.DB

var c *cron.Cron

type Config struct {
    ID          int    `json:"id"`
    TableName   string `json:"table_name"`
    MonitorTime string `json:"monitor_time" ` 
    Indicators  []string `json:"indicators"`
    PrimaryKey  []string `json:"primary_key"`
}

// 初始化数据库连接
func initDB() error {
    dbConfig := config.NewDBConfig()
    var err error
    db, err = sql.Open("mysql", dbConfig.GetDSN())
    if err != nil {
        return fmt.Errorf("Error opening database connection: %v", err)
    }
    if err := db.Ping(); err != nil {
        return fmt.Errorf("Error pinging database: %v", err)
    }
    return nil
}

// GetDatabases 用于获取数据库列表的API
func GetDatabases(c *gin.Context) {
    rows, err := db.Query("SHOW DATABASES")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var databases []string
    for rows.Next() {
        var database string
        if err := rows.Scan(&database); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        databases = append(databases, database)
    }
    c.JSON(http.StatusOK, databases)
}

// GetTables 用于获取指定数据库的表的API
func GetTables(c *gin.Context) {
    database := c.Query("database")
    if database == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Database name is required"})
        return
    }

    rows, err := db.Query("SHOW TABLES IN " + database)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var table string
        if err := rows.Scan(&table); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        tables = append(tables, table)
    }
    c.JSON(http.StatusOK, tables)
}

// GetColumns 
func GetColumns(c *gin.Context) {
    database := c.Query("database")
    table := c.Query("table")

    query := fmt.Sprintf("SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'", database, table)
    rows, err := db.Query(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch columns"})
        return
    }
    defer rows.Close()

    columns := []string{}
    for rows.Next() {
        var columnName string
        rows.Scan(&columnName)
        columns = append(columns, columnName)
    }
    c.JSON(http.StatusOK, columns)
}

// AddConfig 用于添加监控配置的API
func AddConfig(c *gin.Context) {
    var config Config
    if err := c.ShouldBindJSON(&config); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    log.Printf("Received config: %+v", config)

    if len(config.Indicators) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "At least one indicator required"})
        return
    }

    for _, indicator := range config.Indicators {

        // 生成唯一 ID
        id := int(time.Now().UnixNano() % 2147483647)

        query := fmt.Sprintf("INSERT INTO dw_monitor.table_configs (id, table_name, monitor_time,indicator) VALUES (%d, '%s', '%s','%s')", id,config.TableName, config.MonitorTime,indicator)
        log.Printf("Executing query: %s", query)

        _, err := db.Exec(query)
        if err != nil {
           log.Printf("Failed to save config: %v", err)
           c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add config"})
           return
        }

         // 动态添加新的 cron 任务
         if err := addCronJob(&config); err != nil {
         log.Printf("Error adding cron job for %s: %v", config.TableName, err)
        }
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}



func addCronJob(config *Config) error {

    monitorTime, err := time.Parse("15:04:05", config.MonitorTime)
    if err != nil {
        return fmt.Errorf("Error parsing monitor time for %s: %v", config.TableName, err)
    }
    cronSpec := fmt.Sprintf("%d %d %d * * *",monitorTime.Second(),monitorTime.Minute(), monitorTime.Hour())

    _, err = c.AddFunc(cronSpec, func() {
        performMonitoring(config.TableName,config.Indicators,config.PrimaryKey)
    })
    if err != nil {
        return fmt.Errorf("Error adding monitoring task to cron: %v", err)
    }

    log.Printf("Added cron job for %s with spec %s", config.TableName, cronSpec)
    fmt.Printf("Added cron job for %s with spec %s\n", config.TableName, cronSpec)
    return nil
}


// getRowCount, getPreviousRowCount, storeRowCount
func getRowCount(tableName string) (int, error) {
    var count int
    query := fmt.Sprintf("SELECT COUNT(*) FROM dw.%s WHERE DATE(dw_ins_time) = CURDATE()", tableName)
    err := db.QueryRow(query).Scan(&count)

    if err != nil {
        fmt.Printf("SQL error: %v\n", err) // 打印 SQL 错误
        return 0, err
    }
    return count, nil
}

func getPreviousRowCount(tableName string) (int, error) {
    var count int
    query := fmt.Sprintf("SELECT count(*) FROM dw.%s WHERE DATE(dw_ins_time) = CURDATE() - 1", tableName)
    err := db.QueryRow(query).Scan(&count)

    if err != nil {
        fmt.Printf("SQL error: %v\n", err) // 打印 SQL 错误
        return 0, err
    }
    return count, nil
}

func GetZeroCount(c *gin.Context) {
    var input struct {
        Database string `json:"database"`
        Table    string `json:"table"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
        return
    }

    currentRowCount, err := getRowCount(input.Table)

    fmt.Println("currentRowCount ",err)
    
    c.JSON(http.StatusOK, gin.H{
        "current_row_count": currentRowCount,
        "is_zero":           currentRowCount == 0,
    })
}

func getDuplicateKey(tableName string, primaryKeyFields []string) (int, error) {
     
    query := fmt.Sprintf("select count(*) from (select count(*) as cnt from dw.%s where DATE(dw_ins_time) = CURDATE() group by %s having cnt >1) aa",tableName, strings.Join(primaryKeyFields,","))
    
    var count int
    err := db.QueryRow(query).Scan(&count)
    if err != nil {
        fmt.Printf("Error checking for duplicate primary keys in table %s: %v\n", tableName, err) // 打印 SQL 错误
        return 0, err
    }

    return count, nil 

}


func getTableConfigs() ([]*Config, error) {
    var configs []*Config
    rows, err := db.Query("SELECT id, table_name, monitor_time FROM table_configs")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var config Config
        if err := rows.Scan(&config.ID, &config.TableName, &config.MonitorTime); err != nil {
            return nil, err
        }
        configs = append(configs, &config)
    }
    return configs, nil
}


// StartMonitoring 
func StartMonitoring() {

    if db == nil { // 确保在所有操作前 db 被初始化
        if err := initDB(); err != nil {
            log.Fatalf("Error initializing database: %v", err)
        }
    }

    c = cron.New(cron.WithSeconds())

    configs, err := getTableConfigs()

    if err != nil {
        log.Fatalf("Error getting table configs: %v", err)
    }

    for _, config := range configs {

        if err := addCronJob(config); err != nil {
            log.Printf("Error adding initial cron job for %s: %v", config.TableName, err)
        }

    }

    c.Start()

    entries := c.Entries()

    for _, entry := range entries {
        fmt.Printf("ID: %s, Expression: %s, Next Run: %s\n", 
        entry.ID, entry.Schedule, entry.Next)
    }
}


func performMonitoring(tableName string, indicators []string, primaryKeyFields []string) {
    if err := initDB(); err != nil {
        log.Printf("Error initializing database: %v", err)
        return
    }

    rowCount, err := getRowCount(tableName)
    if err != nil {
        log.Printf("Error getting row count for %s: %v", tableName, err)
        return
    }

    previousRowCount, err := getPreviousRowCount(tableName)
    if err != nil {
        log.Printf("Error getting previous row count for %s: %v", tableName, err)
        return
    }

    // 检测数据量为 0 的情况
    if rowCount == 0 {
        log.Printf("***WARNING*** Table %s has 0 rows.", tableName)
    }

    duplicateCount , err := getDuplicateKey(tableName, primaryKeyFields)
    if err != nil {
        log.Printf("Error getting duplicateCount row count for %s: %v", tableName, err)
        return
    }


    for _, indicator := range indicators {
        switch indicator {
        case "环比数据量 < 50%":
            if previousRowCount > 0 && ((rowCount-previousRowCount)*100/previousRowCount < 50) {
                log.Printf("Warning: The increase for table %s is below 50%%. Current: %d, Previous: %d.", tableName, rowCount, previousRowCount)
            }
        case "环比数据量 > 50%":
            if previousRowCount > 0 && ((rowCount-previousRowCount)*100/previousRowCount > 50) {
                log.Printf("Alert: The increase for table %s is above 50%%. Current: %d, Previous: %d.", tableName, rowCount, previousRowCount)
            }
        case "数量为零":
            if rowCount == 0 {
                log.Printf("***WARNING*** Table %s has 0 rows.", tableName)
            }
        case "主键重复":
            if duplicateCount > 0  {
                log.Printf("Alert: Duplicate primary keys found in table %s with fields %v, %d.", tableName, primaryKeyFields, duplicateCount)
            }
        }
    }

}