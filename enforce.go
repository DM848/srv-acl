package aclsrv

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func enforceURLQueryParams(values *url.Values, user *User) {
	if values.Get("acle_user_id") != "" {
		values.Set("acle_user_id", user.ID.Str())
	}
	if values.Get("acle_user_level") != "" {
		values.Set("acle_user_level", user.Permission.Str())
	}
	if values.Get("acle_user_level_str") != "" {
		var role string = "nobody"
		if (PermissionLvlAdm & user.Permission) == PermissionLvlAdm {
			role = "adm"
		} else if (PermissionLvlDev & user.Permission) == PermissionLvlDev {
			role = "dev"
		} else if (PermissionLvlUsr & user.Permission) == PermissionLvlUsr {
			role = "usr"
		}
		values.Set("acle_user_level_str", role)
	}
}

type RawJSONStruct = map[string]json.RawMessage

func enforceJSONBodyParams(r *http.Request, user *User) io.ReadCloser {
	return nil
}
