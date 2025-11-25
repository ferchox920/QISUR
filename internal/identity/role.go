package identity

// RoleName enumerates supported roles in the system.
type RoleName string

const (
	RoleAdmin   RoleName = "admin"
	RoleUser    RoleName = "user"
	RoleClient  RoleName = "client"
	RoleUnknown RoleName = ""
)
