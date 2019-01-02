package aclsrv

type UserID string // some type of token
func (uid UserID) Str() string {
	return string(uid)
}

type User struct {
	ID         UserID `json:"uid,omitempty"`
	Permission Permission `json:"p"`
}
