package core

type PermissionCode string

const (
	PermissionSuperUser        PermissionCode = "super_user"
	PermissionCreateUser       PermissionCode = "create_user"
	PermissionViewUser         PermissionCode = "view_user"
	PermissionUpdateUser       PermissionCode = "update_user"
	PermissionCreatePermission PermissionCode = "create_permission"
	PermissionViewPermission   PermissionCode = "view_permission"
	PermissionUpdatePermission PermissionCode = "update_permission"
	PermissionCreateRole       PermissionCode = "create_role"
	PermissionViewRole         PermissionCode = "view_role"
	PermissionUpdateRole       PermissionCode = "update_role"
)
