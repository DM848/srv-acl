package aclsrv

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

const jwksURL = "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_AMfopmP6e/.well-known/jwks.json"

// Permission is user level. It represents a group of different permissions/activities/actions
// a user can execute on the platform
type Permission uint32
func (p Permission) Str() string {
	return strconv.FormatInt(int64(p), 10)
}

// permission flags
const (
	/*01*/ PFlagSeeUsers Permission = 0x1 << iota
	/*02*/ PFlagUserSelf // manage only own user data
	/*03*/ PFlagUsersAll // manage all user data
	/*04*/ PFlagSrvLogsSelf // see only logs regarding their own (service, scripts, platform)
	/*05*/ PFlagSrvLogsAll // See logs for every service and script and platform
	/*06*/ PFlagPlatformLogs // see all platform logs
	/*07*/ PFlagSeeJolieAll // see all jolie services
	/*08*/ PFlagSeeUserSafeSrv // see safe services (created for user scripts)
	/*09*/ PFlagDeployJolie // deploy a jolie script
	/*10*/ PFlagManageJolieSelf // manage own deployed jolie script (edit, undeploy)
	/*11*/ PFlagManageJolieAll // manage all deployed jolie script (edit, undeploy)
	/*12*/ PFlagSeeSrvAll // see every service (including platform services)
	/*13*/ PFlagCreateSrv // create a new service through the generator
	/*14*/ PFlagManageSrvSelf // manage own created service (edit, undeploy)
	/*15*/ PFlagManageSrvAll // manage all created services
	/*16*/ PFlagManageGCloud // send commands to google cloud (scaling perhaps)
	/*17*/ PFlagSeeClusterInfo // see google cloud/k8s info about our cluster
	/*18*/ PFlagSeePlatformDocs // see platform docs
	/*19*/ PFlagManagePlatformDocs // manage platform docs
	/*20*/ PFlagMoveSrv // move a service from one node to another

	// To add move permission flags, create a PR or a GitHub issue.
)

// basic roles
//

// PermissionLvlNobody self registered, but not approved by an admin
const PermissionLvlNobody = 0

// PermissionLvlUsr user of our platform
const PermissionLvlUsr = PFlagSeeJolieAll |
	PFlagSeeUserSafeSrv |
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
	Header    string
	Body      string
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
		section = jwtp[:index]
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
		Header:    strs[0],
		Body:      strs[1],
		Signature: strs[2],
	}, nil
}

func getJWT(header http.Header) string {
	jwt := header.Get("jwt")
	jwt2 := header.Get("Authorization")

	if jwt == "" && jwt2 != "" {
		jwt = jwt2[len("Bearer "):]
	}

	if jwt == "" {
		jwt = header.Get("JWT")
	}

	return jwt
}
