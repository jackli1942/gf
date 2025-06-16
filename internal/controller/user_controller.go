package controller

import (
	// "context" // Not strictly needed for every method if r.Context() is used
	"yuncms/internal/model/input"
	// "yuncms/internal/model/output" // Service layer returns these, controller just passes them
	"yuncms/internal/service"
    "yuncms/internal/utility/i18nutil"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	// "github.com/gogf/gf/v2/util/gconv" // Not used here
    // "github.com/gogf/gf/v2/errors/gerror" // For type checking errors if needed
)

type UserController struct {
	userService service.IUserService
}

func NewUserController() *UserController {
	return &UserController{userService: service.NewUserService()}
}

// Create handles user creation.
func (c *UserController) Create(r *ghttp.Request) {
	var in input.UserCreateInput
	if err := r.Parse(&in); err != nil {
        // Use i18nutil for generic invalid input, provide err.Error() as Reason
		r.Response.WriteJsonExit(g.Map{"error": i18nutil.T(r.Context(), "error_invalid_input", "Reason", err.Error())})
		return
	}
    // Use GoFrame validation, errors should use i18n keys if defined in struct tags
    if err := g.Validator().Data(in).Run(r.Context()); err != nil {
        firstErr := err.FirstError()
        // Attempt to translate the error string directly (which might be a key like "error_username_required")
        // If no translation, use the original error message from validator.
        translatedMsg := i18nutil.T(r.Context(), firstErr.String())
        if translatedMsg == firstErr.String() { // Check if translation happened or key was returned
             // If the "key" from the validator (e.g. "Password is required") isn't in i18n files,
             // we might want to provide a more generic translated error, or ensure all validator messages have keys.
             // For now, falling back to the raw message is acceptable if no direct key match.
             // A better approach for validator messages is to use keys in the struct tags like `v:"required#key_for_required"`
             translatedMsg = firstErr.Current().Error()
        }
        r.Response.WriteJsonExit(g.Map{"error": translatedMsg })
        return
    }

	user, err := c.userService.Create(r.Context(), &in)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"error": err.Error()}) // Errors from logic should already be i18n'd
		return
	}
	r.Response.WriteJson(g.Map{"message": i18nutil.T(r.Context(),"user_created_successfully"), "data": user})
}

// GetById handles fetching a single user.
func (c *UserController) GetById(r *ghttp.Request) {
    id := r.Get("id").Uint64()
    if id == 0 {
        r.Response.WriteJsonExit(g.Map{"error": i18nutil.T(r.Context(), "error_invalid_input", "Reason", "Invalid user ID")})
        return
    }
    user, err := c.userService.GetById(r.Context(), id)
    if err != nil {
        r.Response.WriteJsonExit(g.Map{"error": err.Error()})
        return
    }
    r.Response.WriteJson(user)
}

// Update handles user updates.
func (c *UserController) Update(r *ghttp.Request) {
    id := r.Get("id").Uint64()
     if id == 0 {
        r.Response.WriteJsonExit(g.Map{"error": i18nutil.T(r.Context(), "error_invalid_input", "Reason", "Invalid user ID")})
        return
    }
    var in input.UserUpdateInput
    if err := r.Parse(&in); err != nil {
        r.Response.WriteJsonExit(g.Map{"error": i18nutil.T(r.Context(), "error_invalid_input", "Reason", err.Error())})
        return
    }
    if err := g.Validator().Data(in).Run(r.Context()); err != nil {
        firstErr := err.FirstError()
        translatedMsg := i18nutil.T(r.Context(), firstErr.String())
        if translatedMsg == firstErr.String() {
             translatedMsg = firstErr.Current().Error()
        }
        r.Response.WriteJsonExit(g.Map{"error": translatedMsg})
        return
    }
    err := c.userService.Update(r.Context(), id, &in)
    if err != nil {
        r.Response.WriteJsonExit(g.Map{"error": err.Error()})
        return
    }
    r.Response.WriteJson(g.Map{"message": i18nutil.T(r.Context(), "user_updated_successfully")})
}

// Delete handles user deletion.
func (c *UserController) Delete(r *ghttp.Request) {
    id := r.Get("id").Uint64()
    if id == 0 {
        r.Response.WriteJsonExit(g.Map{"error": i18nutil.T(r.Context(), "error_invalid_input", "Reason", "Invalid user ID")})
        return
    }
    err := c.userService.Delete(r.Context(), id)
    if err != nil {
        r.Response.WriteJsonExit(g.Map{"error": err.Error()})
        return
    }
    r.Response.WriteJson(g.Map{"message": i18nutil.T(r.Context(), "user_deleted_successfully")})
}

// List handles fetching users with pagination.
func (c *UserController) List(r *ghttp.Request) {
    var in input.UserListInput
    if err := r.Parse(&in); err != nil {
        r.Response.WriteJsonExit(g.Map{"error": i18nutil.T(r.Context(), "error_invalid_input", "Reason", err.Error())})
        return
    }
     if err := g.Validator().Data(in).Run(r.Context()); err != nil {
        firstErr := err.FirstError()
        translatedMsg := i18nutil.T(r.Context(), firstErr.String())
        if translatedMsg == firstErr.String() {
             translatedMsg = firstErr.Current().Error()
        }
        r.Response.WriteJsonExit(g.Map{"error": translatedMsg})
        return
    }
    listOutput, err := c.userService.List(r.Context(), &in)
    if err != nil {
        r.Response.WriteJsonExit(g.Map{"error": err.Error()})
        return
    }
    r.Response.WriteJson(listOutput)
}
