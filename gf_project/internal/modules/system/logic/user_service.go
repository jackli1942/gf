package logic

import (
	"context"
	"fmt"
	"strconv" // For converting uint64 to string for Casbin subject

	"gf_project/internal/modules/system/model/do"
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"gf_project/internal/modules/system/model/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"golang.org/x/crypto/bcrypt"
)

type sSystemUser struct{}

var UserService = sSystemUser{}

const (
	DefaultPasswordCost = 10
)

// hashPassword generates a bcrypt hash for the given password.
func (s *sSystemUser) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultPasswordCost)
	if err != nil {
		return "", gerror.Wrap(err, "failed to generate password hash")
	}
	return string(hashedBytes), nil
}

// verifyPassword compares a plaintext password with a stored bcrypt hash.
func (s *sSystemUser) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// CreateUser creates a new user.
func (s *sSystemUser) CreateUser(ctx context.Context, in *input.UserCreateInp) (userId int64, err error) {
	count, err := dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Username, in.Username).Count()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to check username existence")
	}
	if count > 0 {
		return 0, gerror.Newf("username '%s' already exists", in.Username)
	}

	if in.Email != "" {
		count, err = dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Email, in.Email).Count()
		if err != nil {
			return 0, gerror.Wrap(err, "failed to check email existence")
		}
		if count > 0 {
			return 0, gerror.Newf("email '%s' already exists", in.Email)
		}
	}
	if in.Mobile != "" {
		count, err = dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Mobile, in.Mobile).Count()
		if err != nil {
			return 0, gerror.Wrap(err, "failed to check mobile existence")
		}
		if count > 0 {
			return 0, gerror.Newf("mobile number '%s' already exists", in.Mobile)
		}
	}

	hashedPassword, err := s.hashPassword(in.Password)
	if err != nil {
		return 0, err
	}

	var deptIdValue interface{}
	if in.DeptId > 0 {
		deptIdValue = in.DeptId
	}
	var genderValue interface{}
	if in.Gender != nil {
		genderValue = *in.Gender
	}

	newUser := do.SystemUser{
		Username:     in.Username,
		PasswordHash: hashedPassword,
		Nickname:     in.Nickname,
		Email:        in.Email,
		Mobile:       in.Mobile,
		DeptId:       deptIdValue,
		AvatarUrl:    in.Avatar,
		Gender:       genderValue,
		Status:       in.Status,
		Remark:       in.Remark,
		CreatedAt:    gtime.Now(),
		UpdatedAt:    gtime.Now(),
	}

	result, err := dao.SystemUser.Ctx(ctx).Data(newUser).Insert()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to create user")
	}
	newUserId, err := result.LastInsertId()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to retrieve last insert ID for user")
	}
	return newUserId, nil
}

// UserLogin performs user login.
func (s *sSystemUser) UserLogin(ctx context.Context, in *input.UserLoginInp) (user *entity.SystemUser, err error) {
	err = dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Username, in.Username).Scan(&user)
	if err != nil {
		if err == gdb.ErrNoRows {
			return nil, gerror.New("invalid username or password")
		}
		return nil, gerror.Wrap(err, "login query failed")
	}
	if user == nil {
		return nil, gerror.New("invalid username or password")
	}

	if err = s.verifyPassword(user.PasswordHash, in.Password); err != nil {
		return nil, gerror.New("invalid username or password")
	}

	if user.Status == 1 {
		return nil, gerror.New("user account is disabled")
	}

	_, updateErr := dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Id, user.Id).Data(do.SystemUser{
		LoginTime: gtime.Now(),
	}).Update()
	if updateErr != nil {
		g.Log().Warningf(ctx, "failed to update user login time for user ID %d: %v", user.Id, updateErr)
	}
	return user, nil
}

// GetUserById retrieves a user by their ID.
func (s *sSystemUser) GetUserById(ctx context.Context, id uint64) (user *entity.SystemUser, err error) {
	err = dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Id, id).Scan(&user)
	if err != nil {
		if err == gdb.ErrNoRows {
			return nil, gerror.Newf("user with ID %d not found", id)
		}
		return nil, gerror.Wrapf(err, "failed to retrieve user by ID %d", id)
	}
	return user, nil
}

