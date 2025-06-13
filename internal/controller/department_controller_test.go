package controller_test

import (
	"context"
	"fmt"
	"testing"
	"yuncms/internal/controller" // controller package
	"yuncms/internal/dao"        // For creating test data directly if needed
	"yuncms/internal/model/entity"
	"yuncms/internal/service" // service package

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/test/gtest"
	// "github.com/gogf/gf/v2/errors/gerror" // For error code checking if needed
)

// Helper to create a department for testing purposes
func createTestDepartment(ctx context.Context, name string, parentId uint, sort int, status int) (*entity.Department, error) {
	deptEntity := &entity.Department{
		Name:     name,
		ParentId: parentId,
		Sort:     sort,
		Status:   &status,
		// CreatedAt and UpdatedAt will be handled by ORM or DB
	}
	// Directly use DAO to create for test setup simplicity
	result, err := dao.Department.Ctx(ctx).Data(deptEntity).Insert()
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	deptEntity.Id = uint(id)
	return deptEntity, nil
}

// Helper to clean up departments by IDs
func cleanupDepartments(ctx context.Context, ids ...uint) {
	if len(ids) == 0 {
		return
	}
	_, err := dao.Department.Ctx(ctx).WhereIn(dao.Department.Columns().Id, ids).Delete()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to cleanup departments: %v", err)
	}
}

