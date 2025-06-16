package m_20240101000000_CreateInitialTables

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func Up(ctx context.Context, tx gdb.TX) error {
	// SQL for creating the test_migrations table
	sql := `
	CREATE TABLE test_migrations (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	_, err := tx.Exec(ctx, sql)
	if err != nil {
		return err
	}
	g.Log().Info(ctx, "Migration CreateInitialTables: table test_migrations created successfully")
	return nil
}

func Down(ctx context.Context, tx gdb.TX) error {
	// SQL for dropping the test_migrations table
	sql := `DROP TABLE IF EXISTS test_migrations;`
	_, err := tx.Exec(ctx, sql)
	if err != nil {
		return err
	}
	g.Log().Info(ctx, "Migration CreateInitialTables: table test_migrations dropped successfully")
	return nil
}
