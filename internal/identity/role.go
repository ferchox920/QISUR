package identity

// RoleName enumera roles soportados en el sistema.
type RoleName string

const (
	RoleAdmin   RoleName = "admin"
	RoleUser    RoleName = "user"
	RoleClient  RoleName = "client"
	RoleUnknown RoleName = ""
)
