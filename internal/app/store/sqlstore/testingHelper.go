package sqlstore

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"testing"
	"time"
)

func TestDB(t *testing.T, databaseDSN string) (*sql.DB, func(...string)) {
	t.Helper()
	db, err := sql.Open("mysql", databaseDSN)
	if err != nil {
		t.Fatal(err)
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	return db, func(tables ...string) {
		if len(tables) > 0 {
			if _, err := db.Exec(fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", "))); err != nil {
				t.Fatal(err)
			}
		}
		db.Close()
	}
}
