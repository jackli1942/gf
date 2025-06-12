package service

import (
	"sync"
	"time"

	// "github.com/casbin/casbin/v2" // Removed unused import alias
	dobytecasbin "github.com/dobyte/gf-casbin"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

var (
	globalEnforcer *dobytecasbin.Enforcer // Alias for *casbin.Enforcer
	onceEnforcer   sync.Once
)

func Casbin() *dobytecasbin.Enforcer {
	onceEnforcer.Do(func() {
		ctx := gctx.New()

		modelPath := g.Cfg().MustGet(ctx, "casbin.model", "manifest/config/casbin_model.conf").String()
		if !gfile.Exists(modelPath) {
			g.Log().Fatalf(ctx, "Casbin Service: Casbin model file '%s' not found.", modelPath)
			return
		}

		casbinTable := g.Cfg().MustGet(ctx, "casbin.table", "casbin_rules").String()
		dbLink := g.Cfg().MustGet(ctx, "database.default.link").String()
		if dbLink == "" {
			g.Log().Fatal(ctx, "Casbin Service: 'database.default.link' not found or is empty in configuration.")
			return
		}

		options := &dobytecasbin.Options{
			Model:    modelPath,
			Table:    casbinTable,
			Link:     dbLink,
			Enable:   true,
			AutoLoad: true,
			Duration: 5 * time.Second,
			Debug:    g.Cfg().MustGet(ctx, "database.default.debug", false).Bool(),
		}

		var errEnforcer error
		globalEnforcer, errEnforcer = dobytecasbin.NewEnforcer(options)

		if errEnforcer != nil {
			g.Log().Fatalf(ctx, "Casbin Service: Failed to create Casbin enforcer with dobyte/gf-casbin: %v", errEnforcer)
			return
		}

		if err := globalEnforcer.LoadPolicy(); err != nil {
			g.Log().Warningf(ctx, "Casbin Service: Error loading policy from database: %v. (Table: %s). This might be okay if table is new.", err, casbinTable)
		} else {
			g.Log().Info(ctx, "Casbin Service: Policies loaded successfully from database.")
		}

		g.Log().Info(ctx, "Casbin Service: Enforcer initialized successfully using dobyte/gf-casbin.")
		g.Log().Infof(ctx, "Casbin Service: Model: %s, Table: %s", modelPath, casbinTable)

		if len(globalEnforcer.GetPolicy()) == 0 && len(globalEnforcer.GetGroupingPolicy()) == 0 {
			g.Log().Info(ctx, "Casbin Service: No policies found. Adding default admin role and permissions...")

			added, err := globalEnforcer.AddPolicy("admin_role", "/*", "*")
			if err != nil {
				g.Log().Errorf(ctx, "Casbin Service: Failed to add default admin policy (admin_role, /*, *): %v", err)
			} else if !added {
				g.Log().Warningf(ctx, "Casbin Service: Default admin policy (admin_role, /*, *) already exists or was not added.")
			} else {
				g.Log().Info(ctx, "Casbin Service: Default admin policy (admin_role, /*, *) added successfully.")
			}

			if errSave := globalEnforcer.SavePolicy(); errSave != nil {
			    g.Log().Errorf(ctx, "Casbin Service: Failed to save default policies to database: %v", errSave)
			} else {
			    g.Log().Info(ctx, "Casbin Service: Default policies attempt and SavePolicy() executed.")
			}
		} else {
			g.Log().Infof(ctx, "Casbin Service: Existing policies found (Policy rules: %d, Grouping rules: %d). Default policies not added.", len(globalEnforcer.GetPolicy()), len(globalEnforcer.GetGroupingPolicy()))
		}
	})
	if globalEnforcer == nil { // Should be caught by Fatalf, but defensive
	    panic("Casbin globalEnforcer is nil after initialization attempt. Check logs.")
	}
	return globalEnforcer
}

func InitCasbin() {
	if Casbin() == nil {
		g.Log().Error(gctx.New(), "Casbin enforcer failed to initialize and is nil (InitCasbin).")
	}
}

// --- Casbin Operation Helper Functions ---

func CheckPermission(sub string, obj string, act string) (bool, error) {
	e := Casbin()
	// e.Enforce returns (bool, error) where error is for issues during enforcement, not for "access denied"
	return e.Enforce(sub, obj, act)
}

func AddPolicy(sub string, obj string, act string) (bool, error) {
	e := Casbin()
	return e.AddPolicy(sub, obj, act)
}

func RemovePolicy(sub string, obj string, act string) (bool, error) {
	e := Casbin()
	return e.RemovePolicy(sub, obj, act)
}

func RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
    e := Casbin()
    return e.RemoveFilteredPolicy(fieldIndex, fieldValues...)
}

