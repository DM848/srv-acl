package aclsrv



type ACLEntry struct {
	Service string `json:"service"`
	MinimumPermission Permission `json:"min_permission"`
	AllowedUserIDs []UserID
	BlockedUserIDs []UserID
	LastUpdated int64 // unix
}

func (e *ACLEntry) Empty() bool {
	return e.Service == ""
}

func (e *ACLEntry) HasAccess(user *User) bool {
	// check if explicitly blocked
	for _, blocked := range e.BlockedUserIDs {
		if user.ID == blocked {
			return false
		}
	}

	// check if explicitly whitelisted
	for _, allowed := range e.AllowedUserIDs {
		if user.ID == allowed {
			return true
		}
	}

	return (user.Permission & e.MinimumPermission) == e.MinimumPermission
}