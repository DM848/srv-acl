package aclsrv

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJWTParser(t *testing.T) {
	//jwt1 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiIxMjM0NTY3ODkwIiwicGVybWlzc2lvbiI6MX0.-F0nDX9liSaLqSo63gjA6q0ph46SEG98dl1lvGRIt-U"
	//jwt1_head := `{"alg":"HS256","typ":"JWT"}`
	//jwt1_body := `{"userID":"1234567890","permission":1}`
	//jwt1_sign := `fa64g67wef67afg67eae76ufsea7ufau4gf6uwg4f_a4ffg467w`
	//
	//jwt, err := extractJWT(jwt1)
	//if err != nil {
	//	t.Error(err)
	//}
	//if jwt.Header != jwt1_head {
	//	t.Errorf("header missmatch. Got %s, wants %s", jwt.Header, jwt1_head)
	//}
	//if jwt.Body != jwt1_body {
	//	t.Errorf("body missmatch. Got %s, wants %s", jwt.Body, jwt1_body)
	//}
	//if jwt.Signature != jwt1_sign {
	//	t.Errorf("signature missmatch. Got %s, wants %s", jwt.Signature, jwt1_sign)
	//}

	a := []byte("{\"status\":\"ok\"}")
	b := struct {
		S string `json:"status"`
	}{}
	err := json.Unmarshal(a, &b)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", b)
}

func TestPermission(t *testing.T) {
	fmt.Println(PermissionLvlUsr)
	fmt.Println(PermissionLvlDev)
	fmt.Println(PermissionLvlAdm)
}