// GetUserList retrieves a paginated list of users.
func (s *sSystemUser) GetUserList(ctx context.Context, in *input.UserListInp) (list []*entity.SystemUser, total int, err error) {
	m := dao.SystemUser.Ctx(ctx).OmitEmptyWhere()
	if in.Username != "" {
		m = m.WhereLike(dao.SystemUser.Columns().Username, "%"+in.Username+"%")
	}
	if in.Mobile != "" {
		m = m.WhereLike(dao.SystemUser.Columns().Mobile, "%"+in.Mobile+"%")
	}
	if in.Status != nil {
		m = m.Where(dao.SystemUser.Columns().Status, *in.Status)
	}
	if in.DeptId != nil && *in.DeptId > 0 {
		m = m.Where(dao.SystemUser.Columns().DeptId, *in.DeptId)
	}

	total, err = m.Count()
	if err != nil {
		return nil, 0, gerror.Wrap(err, "failed to count users")
	}
	if total == 0 {
		return []*entity.SystemUser{}, 0, nil
	}

	err = m.Page(in.Page, in.PageSize).OrderDesc(dao.SystemUser.Columns().CreatedAt).Scan(&list)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "failed to retrieve user list")
	}
	return list, total, nil
}

// UpdateUser updates an existing user's information.
func (s *sSystemUser) UpdateUser(ctx context.Context, in *input.UserUpdateInp) (err error) {
	currentUser, err := s.GetUserById(ctx, in.Id)
	if err != nil {
		return err
	}

	if in.Email != "" && in.Email != currentUser.Email {
		count, err := dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Email, in.Email).AndNot(dao.SystemUser.Columns().Id, in.Id).Count()
		if err != nil {
			return gerror.Wrap(err, "failed to check email uniqueness for update")
		}
		if count > 0 {
			return gerror.Newf("email '%s' already in use by another account", in.Email)
		}
	}
	if in.Mobile != "" && in.Mobile != currentUser.Mobile {
		count, err := dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Mobile, in.Mobile).AndNot(dao.SystemUser.Columns().Id, in.Id).Count()
		if err != nil {
			return gerror.Wrap(err, "failed to check mobile uniqueness for update")
		}
		if count > 0 {
			return gerror.Newf("mobile number '%s' already in use by another account", in.Mobile)
		}
	}

	var deptIdValue interface{}
	if in.DeptId > 0 {
		deptIdValue = in.DeptId
	}
	var genderValue interface{}
	if in.Gender != nil {
		genderValue = *in.Gender
	}

	updateData := do.SystemUser{
		Nickname: in.Nickname, Email: in.Email, Mobile: in.Mobile, DeptId: deptIdValue,
		AvatarUrl: in.Avatar, Gender: genderValue, Status: in.Status, Remark: in.Remark,
		UpdatedAt: gtime.Now(),
	}
	_, err = dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Id, in.Id).Data(updateData).Update()
	if err != nil {
		return gerror.Wrapf(err, "failed to update user ID %d", in.Id)
	}
	return nil
}

// DeleteUser performs a soft delete on a user and removes associated Casbin policies.
func (s *sSystemUser) DeleteUser(ctx context.Context, id uint64) (err error) {
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Soft delete user
		_, err := dao.SystemUser.Ctx(ctx).TX(tx).Where(dao.SystemUser.Columns().Id, id).
			Data(g.Map{dao.SystemUser.Columns().DeletedAt: gtime.Now()}).Update()
		if err != nil {
			return gerror.Wrapf(err, "failed to soft delete user ID %d", id)
		}

		// Remove user-role associations from DB
		_, err = dao.SystemUserRole.Ctx(ctx).TX(tx).Where(dao.SystemUserRole.Columns().UserId, id).Delete()
		if err != nil {
			return gerror.Wrapf(err, "failed to delete user-role associations for user ID %d", id)
		}

		e, casbinErr := CasbinEnforcer() // Assuming CasbinEnforcer is defined in casbin_service.go
		if casbinErr != nil {
			return gerror.Wrap(casbinErr, "failed to get Casbin enforcer for deleting user policies")
		}

		userIdStr := strconv.FormatUint(id, 10)
		// Remove all roles associated with the user (e.g. g(userId, roleId))
		if _, casbinErr = e.RemoveFilteredGroupingPolicy(0, userIdStr); casbinErr != nil {
			return gerror.Wrapf(casbinErr, "failed to remove Casbin grouping policies for user ID %s", userIdStr)
		}
		// Optionally, remove direct permissions if your model supports p(userId, obj, act)
		// if _, casbinErr = e.RemoveFilteredPolicy(0, userIdStr); casbinErr != nil {
		//    return gerror.Wrapf(casbinErr, "failed to remove Casbin direct policies for user ID %s", userIdStr)
		// }
		// if errSave := e.SavePolicy(); errSave != nil { // Adapter might auto-save
		// 	return gerror.Wrap(errSave, "failed to save policy after deleting user roles")
		// }
		return nil
	})
	return err
}

