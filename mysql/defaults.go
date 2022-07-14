package mysql

import (
	"database/sql"
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