func AddRoleForUser(user string, role string, domain ...string) (bool, error) {
	e := Casbin()
	if len(domain) > 0 {
		return e.AddRoleForUserInDomain(user, role, domain[0])
	}
	return e.AddGroupingPolicy(user, role)
}

func RemoveRoleForUser(user string, role string, domain ...string) (bool, error) {
	e := Casbin()
	if len(domain) > 0 {
		// Corrected method name: DeleteRoleForUserInDomain
		return e.DeleteRoleForUserInDomain(user, role, domain[0])
	}
	return e.RemoveGroupingPolicy(user, role)
}

func GetRolesForUser(user string, domain ...string) ([]string, error) {
	e := Casbin()
	if len(domain) > 0 {
		// Compiler says: GetRolesForUserInDomain(user, domain[0]) returns []string
		return e.GetRolesForUserInDomain(user, domain[0]), nil
	}
	// Compiler says: GetRolesForUser(user) returns ([]string, error)
	return e.GetRolesForUser(user)
}

func GetUsersForRole(role string, domain ...string) ([]string, error) {
	e := Casbin()
	if len(domain) > 0 {
		// Compiler says: GetUsersForRoleInDomain(role, domain[0]) returns []string
		return e.GetUsersForRoleInDomain(role, domain[0]), nil
	}
	// Compiler says: GetUsersForRole(role) returns ([]string, error)
	return e.GetUsersForRole(role)
}

func DeleteRole(role string) (bool, error) {
	e := Casbin()
	// This deletes policy rules (p, role, *, *) and grouping rules (g, *, role)
	// It does not delete g rules like (g, user, role) which assign users to this role.
	// For a full role cleanup, one might need to remove user assignments first.
	// e.RemoveFilteredGroupingPolicy(1, role) // Removes g, *, role_to_delete
	return e.DeleteRole(role)
}

func DeleteUser(user string) (bool, error) {
	e := Casbin()
	// This removes p rules (p, user, *, *) and g rules (g, user, *)
	return e.DeleteUser(user)
}

func GetAllSubjects() ([]string, error) {
	e := Casbin()
	return e.GetAllSubjects(), nil // This is correct as per casbin.Enforcer
}

func GetAllNamedSubjects(ptype string) ([]string, error) {
	e := Casbin()
	return e.GetAllNamedSubjects(ptype), nil // This is correct
}

func GetAllRoles() ([]string, error) {
	e := Casbin()
	return e.GetAllRoles(), nil // This is correct
}

func GetAllObjects() ([]string, error) {
	e := Casbin()
	return e.GetAllObjects(), nil // This is correct
}

func GetAllActions() ([]string, error) {
	e := Casbin()
	return e.GetAllActions(), nil // This is correct
}
// Re-evaluating GetRolesForUser and GetUsersForRole based on standard casbin.Enforcer
// They indeed return []string, not ([], error).
// The wrapper functions in this service file have an incorrect signature if they intend to match e.g. e.Enforce
// which returns (bool,error). For getters that don't error, the signature should be ([]string) or add 'nil' for error.
// Let's fix them to return (result, nil) to match the declared function signature.
