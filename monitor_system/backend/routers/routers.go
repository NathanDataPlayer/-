package routers

import (
    "github.com/gin-gonic/gin"
    "monitor_system/backend/monitor"
    "monitor_system/backend/admin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
    r := gin.Default()

    r.Static("/static", "/home/apps/monitor_system/frontend")

    r.GET("/", func(c *gin.Context) {
        c.File("/home/apps/monitor_system/frontend/index.html")
    })

    // 新增路由，访问 /task_management 时返回 task_management.html
    r.GET("/task_management", func(c *gin.Context) {
        c.File("/home/apps/monitor_system/frontend/task_management.html")
    })


    r.GET("/monitor/getDatabases", monitor.GetDatabases)
    r.GET("/monitor/getTables", monitor.GetTables)
    r.GET("/monitor/getColumns", monitor.GetColumns)
    r.POST("/monitor/addMonitorConfig", monitor.AddConfig)


    // 管理模块的路由
    r.GET("/monitor/tasks", admin.GetTasks)
    r.GET("/monitor/tasks/:id", admin.GetTask)
    r.POST("/monitor/tasks", admin.CreateTask)
    r.PUT("/monitor/tasks/:id", admin.UpdateTask)
    r.DELETE("/monitor/tasks/:id", admin.DeleteTask)

    // r.POST("/monitor/getIncreaseAbove50", monitor.GetIncreaseAbove50)
    // r.POST("/monitor/getDecreaseBelow50", monitor.GetDecreaseBelow50)
    // r.POST("/monitor/getZeroCount", monitor.GetZeroCount)

    return r
}