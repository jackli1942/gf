package department_test

import (
    "context"
    "testing"
    "yuncms/internal/logic/department"
    "yuncms/internal/model/input"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gcfg"
    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"
    _ "github.com/gogf/gf/contrib/drivers/mysql/v2"
    _ "yuncms/internal/boot" // For i18n
)

func init() {
    ctx := context.Background()
    configPath := ""
    // Determine config path based on common execution CWDs for tests
    if gfile.Exists("/app/manifest/config") {
        configPath = "/app/manifest/config"
    } else if gfile.Exists("../../../manifest/config") {
        configPath = "../../../manifest/config"
    } else if gfile.Exists("../../manifest/config") {
         configPath = "../../manifest/config"
    } else if gfile.Exists("manifest/config") {
         configPath = "manifest/config"
    } else {
        glog.Fatal(ctx, "DeptLogicTest Init: Config directory `manifest/config` not found. Tests cannot run.")
    }

    if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
        // Directly set the path for the test context.
        adapter.SetPath(configPath)
        glog.Debug(ctx, "DeptLogicTest Init: Config path set to:", configPath)

        i18nPath := g.Cfg().MustGet(ctx, "i18n.path", "resource/i18n").String()
        effectiveI18nPath := i18nPath
        if !gfile.Exists(effectiveI18nPath) {
            if gfile.Exists("/app/" + i18nPath) {
                effectiveI18nPath = "/app/" + i18nPath
            } else if gfile.Exists("../../../" + i18nPath) {
                 effectiveI18nPath = "../../../" + i18nPath
            } else if gfile.Exists("../../" + i18nPath) {
                 effectiveI18nPath = "../../" + i18nPath
            }
        }
        if gfile.Exists(effectiveI18nPath) {
            g.I18n().SetPath(effectiveI18nPath)
            glog.Debugf(ctx, "DeptLogicTest Init: i18n path set to: %s", effectiveI18nPath)
        } else {
            glog.Warningf(ctx, "DeptLogicTest Init: i18n path not found at %s or common relatives.", i18nPath)
        }
    } else {
        glog.Warning(ctx, "DeptLogicTest Init: Default config adapter is not *gcfg.AdapterFile.")
    }
}

func TestDepartmentLogic_CreateAndGet(t *testing.T) {
    ctx := context.Background()
    deptLogic := department.NewDepartmentLogic()
    testDeptName := "Test Dept Logic Main"

    _, _ = g.DB().Model("departments").Where("name", testDeptName).Delete()

    createIn := &input.DepartmentCreateInput{Name: testDeptName, ParentId: 0, Leader: "Test Leader", Status: 1, SortOrder: 10}
    created, err := deptLogic.Create(ctx, createIn)
    if err != nil || created == nil { t.Fatalf("Create failed: %v", err) }
    if created.Name != testDeptName { t.Errorf("Expected name %s, got %s", testDeptName, created.Name) }
    if created.Id == 0 { t.Errorf("Expected created ID to be non-zero")}

    fetched, err := deptLogic.GetById(ctx, created.Id)
    if err != nil || fetched == nil { t.Fatalf("GetById failed: %v", err) }
    if fetched.Department.Name != testDeptName { t.Errorf("Fetched name mismatch") }

    _, _ = g.DB().Model("departments").Where("id", created.Id).Delete()
}

