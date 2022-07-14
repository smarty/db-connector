package mysql

import (
	"database/sql"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
)

var isolationLevels = map[sql.IsolationLevel]string{
	sql.LevelDefault:         "READ-COMMITTED",
	sql.LevelReadUncommitted: "READ-UNCOMMITTED",
	sql.LevelReadCommitted:   "READ-COMMITTED",
	sql.LevelWriteCommitted:  "WRITE-COMMITTED",
	sql.LevelRepeatableRead:  "REPEATABLE-READ",
	sql.LevelSnapshot:        "SNAPSHOT",
	sql.LevelSerializable:    "SERIALIZABLE",
	sql.LevelLinearizable:    "LINEARIZABLE",
}

func tryReadValue(value string) string {
	if len(value) == 0 {
		return ""
	} else if parsed := parseURL(value); parsed != nil && parsed.Scheme == "env" {
		return os.Getenv(parsed.Host)
	} else if parsed != nil && parsed.Scheme == "file" {
		raw, _ := ioutil.ReadFile(parsed.Path)
		value = strings.TrimSpace(string(raw))
		return value
	} else {
		return value
	}
}
func parseURL(value string) *url.URL {
	value = strings.TrimSpace(value)
	if len(value) == 0 {
		return nil
	} else if parsed, err := url.Parse(value); err != nil {
		return nil
	} else {
		return parsed
	}
}
