package models

import (
    "database/sql"
    "fmt"
    "time"
)

// Config 表示监控配置
type Config struct {
    ID          int       `db:"id"`
    TableName   string    `db:"table_name"`
    MonitorTime time.Time `db:"monitor_time"`
}

// TableMetric 表示监控指标
type TableMetric struct {
    ID        int       `db:"id"`
    TableName string    `db:"table_name"`
    Timestamp time.Time `db:"timestamp"`
    RowCount  int       `db:"row_count"`
}

// FetchAllConfigs 从数据库获取所有的监控配置
func FetchAllConfigs(db *sql.DB) ([]Config, error) {
    var configs []Config
    rows, err := db.Query("SELECT id, table_name, monitor_time FROM table_configs")
    if err != nil {
        return nil, fmt.Errorf("error fetching table configs: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var config Config
        if err := rows.Scan(&config.ID, &config.TableName, &config.MonitorTime); err != nil {
            return nil, fmt.Errorf("error scanning config row: %w", err)
        }
        configs = append(configs, config)
    }
    return configs, nil
}

// StoreMetric 存储监控指标
func StoreMetric(db *sql.DB, metric TableMetric) error {
    _, err := db.Exec("INSERT INTO table_metrics (table_name, timestamp, row_count) VALUES (?, ?, ?)",
        metric.TableName, metric.Timestamp, metric.RowCount)
    return err
}

// FetchPreviousRowCount 获取指定表的上一个行数
func FetchPreviousRowCount(db *sql.DB, tableName string) (int, error) {
    var count int
    query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE DATE(dw_ins_time) = CURDATE() - INTERVAL 1 DAY AND TIME(dw_ins_time) = '%s'", tableName, time)
    err := db.QueryRow(query).Scan(&count)
    if err != nil {
        log.Printf("Error fetching previous row count for %s: %v", tableName, err)
        return 0, err
    }
    return count, nil
}

// FetchCurrentRowCount 获取指定表的当前行数
func FetchCurrentRowCount(db *sql.DB, tableName string, time string) (int, error) {
    var count int
    query := fmt.Sprintf("SELECT COUNT(*) FROM %s where DATE(dw_ins_time) = CURDATE() AND TIME(dw_ins_time) = '%s'", tableName, time)
    err := db.QueryRow(query).Scan(&count)
    if err != nil {
        log.Printf("Error fetching row count for %s: %v", tableName, err)
        return 0, err
    }
    return count, err
}