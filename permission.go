package aclsrv

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Permission is user level. It represents a group of different permissions/activities/actions
// a user can execute on the platform
type Permission uint32

// permission flags
const (
	/*01*/PFlagSeeUsers Permission = 0x1 << iota
	/*02*/PFlagUserSelf
	/*03*/PFlagUsersAll
	/*04*/PFlagSrvLogsSelf
	/*05*/PFlagSrvLogsAll
	/*06*/PFlagPlatformLogs
	/*07*/PFlagSeeJolieAll
	/*08*/PFlagSeeUserSafeSrv
	/*09*/PFlagDeployJolie
	/*10*/PFlagManageJolieSelf
	/*11*/PFlagManageJolieAll
	/*12*/PFlagSeeSrvAll
	/*13*/PFlagCreateSrv
	/*14*/PFlagManageSrvSelf
	/*15*/PFlagManageSrvAll
	/*16*/PFlagManageGCloud
	/*17*/PFlagSeeClusterInfo
	/*18*/PFlagSeePlatformDocs
	/*19*/PFlagManagePlatformDocs
	/*20*/PFlagMoveSrv //including jolie scripts

)

// basic roles
//

// PermissionLvlUnverified self registered, but not approved by an admin
const PermissionLvlUnverified = 0
// PermissionLvlGuest authenticated, but not accepted as a user
const PermissionLvlGuest = PFlagSeeJolieAll |
	PFlagSeeUserSafeSrv

// PermissionLvlUsr user of our platform
const PermissionLvlUsr = PermissionLvlGuest |
	PFlagUserSelf |
	PFlagSrvLogsSelf |
	PFlagDeployJolie |
	PFlagManageJolieSelf |
	PFlagSeePlatformDocs

// PermissionLvlDev platform developer
const PermissionLvlDev = PermissionLvlUsr |
	PFlagManageJolieAll |
	PFlagManagePlatformDocs |
	PFlagManageSrvSelf |
	PFlagCreateSrv |
	PFlagSeeSrvAll |
	PFlagSeeUsers

// PermissionLvlAdm platform administrator / system administrator (sysadm)
const PermissionLvlAdm = PermissionLvlDev |
	PFlagUsersAll |
	PFlagSrvLogsAll |
	PFlagPlatformLogs |
	PFlagManageSrvAll |
	PFlagManageGCloud |
	PFlagSeeClusterInfo |
	PFlagMoveSrv |
	Permission(^uint32(0))



type JWT struct {
	Header string
	Body string
	Signature string
}

// getServiceName extract the service name from the uri path, after the /api prefix
func getNextJWTSection(jwtp string) (section string, err error) {
	if len(jwtp) == 0 {
		err = errors.New("empty")
		return
	}

	index := strings.IndexByte(jwtp, '.')

	section = jwtp // fallback
	if index >= 0 {
		section= jwtp[:index]
	}

	if len(section) == 0 {
		err = errors.New("missing service name")
	}
	return
}

func extractJWT(jwtStr string) (jwt *JWT, err error) {
	strs := strings.Split(jwtStr, ".")
	if len(strs) != 3 {
		err = errors.New("jwt format is incorrect. Unable to extract 3 sections")
		return
	}

	for i := range strs {
		var dec []byte
		dec, err = base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(strs[i])
		if err != nil {
			return
		}
		strs[i] = string(dec)
	}

	return &JWT{
		Header: strs[0],
		Body: strs[1],
		Signature: strs[2],
	}, nil
}

func getJWTUser(r *http.Request) (user *User, length int, err error) {
	jwt := r.Header.Get("jwt")
	jwt2 := r.Header.Get("Authorization")

	if jwt == "" && jwt2 != "" {
		jwt = jwt2[len("Bearer "):]
	}
	length = len(jwt)

	// load sections header, body, signature
	tmp := jwt
	var sections []string
	for {
		var section string
		section, err = getNextJWTSection(tmp)
		if err != nil {
			return
		}

		var dec []byte
		dec, err = base64.StdEncoding.DecodeString(section)
		if err != nil {
			return
		}
		sections = append(sections, string(dec))

		if section == tmp {
			break
		}

		tmp = tmp[len(section) + 1:]
	}

	if len(sections) != 3 {
		err = errors.New("invalid JWT token")
		return
	}

	// TODO: verify JWT signature

	// get user
	err = json.Unmarshal([]byte(sections[1]), &user)
	return
}