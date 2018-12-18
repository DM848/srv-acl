package aclsrv

type UserID string // some type of token

type User struct {
	ID UserID
	Permission Permission
}