package logic

import (
	"context"
	"errors" // For errors.Is
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"gf_project/api/v1/system" // For DeptTreeItem
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"gf_project/internal/modules/system/model/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/grand"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	_ "github.com/hailaz/gf-casbin-adapter/v2"
)

var testDeptCtxGlobal = gctx.New()

func TestMain(m *testing.M) {
	ctx := testDeptCtxGlobal
	projectRoot := gfile.normalize(gfile.Pwd() + "/../../../")

	configDirPath := gfile.Join(projectRoot, "manifest/config")
	configFileActualPath := gfile.Join(configDirPath, "config.yaml")

	if gfile.Exists(configFileActualPath) {
		if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
			adapter.SetPath(configDirPath)
			g.Log().Debugf(ctx, "TestMain (dept_service_test): Set config path to: %s", configDirPath)
		} else {
			g.Log().Warningf(ctx, "TestMain (dept_service_test): Config adapter is not *gcfg.AdapterFile, path not set.")
		}
	} else {
		g.Log().Warningf(ctx, "TestMain (dept_service_test): config.yaml not found at: %s", configFileActualPath)
	}

	modelPathFromConfig := g.Cfg().MustGet(ctx, "casbin.modelPath", "manifest/config/casbin_model.conf").String()
	resolvedModelPath := modelPathFromConfig
	if !gfile.IsAbs(modelPathFromConfig) {
		cfgAdapterPath := ""
		if adapterFile, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
			cfgAdapterPath = adapterFile.GetPath()
		}
		if cfgAdapterPath != "" && gfile.Exists(gfile.Join(cfgAdapterPath, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(cfgAdapterPath, modelPathFromConfig)
		} else if projectRoot != "" && gfile.Exists(gfile.Join(projectRoot, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(projectRoot, modelPathFromConfig)
		} else if gfile.Exists(modelPathFromConfig) {
			resolvedModelPath = modelPathFromConfig
		}
	}
	if !gfile.Exists(resolvedModelPath) {
		g.Log().Fatalf(ctx, "TestMain (dept_service_test): Casbin model file '%s' not found.", resolvedModelPath)
	}
	g.Cfg().Set("casbin.modelPath", resolvedModelPath)
	g.Log().Infof(ctx, "TestMain (dept_service_test): Effective Casbin model path: %s", resolvedModelPath)

	if _, err := CasbinEnforcer(); err != nil {
		g.Log().Fatalf(testDeptCtxGlobal, "TestMain (dept_service_test): Failed to initialize Casbin enforcer: %v", err)
	}

	clearTestDeptAndRelatedTables(ctx) // Clear once before all tests in this file

	exitCode := m.Run()
	os.Exit(exitCode)
}

func clearTestDeptAndRelatedTables(ctx context.Context) {
	tablesToClear := []string{
		dao.SystemUser.Table(), // Clear users first due to potential FK on dept_id
		dao.SystemDept.Table(),
	}
	db := g.DB()
	allTablesInDb, _ := db.Ctx(ctx).Tables()
	existingTablesMap := make(map[string]bool)
	for _, t := range allTablesInDb {
		existingTablesMap[t] = true
	}

	for _, table := range tablesToClear {
		if !existingTablesMap[table] {
			g.Log().Debugf(ctx, "Table %s not found, skipping clear for dept tests.", table)
			continue
		}
		_, err := db.Ctx(ctx).Delete(table)
		if err != nil {
			g.Log().Errorf(ctx, "Failed to clear table %s for dept tests: %v", table, err)
		}
		if db.GetConfig().DriverName == "sqlite" {
			_, errSeq := db.Ctx(ctx).Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s';", table))
			if errSeq != nil {
				g.Log().Debugf(ctx, "Failed to reset sqlite_sequence for table %s (dept tests): %v", table, errSeq)
			}
		}
	}
}

