package post_test

import (
    "context"
    "testing"
    "yuncms/internal/logic/post"
    "yuncms/internal/model/input"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gcfg"
    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"
    _ "github.com/gogf/gf/contrib/drivers/mysql/v2"
    _ "yuncms/internal/boot"
    // "yuncms/internal/utility/i18nutil" // For checking specific translated errors if needed
)

func init() {
    ctx := context.Background()
    configPath := "/app/manifest/config"
    if !gfile.Exists(configPath) {
        configPath = "../../../manifest/config"
        if !gfile.Exists(configPath) {
             configPath = "../../manifest/config"
             if !gfile.Exists(configPath) {
                configPath = "manifest/config"
             }
        }
    }

    if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
        adapter.SetPath(configPath)
        glog.Debug(ctx, "PostLogicTest Init: Config path set to:", configPath)

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
            glog.Debugf(ctx, "PostLogicTest Init: i18n path set to: %s", effectiveI18nPath)
        } else {
            glog.Warningf(ctx, "PostLogicTest Init: i18n path not found at %s or common relatives.", i18nPath)
        }
    } else {
        glog.Warning(ctx, "PostLogicTest Init: Default config adapter is not *gcfg.AdapterFile.")
    }
}

func TestPostLogic_CreateAndGet(t *testing.T) {
    ctx := context.Background()
    logic := post.NewPostLogic()
    testCode := "post_logic_test_c001"
    testName := "Test Post Create"

    // Cleanup if exists
    _, _ = g.DB().Model("posts").Where("code", testCode).Delete()

    in := &input.PostCreateInput{Code: testCode, Name: testName, Status: 1, SortOrder: 1, Remark: "Test Create"}
    created, err := logic.Create(ctx, in)
    if err != nil || created == nil { t.Fatalf("Create failed: %v", err) }
    if created.Code != testCode { t.Errorf("Expected code %s, got %s", testCode, created.Code)}
    if created.Id == 0 { t.Errorf("Expected created ID to be non-zero")}

    fetched, err := logic.GetById(ctx, created.Id)
    if err != nil || fetched == nil { t.Fatalf("GetById failed: %v", err) }
    if fetched.Name != testName { t.Errorf("Name mismatch on GetById") }

    // Test GetByCode (implicitly tested by Create's duplicate check, but good to test directly)
    fetchedByCode, err := logic.GetById(ctx, created.Id) // Re-fetch by ID to get the entity for GetByCode test
    if err != nil || fetchedByCode == nil {t.Fatalf("Re-fetch for GetByCode test failed: %v", err)}

    // Actually, the logic layer doesn't have GetByCode, DAO does.
    // Create's duplicate check is sufficient for testing that GetByCode in DAO works.

    // Test duplicate code creation
    _, err = logic.Create(ctx, in) // Try to create again
    if err == nil { t.Errorf("Expected error when creating post with duplicate code, got nil") }
    // Add more specific error check if needed:
    // expectedErrStr := i18nutil.T(ctx, "post_code_exists", "Code", testCode)
    // if err.Error() != expectedErrStr { t.Errorf("Expected error '%s', got '%s'", expectedErrStr, err.Error())}


    // Cleanup
    _, _ = g.DB().Model("posts").Where("id", created.Id).Delete()
}

func TestPostLogic_Update(t *testing.T) {
    ctx := context.Background()
    logic := post.NewPostLogic()
    baseCode := "post_update_001"
    baseName := "Base Update Post"

    _, _ = g.DB().Model("posts").Where("code LIKE ?", baseCode+"%").Delete()

    in := &input.PostCreateInput{Code: baseCode, Name: baseName, Status: 1}
    created, _ := logic.Create(ctx, in)
    if created == nil { t.Fatal("Setup: Create failed for UpdateTest") }

    updatedName := "Updated Post Name"
    updatedCode := baseCode + "_updated" // Test code update as well
    updatedStatus := 2
    updateIn := &input.PostUpdateInput{Name: &updatedName, Code: &updatedCode, Status: &updatedStatus}

    err := logic.Update(ctx, created.Id, updateIn)
    if err != nil { t.Fatalf("Update failed: %v", err)}

    fetched, _ := logic.GetById(ctx, created.Id)
    if fetched.Name != updatedName { t.Errorf("Expected updated name %s, got %s", updatedName, fetched.Name)}
    if fetched.Code != updatedCode { t.Errorf("Expected updated code %s, got %s", updatedCode, fetched.Code)}
    if fetched.Status != updatedStatus { t.Errorf("Expected updated status %d, got %d", updatedStatus, fetched.Status)}

    _, _ = g.DB().Model("posts").Where("id", created.Id).Delete()
}

func TestPostLogic_Delete(t *testing.T) {
    ctx := context.Background()
    logic := post.NewPostLogic()
    testCode := "post_delete_001"

    _, _ = g.DB().Model("posts").Where("code", testCode).Delete()
    created, _ := logic.Create(ctx, &input.PostCreateInput{Code: testCode, Name: "To Be Deleted"})
    if created == nil { t.Fatal("Setup: Create failed for DeleteTest") }

    err := logic.Delete(ctx, created.Id)
    if err != nil { t.Fatalf("Delete failed: %v", err) }

    deleted, _ := logic.GetById(ctx, created.Id)
    if deleted != nil { t.Errorf("Expected post to be nil after delete, got %+v", deleted)}
}

func TestPostLogic_List(t *testing.T) {
    ctx := context.Background()
    logic := post.NewPostLogic()
    prefix := "post_list_test_"
    code1, name1 := prefix+"c1", prefix+"n1"
    code2, name2 := prefix+"c2", prefix+"n2"

    _, _ = g.DB().Model("posts").Where("code LIKE ?", prefix+"%").Delete()
    p1, _ := logic.Create(ctx, &input.PostCreateInput{Code: code1, Name: name1})
    p2, _ := logic.Create(ctx, &input.PostCreateInput{Code: code2, Name: name2})
    if p1 == nil || p2 == nil { t.Fatal("Setup: Failed to create posts for List test")}

    listIn := &input.PostListInput{Page: 1, PageSize: 5}
    result, err := logic.List(ctx, listIn)
    if err != nil { t.Fatalf("List failed: %v", err)}
    if result.Total < 2 { t.Errorf("Expected total at least 2, got %d", result.Total)}

    found1, found2 := false, false
    for _, p := range result.List {
        if p.Code == code1 { found1 = true }
        if p.Code == code2 { found2 = true }
    }
    if !found1 || !found2 {t.Errorf("Expected created posts in list. Found1: %t, Found2: %t", found1, found2)}

    _, _ = g.DB().Model("posts").Where("code LIKE ?", prefix+"%").Delete()
}
