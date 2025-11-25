package identity

// Permission represents coarse-grained capabilities.
type Permission string

const (
	PermManageUsers    Permission = "manage_users"
	PermManageCatalog  Permission = "manage_catalog"
	PermViewCatalog    Permission = "view_catalog"
	PermManageIdentity Permission = "manage_identity"
)
