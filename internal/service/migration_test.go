package service_test

import (
	"context"
	"testing"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/test/gtest" // Still use gtest.C for structure
    "github.com/gogf/gf/v2/os/gcfg"
    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"
     // Import the MySQL driver
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func init() {
    configPath := "/app/manifest/config"
    ctx := context.Background()
    if !gfile.Exists(configPath) {
        altPath := "../../manifest/config"
        if gfile.Exists(altPath) {
            configPath = altPath
        } else {
            configPath = "manifest/config"
        }
    }
    if adapterFile, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
        adapterFile.SetPath(configPath)
        glog.Debug(ctx, "Test Init: Configuration path set to:", configPath)
    } else {
        glog.Warning(ctx, "Test Init: Default config adapter is not *gcfg.AdapterFile, path not set for tests.")
    }
}

func TestMigration(t *testing.T) {
    gtest.C(t, func(t *gtest.T) { // Use gtest.C for setup/teardown context if needed, but use std assertions
        ctx := context.Background()
        db := g.DB()

        tables, err := db.Tables(ctx)
        if err != nil { // Using std lib style check
            t.Fatalf("Failed to connect to database or list tables: %v. Ensure DB is running and config is correct.", err)
        }
        glog.Debugf(ctx, "Tables found in database: %#v", tables)

        var foundMigrationsTable bool = false
        for _, tableName := range tables {
            glog.Debugf(ctx, "Checking table: '%s', comparing with 'gf_migrations', result: %v", tableName, tableName == "gf_migrations")
            if tableName == "gf_migrations" {
                glog.Debugf(ctx, "Found gf_migrations table: %s", tableName)
                foundMigrationsTable = true
                break;
            }
        }
        if !foundMigrationsTable {
             t.Errorf("gf_migrations table should exist after manual SQL execution, but not found.")
        }

        var foundTestMigrationsTable bool = false
        for _, tableName := range tables {
            glog.Debugf(ctx, "Checking table: '%s', comparing with 'test_migrations', result: %v", tableName, tableName == "test_migrations")
            if tableName == "test_migrations" {
                glog.Debugf(ctx, "Found test_migrations table: %s", tableName)
                foundTestMigrationsTable = true
                break;
            }
        }
        if !foundTestMigrationsTable {
             t.Errorf("test_migrations table should exist after manual SQL execution, but not found.")
        }
    })
}
