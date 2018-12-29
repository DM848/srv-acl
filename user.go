package aclsrv

type UserID string // some type of token
func (uid UserID) Str() string {
	return string(uid)
}

type User struct {
	ID         UserID
	Permission Permission
}
