package controller

import (
	"context"
	"yuncms/internal/service" // Correct service path

	"github.com/gogf/gf/v2/frame/g"
	// "github.com/gogf/gf/v2/errors/gerror" // If specific controller errors needed
)

// DepartmentController manages department related web requests.
type DepartmentController struct{}

// NewDepartmentController creates and returns a new DepartmentController.
func NewDepartmentController() *DepartmentController {
	return &DepartmentController{}
}

// DepartmentListApiReq defines the request structure for listing departments via API.
type DepartmentListApiReq struct {
	g.Meta `path:"/department/list" method:"get" summary:"List all departments (tree structure)" tags:"Department Management"`
	Name   string `json:"name,omitempty" in:"query" description:"Filter by department name (fuzzy match)"` // omitempty for json, in:query for http
	Status *int   `json:"status,omitempty" in:"query" v:"in:0,1#department.statusInvalid" description:"Filter by status (0:disabled, 1:active)"`
}

// DepartmentListApiRes defines the response structure for listing departments.
type DepartmentListApiRes struct {
	List []*service.DepartmentListItem `json:"list"`
}

// List handles the HTTP GET request to list departments.
func (c *DepartmentController) List(ctx context.Context, req *DepartmentListApiReq) (res *DepartmentListApiRes, err error) {
	// Validation for req.Status is handled by GoFrame based on struct tags.
	// req.Name does not have validation tags beyond being a string.

	serviceInput := service.DepartmentListInput{
		Name:   req.Name,
		Status: req.Status,
	}

	deptList, err := service.NewDepartmentService().ListDepartments(ctx, serviceInput)
	if err != nil {
		// Propagate error from service layer. Service layer should handle error wrapping.
		return nil, err
	}

	res = &DepartmentListApiRes{
		List: deptList,
	}

	return res, nil
}

// DepartmentUpdateApiReq defines the request structure for updating a department.
type DepartmentUpdateApiReq struct {
	g.Meta `path:"/department/{id}" method:"put" summary:"Update department by ID" tags:"Department Management"`
	Id     uint `json:"id" path:"id" v:"required|min:1#department.idRequired|department.idMin"`
	service.DepartmentUpdateInput // Embed the service input DTO for request body fields
}

// DepartmentUpdateApiRes defines the response structure after updating a department.
type DepartmentUpdateApiRes struct {
	*entity.Department
}

// Update handles the HTTP PUT request to update an existing department.
func (c *DepartmentController) Update(ctx context.Context, req *DepartmentUpdateApiReq) (res *DepartmentUpdateApiRes, err error) {
	// Validation of req.Id (path parameter) and embedded DepartmentUpdateInput fields (body)
	// is automatically handled by GoFrame based on struct tags.

	err = service.NewDepartmentService().UpdateDepartment(ctx, req.Id, req.DepartmentUpdateInput)
	if err != nil {
		// Service layer handles specific error codes (NotFound, BusinessValidationFailed with i18n)
		// Propagate the error directly.
		return nil, err
	}

	// If successful, fetch the updated department to return it.
	updatedDept, serviceErr := service.NewDepartmentService().GetDepartmentById(ctx, req.Id)
	if serviceErr != nil {
		// This would be unusual if UpdateDepartment succeeded without error.
		// Log it and potentially return the original error from UpdateDepartment, or a new one.
		g.Log().Errorf(ctx, "UpdateDepartment: Succeeded updating department %d but failed to retrieve it subsequently: %v", req.Id, serviceErr)
		return nil, gerror.Wrapf(serviceErr, "Failed to retrieve department after update (ID: %d)", req.Id)
	}
	if updatedDept == nil { // Should not happen if Update was successful and GetDepartmentById is consistent
		return nil, gerror.NewCodef(gcode.CodeNotFound, g.I18n().Tf(ctx, "department.notFoundId", req.Id))
	}

	return &DepartmentUpdateApiRes{Department: updatedDept}, nil
}

// DepartmentDeleteApiReq defines the request structure for deleting a department by ID.
type DepartmentDeleteApiReq struct {
	g.Meta `path:"/department/{id}" method:"delete" summary:"Delete department by ID" tags:"Department Management"`
	Id     uint `json:"id" path:"id" v:"required|min:1#department.idRequired|department.idMin"`
}

// DepartmentDeleteApiRes defines the response structure after deleting a department.
// Can be empty for a 204 No Content like response from the framework.
type DepartmentDeleteApiRes struct {
	// Message string `json:"message,omitempty"` // Optional: e.g., "Department deleted successfully"
}

// Delete handles the HTTP DELETE request to remove an existing department.
func (c *DepartmentController) Delete(ctx context.Context, req *DepartmentDeleteApiReq) (res *DepartmentDeleteApiRes, err error) {
	// Validation of req.Id is automatically handled by GoFrame based on struct tags.

	err = service.NewDepartmentService().DeleteDepartment(ctx, req.Id)
	if err != nil {
		// Service layer handles specific error codes (NotFound, BusinessValidationFailed for hasChildren)
		// and uses i18n keys. Propagate the error directly.
		return nil, err
	}

	// On successful deletion, return an empty response body.
	// GoFrame default for (res, nil) where res is a struct might be `{}`, with HTTP 200.
	// For a 204, direct response writer manipulation might be needed, or specific framework config.
	return &DepartmentDeleteApiRes{}, nil
}

// DepartmentGetByIdReq defines the request structure for getting a department by ID.
type DepartmentGetByIdReq struct {
	g.Meta `path:"/department/{id}" method:"get" summary:"Get department by ID" tags:"Department Management"`
	Id     uint `json:"id" path:"id" v:"required|min:1#department.idRequired|department.idMin"`
}

// DepartmentGetByIdRes defines the response structure for getting a department by ID.
type DepartmentGetByIdRes struct {
	*entity.Department
}

// GetById handles the HTTP GET request to fetch a single department by its ID.
func (c *DepartmentController) GetById(ctx context.Context, req *DepartmentGetByIdReq) (res *DepartmentGetByIdRes, err error) {
	// Validation of req.Id is automatically handled by GoFrame based on struct tags.

	department, err := service.NewDepartmentService().GetDepartmentById(ctx, req.Id)
	if err != nil {
		// Service layer now returns gcode.CodeNotFound directly with i18n message.
		// It also returns gcode.CodeInvalidArgument for id=0, or wraps other DB errors.
		// So, we can often propagate the error directly.
		// If we want to ensure specific controller-level error structure or logging for other errors:
		// errCode := gerror.Code(err)
		// if errCode == gcode.CodeNotFound || errCode == gcode.CodeInvalidArgument {
		//     return nil, err // Propagate directly as service handles i18n
		// }
		// return nil, gerror.Wrapf(err, "Failed to get department with ID %d", req.Id)
		return nil, err // Propagate error from service
	}
	// The service layer should not return (nil, nil) anymore based on the updated GetDepartmentById logic.
	// It returns an error if department is nil.

	res = &DepartmentGetByIdRes{
		Department: department,
	}
	return res, nil
}
```
