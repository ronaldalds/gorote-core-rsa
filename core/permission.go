package core

type PermissionCode string

const (
	PermissionSuperUser            PermissionCode = "super_user"
	PermissionCreateUser           PermissionCode = "create_user"
	PermissionViewUser             PermissionCode = "view_user"
	PermissionUpdateUser           PermissionCode = "update_user"
	PermissionEditePermissionsUser PermissionCode = "edite_permissions_user"
	PermissionCreateRole           PermissionCode = "create_role"
	PermissionViewRole             PermissionCode = "view_role"
	PermissionUpdateRole           PermissionCode = "update_role"
)
