package domain

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleSeller UserRole = "seller"
	RoleBuyer  UserRole = "buyer"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleSeller, RoleBuyer:
		return true
	}
	return false
}