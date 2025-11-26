package identity

// Permission representa capacidades de alto nivel.
type Permission string

const (
	PermManageUsers    Permission = "manage_users"
	PermManageCatalog  Permission = "manage_catalog"
	PermViewCatalog    Permission = "view_catalog"
	PermManageIdentity Permission = "manage_identity"
)
