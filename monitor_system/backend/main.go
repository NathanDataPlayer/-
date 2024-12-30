package main

import (
    "monitor_system/backend/routers"
    "monitor_system/backend/monitor"
    "log"
)

func main() {
    // 启动监控任务
    go monitor.StartMonitoring()

    // 设置路由
    r := routers.SetupRouter()

    // 启动 HTTP 服务
    if err := r.Run("0.0.0.0:8080"); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}