func TestDeptService_CreateDept_And_Ancestors(t *testing.T) {
	ctx := testDeptCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestDeptAndRelatedTables(ctx)

		// Status: 0 for Normal, 1 for Disabled (as per User input struct, standardizing here)
		rootInp := &input.DeptCreateInp{ParentId: 0, Name: "Root Corp", Status: 0, Sort: 10}
		rootId, err := DeptService.CreateDept(ctx, rootInp)
		t.AssertNil(err, "Creating root department should not error")
		rootDept, _ := DeptService.GetDeptById(ctx, rootId)
		t.AssertNE(rootDept, nil, "Root department should be retrievable")
		// Ancestor path for root (parentId 0) is defined as ",0," by getAncestorsPath in service
		t.AssertEQ(rootDept.Level, ",0,", "Root dept Level (ancestor path) should be ',0,'")

		childInp := &input.DeptCreateInp{ParentId: rootId, Name: "IT Department", Status: 0, Sort: 10}
		childId, errChild := DeptService.CreateDept(ctx, childInp)
		t.AssertNil(errChild)
		childDept, _ := DeptService.GetDeptById(ctx, childId)
		t.AssertNE(childDept, nil)
		expectedChildLevel := fmt.Sprintf(",0,%d,", rootId)
		t.AssertEQ(childDept.Level, expectedChildLevel, "Child dept Level (ancestor path) not as expected")

		grandChildInp := &input.DeptCreateInp{ParentId: childId, Name: "Software Dev", Status: 0, Sort: 10}
		grandChildId, errGrandChild := DeptService.CreateDept(ctx, grandChildInp)
		t.AssertNil(errGrandChild)
		grandChildDept, _ := DeptService.GetDeptById(ctx, grandChildId)
		t.AssertNE(grandChildDept, nil)
		expectedGrandChildLevel := fmt.Sprintf(",0,%d,%d,", rootId, childId)
		t.AssertEQ(grandChildDept.Level, expectedGrandChildLevel, "Grandchild dept Level (ancestor path) not as expected")
	})
}

func TestDeptService_UpdateDept_ChangeParent(t *testing.T) {
	ctx := testDeptCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestDeptAndRelatedTables(ctx)
		// R (id assumed 1, level ",0,") -> D1 (id 2, level ",0,1,") -> D1_C1 (id 3, level ",0,1,2,")
		// R (id assumed 1, level ",0,") -> D2 (id 4, level ",0,1,")
		rId, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: 0, Name: "R", Status: 0, Sort: 1})
		d1Id, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: rId, Name: "D1", Status: 0, Sort: 1})
		d1c1Id, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: d1Id, Name: "D1_C1", Status: 0, Sort: 1})
		d2Id, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: rId, Name: "D2", Status: 0, Sort: 2})

		d1c1gcId, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: d1c1Id, Name: "D1_C1_GC", Status: 0, Sort: 1})

		// Move D1_C1 from D1 to D2
		err := DeptService.UpdateDept(ctx, &input.DeptUpdateInp{
			Id: d1c1Id, ParentId: d2Id, Name: "D1_C1 (Moved)", Status: 0, Sort: 1,
		})
		t.AssertNil(err)

		d1c1_moved, _ := DeptService.GetDeptById(ctx, d1c1Id)
		expectedMovedLevel := fmt.Sprintf(",0,%d,%d,", rId, d2Id)
		t.AssertEQ(d1c1_moved.Level, expectedMovedLevel, "D1_C1 Level after move to D2 is incorrect")

		d1c1gc_moved, _ := DeptService.GetDeptById(ctx, d1c1gcId)
		expectedGCLevel := fmt.Sprintf(",0,%d,%d,%d,", rId, d2Id, d1c1Id)
		t.AssertEQ(d1c1gc_moved.Level, expectedGCLevel, "D1_C1_GC Level after parent D1_C1 moved to D2 is incorrect")
	})
}