func TestDepartmentLogic_Update(t *testing.T) {
    ctx := context.Background()
    deptLogic := department.NewDepartmentLogic()
    testDeptName := "UpdateDeptLogic"

    _, _ = g.DB().Model("departments").Where("name", testDeptName).Delete()
    createIn := &input.DepartmentCreateInput{Name: testDeptName, ParentId: 0, Leader: "Initial Leader", Status: 1}
    created, err := deptLogic.Create(ctx, createIn)
    if err != nil || created == nil { t.Fatalf("Setup: CreateUser failed for UpdateTest: %v", err) }

    updatedName := "Updated Department Name"
    updatedLeader := "New Leader"
    updatedStatus := 2
    updateIn := &input.DepartmentUpdateInput{Name: &updatedName, Leader: &updatedLeader, Status: &updatedStatus}

    err = deptLogic.Update(ctx, created.Id, updateIn)
    if err != nil { t.Fatalf("Update failed: %v", err) }

    fetched, err := deptLogic.GetById(ctx, created.Id)
    if err != nil || fetched == nil { t.Fatalf("GetById after update failed: %v", err)}
    if fetched.Department.Name != updatedName { t.Errorf("Expected updated name %s, got %s", updatedName, fetched.Department.Name) }
    if fetched.Department.Leader != updatedLeader { t.Errorf("Expected updated leader %s, got %s", updatedLeader, fetched.Department.Leader) }
    if fetched.Department.Status != updatedStatus { t.Errorf("Expected updated status %d, got %d", updatedStatus, fetched.Department.Status) }

    _, _ = g.DB().Model("departments").Where("id", created.Id).Delete()
}

func TestDepartmentLogic_Delete(t *testing.T) {
    ctx := context.Background()
    deptLogic := department.NewDepartmentLogic()
    testDeptName := "DeleteDeptLogic"

    _, _ = g.DB().Model("departments").Where("name", testDeptName).Delete()
    // Corrected: Removed Password field from DepartmentCreateInput
    created, err := deptLogic.Create(ctx, &input.DepartmentCreateInput{Name: testDeptName, Leader: "ToDelete"})
    if err != nil || created == nil { t.Fatalf("Setup: Create failed for DeleteTest: %v", err) }

    err = deptLogic.Delete(ctx, created.Id)
    if err != nil { t.Fatalf("Delete failed: %v", err) }

    deleted, _ := deptLogic.GetById(ctx, created.Id)
    if deleted != nil { t.Errorf("Expected department to be nil after delete, got %+v", deleted) }
}

func TestDepartmentLogic_ListAndTree(t *testing.T) {
    ctx := context.Background()
    deptLogic := department.NewDepartmentLogic()
    prefix := "ListDeptLogic_"

    _, _ = g.DB().Model("departments").Where("name LIKE ?", prefix+"%").Delete()

    d1, _ := deptLogic.Create(ctx, &input.DepartmentCreateInput{Name: prefix + "Parent1", SortOrder: 1})
    d2, _ := deptLogic.Create(ctx, &input.DepartmentCreateInput{Name: prefix + "Child1.1", ParentId: d1.Id, SortOrder: 1})
    d3, _ := deptLogic.Create(ctx, &input.DepartmentCreateInput{Name: prefix + "Child1.2", ParentId: d1.Id, SortOrder: 2})
    d4, _ := deptLogic.Create(ctx, &input.DepartmentCreateInput{Name: prefix + "Parent2", SortOrder: 2})

    if d1 == nil || d2 == nil || d3 == nil || d4 == nil { t.Fatal("Setup: Failed to create departments for List/Tree test") }

    listResult, err := deptLogic.List(ctx, &input.DepartmentListInput{Page: 1, PageSize: 10})
    if err != nil { t.Fatalf("List failed: %v", err) }
    if listResult == nil { t.Fatalf("ListUsers result is nil") } // Corrected to ListUsers

    if listResult.Total < 4 { t.Errorf("Expected total at least 4, got %d", listResult.Total)}

    treeResult, err := deptLogic.GetTree(ctx)
    if err != nil { t.Fatalf("GetTree failed: %v", err) }

    foundD1 := false
    for _, rootNode := range treeResult {
        if rootNode.Id == d1.Id {
            foundD1 = true
            if len(rootNode.Children) != 2 { t.Errorf("Expected Parent1 to have 2 children, got %d", len(rootNode.Children))}
            foundD2, foundD3 := false, false
            for _, childNode := range rootNode.Children {
                if childNode.Id == d2.Id { foundD2 = true }
                if childNode.Id == d3.Id { foundD3 = true }
            }
            if !foundD2 || !foundD3 { t.Errorf("Did not find all children of Parent1")}
            break
        }
    }
    if !foundD1 { t.Error("Parent1 not found at root of tree")}

    _, _ = g.DB().Model("departments").Where("name LIKE ?", prefix+"%").Delete()
}
