// File: internal/domain/profile/role.go
package profile

// Role định danh vai trò của một Profile, đồng bộ với enum Role bên auth-service.
type Role string

const (
	RoleAdmin   Role = "ADMIN"
	RoleExpert  Role = "EXPERT"
	RolePatient Role = "PATIENT"
)

func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}
