package controller

import (
	"context"
	"yuncms/internal/app/model"
	"yuncms/internal/app/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type UserController struct{}

func NewUser() *UserController {
	return &UserController{}
}

// UserCreateApiInput is the DTO for API request for creating a user.
type UserCreateApiInput struct {
	Username string `json:"username" v:"required|length:3,30#user_username_required|user_username_length"`
	Password string `json:"password" v:"required|length:6,30#user_password_required|user_password_length"`
	Nickname string `json:"nickname" v:"required|length:1,30#user_nickname_required|user_nickname_length"`
	Email    string `json:"email"    v:"email#user_email_format"`
	Phone    string `json:"phone"    v:"phone-loose#user_phone_format"`
	Avatar   string `json:"avatar"   v:"url#user_avatar_url"`
	Status   *int   `json:"status"   v:"in:0,1#user_status_invalid"` // Using pointer to allow explicit 0 or 1, or nil for default in service
}

// UserCreateReq is the request structure for creating a user, including API metadata.
type UserCreateReq struct {
	g.Meta `path:"/users" method:"post" summary:"Create a new user" tags:"User Management"`
	UserCreateApiInput
}

// UserCreateRes is the response structure after creating a user.
type UserCreateRes struct {
	*model.User // Embed the user model, password will be excluded by json:"-" tag
}

// Create handles the HTTP POST request to create a new user.
func (c *UserController) Create(ctx context.Context, req *UserCreateReq) (res *UserCreateRes, err error) {
	// Convert API input to service input
	serviceInput := service.UserCreateInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   req.Status, // This is *int, service layer handles defaulting if nil
	}

	createdUser, err := service.NewUserService().CreateUser(ctx, serviceInput)
	if err != nil {
		// Check for specific business errors if needed, otherwise let default error handling work
		// Example: if gerror.Code(err) == gcode.BusinessValidationFailed { ... }
		// GoFrame's default error handler will use validation messages for gvalid.Error types
		return nil, err
	}

	// If successful, return HTTP 201 (Created)
	// GoFrame's default success response is HTTP 200.
	// To explicitly set 201, you might need to manipulate ghttp.Request from ctx if available,
	// or rely on a middleware. For now, we'll return the data, and default is 200.
	// A common pattern is to just return the created resource.
	// g.RequestFromCtx(ctx).Response.Status = http.StatusCreated // Example if direct response manipulation is needed

	return &UserCreateRes{User: createdUser}, nil
}