// ChangeUserPassword allows a user to change their own password.
func (s *sSystemUser) ChangeUserPassword(ctx context.Context, userId uint64, oldPassword, newPassword string) error {
	user := &entity.SystemUser{}
	err := dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Id, userId).Scan(&user)
	if err != nil {
		if err == gdb.ErrNoRows {
			return gerror.New("user not found")
		}
		return gerror.Wrapf(err, "failed to find user ID %d for password change", userId)
	}

	if err := s.verifyPassword(user.PasswordHash, oldPassword); err != nil {
		return gerror.New("incorrect old password")
	}

	hashedNewPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Id, userId).Data(do.SystemUser{
		PasswordHash: hashedNewPassword, UpdatedAt: gtime.Now(),
	}).Update()
	if err != nil {
		return gerror.Wrapf(err, "failed to change password for user ID %d", userId)
	}
	return nil
}

// AdminResetPassword allows an admin to reset a user's password.
func (s *sSystemUser) AdminResetPassword(ctx context.Context, userId uint64, newPassword string) error {
	hashedNewPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Id, userId).Data(do.SystemUser{
		PasswordHash: hashedNewPassword, UpdatedAt: gtime.Now(),
	}).Update()
	if err != nil {
		return gerror.Wrapf(err, "failed to reset password by admin for user ID %d", userId)
	}
	return nil
}

// UpdateUserProfile updates the current user's profile information.
func (s *sSystemUser) UpdateUserProfile(ctx context.Context, userId uint64, in *input.UserProfileUpdateInp) error {
	currentUser, err := s.GetUserById(ctx, userId)
	if err != nil {
		return err
	}

	if in.Email != "" && in.Email != currentUser.Email {
		count, err := dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Email, in.Email).AndNot(dao.SystemUser.Columns().Id, userId).Count()
		if err != nil {
			return gerror.Wrap(err, "failed to check email uniqueness for profile update")
		}
		if count > 0 {
			return gerror.Newf("email '%s' already in use by another account", in.Email)
		}
	}
	if in.Mobile != "" && in.Mobile != currentUser.Mobile {
		count, err := dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Mobile, in.Mobile).AndNot(dao.SystemUser.Columns().Id, userId).Count()
		if err != nil {
			return gerror.Wrap(err, "failed to check mobile uniqueness for profile update")
		}
		if count > 0 {
			return gerror.Newf("mobile number '%s' already in use by another account", in.Mobile)
		}
	}

	var genderValue interface{}
	if in.Gender != nil {
		genderValue = *in.Gender
	}

	updateData := g.Map{
		dao.SystemUser.Columns().Nickname:  in.Nickname,
		dao.SystemUser.Columns().Email:     in.Email,
		dao.SystemUser.Columns().Mobile:    in.Mobile,
		dao.SystemUser.Columns().AvatarUrl: in.Avatar,
		dao.SystemUser.Columns().Gender:    genderValue,
		dao.SystemUser.Columns().Signed:    in.Signed,
		dao.SystemUser.Columns().UpdatedAt: gtime.Now(),
	}

	_, err = dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Id, userId).Data(updateData).Update()
	if err != nil {
		return gerror.Wrapf(err, "failed to update profile for user ID %d", userId)
	}
	return nil
}

// UpdateUserStatus updates a user's status (active/disabled).
func (s *sSystemUser) UpdateUserStatus(ctx context.Context, id uint64, status uint) error {
	_, err := dao.SystemUser.Ctx(ctx).Where(dao.SystemUser.Columns().Id, id).Data(do.SystemUser{
		Status: status, UpdatedAt: gtime.Now(),
	}).Update()
	if err != nil {
		return gerror.Wrapf(err, "failed to update status for user ID %d", id)
	}
	return nil
}

