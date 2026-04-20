package repository

import (
	"regexp"
	"strings"
	"time"

	observer "github.com/yca-software/go-common/observer"
)

var (
	// SQL patterns to extract table names
	// Handles: table, "table", 'table', schema.table, "schema"."table"
	selectTableRegex = regexp.MustCompile(`(?i)\bFROM\s+(?:["']?[\w.]+["']?\.)?["']?(\w+)["']?`)
	insertTableRegex = regexp.MustCompile(`(?i)\bINSERT\s+INTO\s+(?:["']?[\w.]+["']?\.)?["']?(\w+)["']?`)
	updateTableRegex = regexp.MustCompile(`(?i)\bUPDATE\s+(?:["']?[\w.]+["']?\.)?["']?(\w+)["']?`)
	deleteTableRegex = regexp.MustCompile(`(?i)\bDELETE\s+FROM\s+(?:["']?[\w.]+["']?\.)?["']?(\w+)["']?`)
)

// QueryInfo contains parsed information about a SQL query
type QueryInfo struct {
	Table     string
	Operation string
	QueryType string
}

// recordMetrics records query metrics if a hook is provided
func recordMetrics(hook observer.QueryMetricsHook, query string, operation string, start time.Time, err error) {
	if hook == nil {
		return
	}

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
	}

	queryInfo := ParseQuery(query, operation)
	hook.RecordQuery(queryInfo.Operation, queryInfo.Table, queryInfo.QueryType, status, duration)
}

// ParseQuery extracts metadata from a SQL query string
func ParseQuery(query string, operation string) QueryInfo {
	query = strings.TrimSpace(query)
	queryUpper := strings.ToUpper(query)

	info := QueryInfo{
		Operation: operation,
		Table:     "unknown",
		QueryType: operation,
	}

	// Extract table name based on operation type
	var tableName string
	switch {
	case strings.HasPrefix(queryUpper, "SELECT"):
		if matches := selectTableRegex.FindStringSubmatch(query); len(matches) > 1 {
			tableName = matches[1]
		}
	case strings.HasPrefix(queryUpper, "INSERT"):
		if matches := insertTableRegex.FindStringSubmatch(query); len(matches) > 1 {
			tableName = matches[1]
		}
	case strings.HasPrefix(queryUpper, "UPDATE"):
		if matches := updateTableRegex.FindStringSubmatch(query); len(matches) > 1 {
			tableName = matches[1]
		}
	case strings.HasPrefix(queryUpper, "DELETE"):
		if matches := deleteTableRegex.FindStringSubmatch(query); len(matches) > 1 {
			tableName = matches[1]
		}
	}

	if tableName != "" {
		info.Table = strings.ToLower(tableName)
		info.QueryType = strings.ToLower(operation) + "_" + info.Table
	}

	return info
}
