package aclsrv

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
)

const (
	ACLE_uid = "acle_user_id"
	ACLE_ulvl = "acle_user_level"
	ACLE_ulvlStr = "acle_user_level_str"
)

func getRoleName(p Permission) (role string) {
	role = "nobody"
	if (PermissionLvlAdm & p) == PermissionLvlAdm {
		role = "adm"
	} else if (PermissionLvlDev & p) == PermissionLvlDev {
		role = "dev"
	} else if (PermissionLvlUsr & p) == PermissionLvlUsr {
		role = "usr"
	}

	return role
}

func enforceURLQueryParams(values *url.Values, user *User) {
	if values.Get(ACLE_uid) != "" {
		values.Set(ACLE_uid, user.ID.Str())
	}
	if values.Get(ACLE_ulvl) != "" {
		values.Set(ACLE_ulvl, user.Permission.Str())
	}
	if values.Get(ACLE_ulvlStr) != "" {
		values.Set(ACLE_ulvlStr, getRoleName(user.Permission))
	}
}

func enforceJSONBodyParams(r io.Reader, user *User) (rc io.ReadCloser, length int64, err error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, 0, err
	}

	var parsed map[string]json.RawMessage
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, 0, err
	}

	var changed bool
	if _, exists := parsed[ACLE_uid]; exists && string(parsed[ACLE_uid]) != user.ID.Str() {
		parsed[ACLE_uid] = json.RawMessage(`"` + user.ID.Str() + `"`)
		changed = true
	}
	if _, exists := parsed[ACLE_ulvl]; exists && string(parsed[ACLE_ulvl]) != user.ID.Str() {
		parsed[ACLE_ulvl] = json.RawMessage(user.Permission.Str())
		changed = true
	}
	if _, exists := parsed[ACLE_ulvlStr]; exists && string(parsed[ACLE_ulvlStr]) != user.ID.Str() {
		parsed[ACLE_ulvlStr] = json.RawMessage(`"` + getRoleName(user.Permission) + `"`)
		changed = true
	}

	if changed {
		body, err = json.Marshal(parsed)
		if err != nil {
			return nil, 0, err
		}
	}

	return ioutil.NopCloser(bytes.NewReader(body)), int64(len(body)),nil
}