// UpdateUserRoles assigns a list of roles to a user.
func (s *sSystemUser) UpdateUserRoles(ctx context.Context, in *input.UserAssignRoleInp) error {
	userIdStr := strconv.FormatUint(in.UserId, 10)

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var currentRoleEntities []*entity.SystemRole
		// Fetch current roles to know their codes/keys for Casbin
		roleTable := dao.SystemRole.Table()
		userRoleTable := dao.SystemUserRole.Table()
		joinOn := fmt.Sprintf("%s.%s = %s.%s", roleTable, dao.SystemRole.Columns().Id, userRoleTable, dao.SystemUserRole.Columns().RoleId)

		err := dao.SystemUserRole.Ctx(ctx).TX(tx).
			LeftJoin(roleTable, joinOn).
			Where(fmt.Sprintf("%s.%s", userRoleTable, dao.SystemUserRole.Columns().UserId), in.UserId).
			Fields(fmt.Sprintf("%s.*", roleTable)). // Select all fields from system_role table
			Scan(&currentRoleEntities)
		if err != nil {
			return gerror.Wrapf(err, "failed to retrieve current roles for user ID %d", in.UserId)
		}

		_, err = dao.SystemUserRole.Ctx(ctx).TX(tx).Where(dao.SystemUserRole.Columns().UserId, in.UserId).Delete()
		if err != nil {
			return gerror.Wrapf(err, "failed to clear existing roles for user ID %d from DB", in.UserId)
		}

		e, casbinErr := CasbinEnforcer()
		if casbinErr != nil {
			return gerror.Wrap(casbinErr, "failed to get Casbin enforcer for updating roles")
		}
		for _, role := range currentRoleEntities {
			if role.Code != "" { // Using role code/key
				if _, err = e.RemoveGroupingPolicy(userIdStr, role.Code); err != nil {
					g.Log().Warningf(ctx, "Failed to remove Casbin policy g(%s, %s): %v", userIdStr, role.Code, err)
				}
			}
		}

		if len(in.RoleIds) > 0 {
			newRolesData := g.List{}
			for _, roleId := range in.RoleIds {
				newRolesData = append(newRolesData, do.SystemUserRole{
					UserId: in.UserId,
					RoleId: roleId,
				})
			}
			_, err = dao.SystemUserRole.Ctx(ctx).TX(tx).Data(newRolesData).Insert()
			if err != nil {
				return gerror.Wrapf(err, "failed to insert new roles for user ID %d into DB", in.UserId)
			}

			var newRoleEntitiesForCasbin []*entity.SystemRole
			err = dao.SystemRole.Ctx(ctx).TX(tx).WhereIn(dao.SystemRole.Columns().Id, in.RoleIds).Scan(&newRoleEntitiesForCasbin)
			if err != nil {
				return gerror.Wrapf(err, "failed to retrieve new role codes for user ID %d for Casbin", in.UserId)
			}

			for _, role := range newRoleEntitiesForCasbin {
				if role.Code != "" {
					if _, err = e.AddGroupingPolicy(userIdStr, role.Code); err != nil {
						g.Log().Warningf(ctx, "Failed to add Casbin policy g(%s, %s): %v", userIdStr, role.Code, err)
					}
				}
			}
		}

		// The hailaz adapter might auto-save on Add/RemoveGroupingPolicy.
		// If not, explicit SavePolicy() would be needed here.
		// if errSave := e.SavePolicy(); errSave != nil {
		// 	return gerror.Wrap(errSave, "failed to save Casbin policy after updating user roles")
		// }
		return nil
	})
}

// GetUserRoleCodes retrieves the role codes for a given user ID.
func (s *sSystemUser) GetUserRoleCodes(ctx context.Context, userId uint64) ([]string, error) {
	var roles []*entity.SystemRole
	// Construct the join condition carefully
	joinTable := dao.SystemRole.Table()
	mainTableAlias := "ur" // Alias for system_user_role
	joinOn := fmt.Sprintf("%s.%s = %s.%s",
		mainTableAlias, dao.SystemUserRole.Columns().RoleId,
		joinTable, dao.SystemRole.Columns().Id)

	err := dao.SystemUserRole.Ctx(ctx).As(mainTableAlias).
		LeftJoin(joinTable, joinOn).
		Where(fmt.Sprintf("%s.%s", mainTableAlias, dao.SystemUserRole.Columns().UserId), userId).
		Fields(fmt.Sprintf("%s.%s", joinTable, dao.SystemRole.Columns().Code)). // Select only role code
		Scan(&roles)

	if err != nil {
		// Do not treat ErrNoRows as a critical error here, user might just have no roles.
		if errors.Is(err, gdb.ErrNoRows) {
			return []string{}, nil
		}
		return nil, gerror.Wrapf(err, "failed to retrieve roles for user ID %d", userId)
	}

	if len(roles) == 0 {
		return []string{}, nil // No roles found for user
	}

	roleCodes := make([]string, 0, len(roles)) // Initialize with 0 length, capacity len(roles)
	for _, role := range roles {
		if role != nil && role.Code != "" { // Ensure role and role code are not nil/empty
			roleCodes = append(roleCodes, role.Code)
		}
	}
	return roleCodes, nil
}