func TestDepartmentController_List(t *testing.T) {
	ctx := gctx.New()
	deptCtrl := controller.NewDepartmentController()
	var testDeptIds []uint

	// Setup: Create some test departments
	root1, err := createTestDepartment(ctx, "Root Dept 1", 0, 10, 1)
	gtest.AssertNil(t, err)
	if root1 != nil { testDeptIds = append(testDeptIds, root1.Id) }

	root2, err := createTestDepartment(ctx, "Root Dept 2", 0, 20, 1)
	gtest.AssertNil(t, err)
	if root2 != nil { testDeptIds = append(testDeptIds, root2.Id) }

	var child1 *entity.Department
	if root1 != nil {
		child1, err = createTestDepartment(ctx, "Child 1 of Root 1", root1.Id, 5, 1)
		gtest.AssertNil(t, err)
		if child1 != nil { testDeptIds = append(testDeptIds, child1.Id) }
	}

	disabledRoot, err := createTestDepartment(ctx, "Disabled Root Dept", 0, 30, 0)
	gtest.AssertNil(t, err)
	if disabledRoot != nil { testDeptIds = append(testDeptIds, disabledRoot.Id)}


	// Teardown: Ensure all created departments are cleaned up
	defer cleanupDepartments(ctx, testDeptIds...)

	gtest.C(t, func(t *gtest.T) {
		t.Run("List all (mostly root)", func(t *gtest.T) {
			req := &controller.DepartmentListApiReq{}
			res, err := deptCtrl.List(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.AssertNE(res.List, nil)

			// Check if root1 and root2 are present (and disabledRoot if not filtered)
			// The exact number of root items depends on if other tests left data or if DB is clean.
			// We expect at least 2 active root departments + 1 disabled one based on setup.
			// The service layer returns all items that match filter, then builds tree.
			// So, if no filter, all items are fetched.

			foundRoot1 := false
			foundRoot2 := false
			foundChild1InRoot1 := false

			for _, item := range res.List {
				if item.Id == root1.Id {
					foundRoot1 = true
					for _, childItem := range item.Children {
						if childItem.Id == child1.Id {
							foundChild1InRoot1 = true
							break
						}
					}
				}
				if item.Id == root2.Id {
					foundRoot2 = true
				}
			}
			t.Assert(foundRoot1, true, "Root Dept 1 not found in the list")
			t.Assert(foundRoot2, true, "Root Dept 2 not found in the list")
			if child1 != nil { // only assert if child1 was created
				t.Assert(foundChild1InRoot1, true, "Child 1 not found under Root Dept 1")
			}
		})

		t.Run("List with name filter", func(t *gtest.T) {
			req := &controller.DepartmentListApiReq{Name: "Root 1"}
			res, err := deptCtrl.List(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.Assert(len(res.List) >= 1, "Should find at least Root Dept 1")
			if len(res.List) > 0 {
				t.Assert(res.List[0].Id, root1.Id, "First item should be Root Dept 1")
				if child1 != nil { // Child might also match if name filter is broad
					foundChildInFiltered := false
					for _, childItem := range res.List[0].Children {
						if childItem.Id == child1.Id {
							foundChildInFiltered = true
							break
						}
					}
					// Depending on how service filters tree, child might not be there if parent doesn't match
					// The current service logic fetches all that match, then builds tree.
					// So if "Child 1 of Root 1" name doesn't match "Root 1", it won't be fetched.
					// If "Root 1" name matches parent, its children are included if they were fetched.
					// Test assumes child1's name "Child 1 of Root 1" also matches "%Root 1%"
					t.Assert(foundChildInFiltered, true, "Child 1 should be under filtered Root Dept 1")
				}
			}
		})

		t.Run("List with name filter (specific child name)", func(t *gtest.T) {
			if child1 == nil {
				t.Log("Skipping specific child name filter test as child1 was not created.")
				return
			}
			req := &controller.DepartmentListApiReq{Name: "Child 1"} // Filter for child's name
			res, err := deptCtrl.List(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.Assert(len(res.List) > 0, "Should find at least one item (Child 1, possibly as a root if parent doesn't match)")

			// The service builds tree from all matched items. If only child matches, it becomes a root.
			foundChild1AsRoot := false
			for _, item := range res.List {
				if item.Id == child1.Id {
					foundChild1AsRoot = true
					break
				}
			}
			t.Assert(foundChild1AsRoot, true, "Child 1 should be found, possibly as a root item")
		})


		t.Run("List with status filter (active)", func(t *gtest.T) {
			activeStatus := 1
			req := &controller.DepartmentListApiReq{Status: &activeStatus}
			res, err := deptCtrl.List(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.Assert(len(res.List) >= 2, "Should find at least 2 active root departments") // root1, root2
			for _, item := range res.List {
				t.Assert(*item.Status, 1, fmt.Sprintf("Department %s should be active", item.Name))
				for _, child := range item.Children {
					t.Assert(*child.Status, 1, fmt.Sprintf("Child Department %s should be active", child.Name))
				}
			}
		})

		t.Run("List with status filter (disabled)", func(t *gtest.T) {
			disabledStatus := 0
			req := &controller.DepartmentListApiReq{Status: &disabledStatus}
			res, err := deptCtrl.List(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			foundDisabled := false
			for _, item := range res.List {
				if item.Id == disabledRoot.Id {
					foundDisabled = true
				}
				// Also check that all items returned *do* have status 0
				t.Assert(*item.Status, 0, fmt.Sprintf("Department %s should be disabled", item.Name))

			}
			t.Assert(foundDisabled, true, "Disabled Root Dept not found when filtering by disabled status")
		})

		t.Run("Validation of Status query parameter", func(t *gtest.T) {
			invalidStatus := 3
			// This validation is on the request struct tag `v:"in:0,1#department.statusInvalid"`
			// For HTTP requests, GoFrame handles this. For direct calls, need to invoke validator.
			req := &controller.DepartmentListApiReq{Status: &invalidStatus}
			err := g.Validator().Data(req).Run(ctx) // Test struct validation
			t.AssertNE(err, nil, "Validation should fail for invalid status value")
			if err != nil {
				gerr := err.(*gerror.Error)
				// Check if the validation error for 'Status' (department.statusInvalid) is present
				validationMap := gerr.Map()
				fieldError, ok := validationMap["Status"]
				t.Assert(ok, true, "Error map should contain Status key")
				t.Assert(fieldError, "department.statusInvalid", "Status validation message mismatch")
			}
		})

		t.Run("Empty list scenario", func(t *gtest.T) {
			// Temporarily delete all test departments to simulate empty list
			cleanupDepartments(ctx, testDeptIds...)
			// Create a new list of IDs to avoid issues with defer
			testDeptIds = []uint{} // Reset for defer

			req := &controller.DepartmentListApiReq{}
			res, err := deptCtrl.List(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.Assert(len(res.List), 0, "List should be empty after deleting all test departments")

			// Re-create one for subsequent tests if any, or rely on defer to clean if test fails mid-way
			// For simplicity, we'll let the main defer handle it if test fails.
			// If more tests followed this, re-add some data.
		})

	})
}

func TestDepartmentController_Delete(t *testing.T) {
	ctx := gctx.New()
	deptCtrl := controller.NewDepartmentController()
	var testDeptIds []uint // Keep track of IDs for cleanup

	// Teardown logic to be run at the end of the test function
	defer func() {
		if len(testDeptIds) > 0 {
			cleanupDepartments(ctx, testDeptIds...)
		}
	}()

	// Helper to add IDs to cleanup list
	trackIdForCleanup := func(id uint) {
		testDeptIds = append(testDeptIds, id)
	}


	gtest.C(t, func(t *gtest.T) {
		// --- Test Case 1: Successfully deleting a department (that has no children) ---
		t.Run("Successfully deleting a department with no children", func(t *gtest.T) {
			status := 1
			deptToDelete, err := createTestDepartment(ctx, "Delete_LeafDept_"+guid.S(), 0, 10, status)
			gtest.AssertNil(t, err)
			if deptToDelete == nil { t.Fatal("Failed to create deptToDelete for success test") }
			trackIdForCleanup(deptToDelete.Id) // Track for cleanup just in case test fails before explicit delete call

			req := &controller.DepartmentDeleteApiReq{Id: deptToDelete.Id}
			res, err := deptCtrl.Delete(ctx, req)
			t.AssertNil(err, "Delete operation should not fail for leaf department")
			t.AssertNE(res, nil, "Response should not be nil on successful delete")

			// Verify it's actually gone
			getReq := &controller.DepartmentGetByIdReq{Id: deptToDelete.Id}
			_, getErr := deptCtrl.GetById(ctx, getReq)
			t.AssertNE(getErr, nil, "Department should not be found after deletion")
			t.Assert(gerror.Code(getErr), gcode.CodeNotFound, "Error code should be NotFound after deletion")
		})

		// --- Test Case 2: Attempting to delete a non-existent department ---
		t.Run("Attempting to delete a non-existent department", func(t *gtest.T) {
			nonExistentId := uint(999888)
			req := &controller.DepartmentDeleteApiReq{Id: nonExistentId}
			res, err := deptCtrl.Delete(ctx, req)

			t.AssertNil(res, "Response should be nil for failed deletion")
			t.AssertNE(err, nil, "Error should not be nil for non-existent department")
			t.Assert(gerror.Code(err), gcode.CodeNotFound, "Error code should be NotFound")
		})

		// --- Test Case 3: Attempting to delete a department that has child departments ---
		t.Run("Attempting to delete a department with children", func(t *gtest.T) {
			status := 1
			parentDept, err := createTestDepartment(ctx, "Delete_ParentDept_"+guid.S(), 0, 20, status)
			gtest.AssertNil(t, err)
			if parentDept == nil { t.Fatal("Failed to create parentDept") }
			trackIdForCleanup(parentDept.Id)

			childDept, err := createTestDepartment(ctx, "Delete_ChildDept_"+guid.S(), parentDept.Id, 1, status)
			gtest.AssertNil(t, err)
			if childDept == nil { t.Fatal("Failed to create childDept") }
			trackIdForCleanup(childDept.Id)

			req := &controller.DepartmentDeleteApiReq{Id: parentDept.Id}
			res, err := deptCtrl.Delete(ctx, req)

			t.AssertNil(res, "Response should be nil when deletion is prevented")
			t.AssertNE(err, nil, "Error should not be nil when department has children")
			t.Assert(gerror.Code(err), gcode.CodeBusinessValidationFailed, "Error code should be BusinessValidationFailed")
			t.AssertContains(err.Error(), g.I18n().T(ctx, "department.hasChildrenCannotDelete"), "Error message mismatch for hasChildren")
		})

		// --- Test Case 4: Validation errors for ID in the path (e.g., ID = 0) ---
		t.Run("Validation error for ID in path (ID = 0)", func(t *gtest.T) {
			req := &controller.DepartmentDeleteApiReq{Id: 0}
			// Direct validation of request struct, as framework handles this for HTTP
			err := g.Validator().Data(req).Run(ctx)
			t.AssertNE(err, nil, "Validation should fail for ID=0")
			// Example: t.Assert(err.(*gerror.Error).Map()["Id"], "department.idMin")
		})
	})
}

func TestDepartmentController_Update(t *testing.T) {
	ctx := gctx.New()
	deptCtrl := controller.NewDepartmentController()
	var testDeptIds []uint // Keep track of IDs for cleanup

	// Helper to create a department specifically for these update tests
	createForUpdateTest := func(nameSuffix string, parentId uint) *entity.Department {
		status := 1
		dept, err := createTestDepartment(ctx, "UpdateTest_"+nameSuffix, parentId, 0, status)
		gtest.AssertNil(t, err, "Failed to create dept for update test: "+nameSuffix)
		if dept != nil {
			testDeptIds = append(testDeptIds, dept.Id)
		} else {
			t.Fatalf("Failed to create department %s, cannot proceed", nameSuffix)
		}
		return dept
	}

	// Teardown
	defer cleanupDepartments(ctx, testDeptIds...)


	gtest.C(t, func(t *gtest.T) {
		// --- Setup initial departments for various test cases ---
		deptToUpdate := createForUpdateTest("Target", 0)
		parentCandidate1 := createForUpdateTest("Parent1", 0)
		parentCandidate2 := createForUpdateTest("Parent2", 0)
		childOfTarget := createForUpdateTest("ChildOfTarget", deptToUpdate.Id)
		anotherRoot := createForUpdateTest("AnotherRoot", 0) // For name conflict test

		t.Run("Successfully updating department name", func(t *gtest.T) {
			newName := "Updated Dept Name " + guid.S()
			req := &controller.DepartmentUpdateApiReq{
				Id: deptToUpdate.Id,
				DepartmentUpdateInput: service.DepartmentUpdateInput{Name: &newName},
			}
			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.AssertNE(res.Department, nil)
			t.Assert(res.Department.Id, deptToUpdate.Id)
			t.Assert(res.Department.Name, newName)
		})

		t.Run("Successfully updating parent ID", func(t *gtest.T) {
			req := &controller.DepartmentUpdateApiReq{
				Id: deptToUpdate.Id,
				DepartmentUpdateInput: service.DepartmentUpdateInput{ParentId: &parentCandidate1.Id},
			}
			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.Assert(res.Department.ParentId, parentCandidate1.Id)
		})

		t.Run("Successfully updating sort, status, remark", func(t *gtest.T) {
			newSort := 50
			newStatus := 0
			newRemark := "Updated remark " + guid.S()
			req := &controller.DepartmentUpdateApiReq{
				Id: deptToUpdate.Id,
				DepartmentUpdateInput: service.DepartmentUpdateInput{
					Sort:   &newSort,
					Status: &newStatus,
					Remark: &newRemark,
				},
			}
			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.Assert(res.Department.Sort, newSort)
			t.Assert(res.Department.Status, &newStatus)
			t.Assert(res.Department.Remark, newRemark)
		})


		t.Run("Attempting to update a non-existent department", func(t *gtest.T) {
			nonExistentId := uint(888888)
			name := "No Dept"
			req := &controller.DepartmentUpdateApiReq{
				Id: nonExistentId,
				DepartmentUpdateInput: service.DepartmentUpdateInput{Name: &name},
			}
			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(res)
			t.AssertNE(err, nil)
			t.Assert(gerror.Code(err), gcode.CodeNotFound)
		})

		t.Run("Validation error for ID in path (ID = 0)", func(t *gtest.T) {
			name := "Invalid ID Test"
			req := &controller.DepartmentUpdateApiReq{
				Id: 0, // Invalid
				DepartmentUpdateInput: service.DepartmentUpdateInput{Name: &name},
			}
			err := g.Validator().Data(req).Run(ctx) // Validate the req struct
			t.AssertNE(err, nil, "Validation should fail for ID=0")
			// Example: t.Assert(err.(*gerror.Error).Map()["Id"], "department.idMin")
		})

		t.Run("Validation error for body field (e.g., name too long if rule exists)", func(t *gtest.T) {
			longName := string(make([]byte, 101)) // Assuming length:1,100 for name
			reqBody := service.DepartmentUpdateInput{Name: &longName}
			err := g.Validator().Data(reqBody).Run(ctx) // Validate the input DTO part
			t.AssertNE(err, nil, "Validation should fail for long name")
			// Example: t.Assert(err.(*gerror.Error).Map()["Name"], "department.nameLength")
		})

		t.Run("Business logic: duplicate name under the same parent", func(t *gtest.T) {
			// Make deptToUpdate have same name as anotherRoot (both under root parentId=0)
			nameOfAnotherRoot := anotherRoot.Name
			req := &controller.DepartmentUpdateApiReq{
				Id: deptToUpdate.Id, // Try to update deptToUpdate
				DepartmentUpdateInput: service.DepartmentUpdateInput{
					Name:     &nameOfAnotherRoot, // to this name which exists for anotherRoot
					ParentId: g.Ptr(uint(0)),     // ensure it's also at root for direct comparison
				},
			}
			// First, ensure deptToUpdate is also at root if it's not already
			if deptToUpdate.ParentId != 0 {
				_, err := deptCtrl.Update(ctx, &controller.DepartmentUpdateApiReq{
					Id: deptToUpdate.Id,
					DepartmentUpdateInput: service.DepartmentUpdateInput{ParentId: g.Ptr(uint(0))},
				})
				t.AssertNil(err, "Setup: Failed to move deptToUpdate to root for duplicate name test")
			}

			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(res)
			t.AssertNE(err, nil)
			t.Assert(gerror.Code(err), gcode.CodeBusinessValidationFailed)
			t.AssertContains(err.Error(), g.I18n().Tf(ctx, "department.nameTakenInParent", nameOfAnotherRoot))
		})

		t.Run("Business logic: cannot set self as parent", func(t *gtest.T) {
			req := &controller.DepartmentUpdateApiReq{
				Id: deptToUpdate.Id,
				DepartmentUpdateInput: service.DepartmentUpdateInput{ParentId: &deptToUpdate.Id},
			}
			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(res)
			t.AssertNE(err, nil)
			t.Assert(gerror.Code(err), gcode.CodeBusinessValidationFailed)
			t.AssertContains(err.Error(), g.I18n().T(ctx, "department.cannotSetSelfAsParent"))
		})

		t.Run("Business logic: cannot move to a non-existent parent", func(t *gtest.T) {
			nonExistentParentId := uint(777777)
			req := &controller.DepartmentUpdateApiReq{
				Id: deptToUpdate.Id,
				DepartmentUpdateInput: service.DepartmentUpdateInput{ParentId: &nonExistentParentId},
			}
			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(res)
			t.AssertNE(err, nil)
			t.Assert(gerror.Code(err), gcode.CodeBusinessValidationFailed)
			t.AssertContains(err.Error(), g.I18n().Tf(ctx, "department.parentNotFound", nonExistentParentId))
		})

		t.Run("Business logic: cannot move parent to its own child (circular dependency)", func(t *gtest.T) {
			// Try to make deptToUpdate (parent) a child of childOfTarget
			req := &controller.DepartmentUpdateApiReq{
				Id: deptToUpdate.Id,
				DepartmentUpdateInput: service.DepartmentUpdateInput{ParentId: &childOfTarget.Id},
			}
			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(res)
			t.AssertNE(err, nil)
			t.Assert(gerror.Code(err), gcode.CodeBusinessValidationFailed)
			t.AssertContains(err.Error(), g.I18n().T(ctx, "department.cannotMoveToDescendant"))
		})

		t.Run("No actual changes submitted", func(t *gtest.T) {
			// Submit an update request with no actual data change in the input
			// The service layer's UpdateDepartment should return nil error if no fields to update.
			req := &controller.DepartmentUpdateApiReq{
				Id: deptToUpdate.Id,
				DepartmentUpdateInput: service.DepartmentUpdateInput{}, // Empty input
			}
			res, err := deptCtrl.Update(ctx, req)
			t.AssertNil(err, "Error should be nil if no changes are made")
			t.AssertNE(res, nil, "Response should not be nil even if no changes")
			t.Assert(res.Department.Id, deptToUpdate.Id, "Department ID should match")
			// Fetch current state to compare
			currentState, _ := service.NewDepartmentService().GetDepartmentById(ctx, deptToUpdate.Id)
			t.Assert(res.Department.Name, currentState.Name, "Name should be unchanged")

		})

	})
}

func TestDepartmentController_GetById(t *testing.T) {
	ctx := gctx.New()
	deptCtrl := controller.NewDepartmentController()
	var testDeptIds []uint // Keep track of IDs for cleanup

	// Setup: Create a test department
	status := 1
	testDept, err := createTestDepartment(ctx, "GetById Test Dept", 0, 100, status)
	gtest.AssertNil(t, err, "Failed to create department for GetById test")
	if testDept != nil {
		testDeptIds = append(testDeptIds, testDept.Id)
	} else {
		t.Fatal("Test department creation returned nil")
	}

	// Teardown
	defer cleanupDepartments(ctx, testDeptIds...)

	gtest.C(t, func(t *gtest.T) {
		t.Run("Successfully finding a department", func(t *gtest.T) {
			req := &controller.DepartmentGetByIdReq{Id: testDept.Id}
			res, err := deptCtrl.GetById(ctx, req)

			t.AssertNil(err)
			t.AssertNE(res, nil)
			t.AssertNE(res.Department, nil)
			t.Assert(res.Department.Id, testDept.Id)
			t.Assert(res.Department.Name, testDept.Name)
		})

		t.Run("Department not found", func(t *gtest.T) {
			nonExistentId := uint(999999)
			req := &controller.DepartmentGetByIdReq{Id: nonExistentId}
			res, err := deptCtrl.GetById(ctx, req)

			t.AssertNil(res)
			t.AssertNE(err, nil)
			gerr := err.(*gerror.Error)
			t.Assert(gerr.Code(), gcode.CodeNotFound)
			// Check for i18n message (service layer is responsible for this)
			// Example: t.AssertContains(err.Error(), g.I18n().Tf(ctx, "department.notFoundId", nonExistentId))
			// For now, just check it contains the ID, as i18n setup for test might not be complete.
			t.AssertContains(err.Error(), fmt.Sprintf("%d", nonExistentId))
		})

		t.Run("Invalid ID in request path (ID = 0)", func(t *gtest.T) {
			req := &controller.DepartmentGetByIdReq{Id: 0}
			// Test struct validation (GoFrame should catch this based on `v` tags)
			err := g.Validator().Data(req).Run(ctx)
			t.AssertNE(err, nil, "Validation should fail for ID=0")

			// If validation passes and goes to service (it shouldn't for HTTP, but for direct call):
			// res, serviceErr := deptCtrl.GetById(ctx, req)
			// t.AssertNil(res)
			// t.AssertNE(serviceErr, nil)
			// gerr := serviceErr.(*gerror.Error)
			// t.Assert(gerr.Code(), gcode.CodeInvalidArgument) // Service returns this for ID 0
			// t.AssertContains(serviceErr.Error(), g.I18n().T(ctx, "department.idInvalid"))
		})
	})
}
```