func TestDeptService_UpdateDept_PreventMoveToChild(t *testing.T) {
	ctx := testDeptCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestDeptAndRelatedTables(ctx)
		rId, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: 0, Name: "R", Status: 0, Sort: 1})
		d1Id, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: rId, Name: "D1", Status: 0, Sort: 1})
		d1c1Id, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: d1Id, Name: "D1_C1", Status: 0, Sort: 1})

		err := DeptService.UpdateDept(ctx, &input.DeptUpdateInp{
			Id: d1Id, ParentId: d1c1Id, Name: "D1 (Illegal Move)", Status: 0, Sort: 1,
		})
		t.AssertNE(err, nil, "Should not be able to move a dept under its own child")
		t.Assert(gerror.Contains(err, "Cannot move a department under one of its own children"), "Error message for illegal move not as expected")
	})
}

func TestDeptService_DeleteDept(t *testing.T) {
	ctx := testDeptCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestDeptAndRelatedTables(ctx)
		rId, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: 0, Name: "R", Status: 0, Sort: 1})
		d1Id, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: rId, Name: "D1 To Delete", Status: 0, Sort: 1})
		d1c1Id, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: d1Id, Name: "D1_C1 Child", Status: 0, Sort: 1})

		err := DeptService.DeleteDept(ctx, d1Id) // This should delete D1 and D1_C1
		t.AssertNil(err)

		d1Deleted, _ := DeptService.GetDeptById(ctx, d1Id)
		t.AssertNil(d1Deleted, "D1 should be deleted")
		d1c1Deleted, _ := DeptService.GetDeptById(ctx, d1c1Id)
		t.AssertNil(d1c1Deleted, "D1_C1 (child) should also be deleted")

		rExists, _ := DeptService.GetDeptById(ctx, rId)
		t.AssertNE(rExists, nil, "Root R should still exist")
	})
}

func TestDeptService_DeleteDept_WithUsers(t *testing.T) {
	ctx := testDeptCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestDeptAndRelatedTables(ctx)
		deptId, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: 0, Name: "Dept With User", Status: 0, Sort: 1})

		_, _ = UserService.CreateUser(ctx, &input.UserCreateInp{
			Username: "user_in_dept_del_test", Password: "pw", Nickname: "UID", Status: 0, DeptId: deptId,
		})

		err := DeptService.DeleteDept(ctx, deptId)
		t.AssertNE(err, nil, "DeleteDept should fail if users are assigned")
		t.Assert(gerror.Contains(err, "users are assigned"), "Error message for users in dept not as expected")
	})
}

func TestDeptService_GetDeptList(t *testing.T) {
	ctx := testDeptCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestDeptAndRelatedTables(ctx)
		r1Id, _ := DeptService.CreateDept(ctx, &input.DeptCreateInp{Name: "R1", ParentId: 0, Status: 0, Sort: 10})
		_, _ = DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: r1Id, Name: "D1.2", Code: "d12", Type: "M", Status: 0, Sort: 20, IsHidden: 0})
		_, _ = DeptService.CreateDept(ctx, &input.DeptCreateInp{ParentId: r1Id, Name: "D1.1", Code: "d11", Type: "M", Status: 0, Sort: 10, IsHidden: 0})
		_, _ = DeptService.CreateDept(ctx, &input.DeptCreateInp{Name: "R2", ParentId: 0, Status: 0, Sort: 20})

		tree, err := DeptService.GetDeptList(ctx, &input.DeptListInp{})
		t.AssertNil(err)
		t.AssertEQ(len(tree), 2)
		t.AssertEQ(tree[0].Name, "R1") // Sorted by SortOrder
		t.AssertEQ(len(tree[0].Children), 2)
		t.AssertEQ(tree[0].Children[0].Name, "D1.1") // Sorted by SortOrder
		t.AssertEQ(tree[0].Children[1].Name, "D1.2")
		t.AssertEQ(tree[1].Name, "R2")
	})
}
