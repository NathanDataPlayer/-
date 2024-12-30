package admin

import (
    "database/sql"
    "net/http"
    "log"
	"fmt"
    "github.com/gin-gonic/gin"
    "monitor_system/backend/config" 
)

var db *sql.DB

type Task struct {
    ID          int      `json:"id"`
    TableName   string   `json:"table_name"`
    MonitorTime string   `json:"monitor_time"`
    Indicator   string `json:"indicators"`
}

func init() {
    var err error
    dbConfig := config.NewDBConfig() // 确保能获取 DB 配置信息
    db, err = sql.Open("mysql", dbConfig.GetDSN())
    if err != nil {
        log.Fatalf("Error opening database connection: %v", err)
    }
    if err := db.Ping(); err != nil {
        log.Fatalf("Error pinging database: %v", err)
    }
}

// getTasks 获取所有任务
func GetTasks(c *gin.Context) {
    rows, err := db.Query("SELECT id, table_name, monitor_time, indicator FROM table_configs")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var tasks []Task
    for rows.Next() {
		var task Task
        if err := rows.Scan(&task.ID, &task.TableName, &task.MonitorTime, &task.Indicator); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        tasks = append(tasks, task)
    }
    
    c.JSON(http.StatusOK, tasks)
}

// 根据 ID 获取指定任务
func GetTask(c *gin.Context) {
    
}

// 创建新任务
func CreateTask(c *gin.Context) {
    var task Task
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    query := "INSERT INTO table_configs (table_name, monitor_time, indicator) VALUES (?, ?, ?)"
    _, err := db.Exec(query, task.TableName, task.MonitorTime, task.Indicator)
    if err != nil {
        log.Printf("Failed to save config: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add config"})
        return
    }
    
    c.JSON(http.StatusCreated, task)
}

// 更新指定任务
func UpdateTask(c *gin.Context) {
    id := c.Param("id")

    var input Task
    // 绑定请求的JSON数据到输入结构体
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    //
    query := fmt.Sprintf("UPDATE table_configs SET table_name = '%s', monitor_time = '%s', indicator = '%s' WHERE id = '%s'",
        input.TableName, input.MonitorTime, input.Indicator, id)

    // 执行更新操作
    _, err := db.Exec(query)
    if err != nil {
        log.Printf("Failed to update task: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
        return
    }

    // 返回更新后的任务信息
    c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

// 删除指定任务
func DeleteTask(c *gin.Context) {
    id := c.Param("id")
	fmt.Println("Deleting task with ID:", id)

    query := fmt.Sprintf("DELETE FROM table_configs WHERE id = %s", id)

	_ ,err := db.Exec(query)
    if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task: " + err.Error()})
        return
    }
   
    c.JSON(http.StatusNoContent, nil)
}