package aclsrv

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
	"testing"
)

func TestEnforceURLQueryParams(t *testing.T) {
	usr := &User{
		ID: "andersfylling",
		Permission: PermissionLvlDev,
	}

	var got string
	var want string

	up1 := "http://example.com/test"
	u, _ := url.Parse(up1 + "?random=0")
	vals := u.Query()

	enforceURLQueryParams(&vals, usr)
	got = up1+"?"+vals.Encode()
	want = u.String()
	if got != u.String() {
		t.Errorf("Incorrect url after manipulation. Got %s, Wants %s", got, want)
	}

	u, _ = url.Parse(up1 + "?acle_user_id=incorrect")
	vals = u.Query()

	enforceURLQueryParams(&vals, usr)
	got = up1+"?"+vals.Encode()
	want = up1+"?"+"acle_user_id=andersfylling"
	if got == u.String() || got != want {
		t.Errorf("Incorrect url after manipulation. Got %s, Wants %s", got, want)
	}

	u, _ = url.Parse(up1 + "?random=0&acle_user_level=56574544")
	vals = u.Query()

	enforceURLQueryParams(&vals, usr)
	got = up1+"?"+vals.Encode()
	want = up1+"?acle_user_level=" + PermissionLvlDev.Str() + "&random=0"
	if got == u.String() || got != want {
		t.Errorf("Incorrect url after manipulation. Got %s, Wants %s", got, want)
	}
}

func validateEnforcer(t *testing.T, body map[string]interface{}, usr *User, wants *acleRes) {
	var rc io.ReadCloser
	var rn io.ReadCloser
	var data []byte
	var err error

	data, err = json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	rc = ioutil.NopCloser(bytes.NewReader(data))
	defer rc.Close()
	rn, _, err = enforceJSONBodyParams(rc, usr)
	if err != nil {
		t.Fatal(err)
	}
	if rc != rn {
		defer rn.Close()
	}

	data2, err := ioutil.ReadAll(rn)
	if err != nil {
		t.Fatal(err)
	}

	var got *acleRes
	err = json.Unmarshal(data2, &got)
	if err != nil {
		t.Fatal(err)
	}
	
	if got.ID != wants.ID {
		t.Errorf("Missmatch. Got %s, wants %s", got.ID, wants.ID)
	}
	if got.Lvl != wants.Lvl {
		t.Errorf("Missmatch. Got %d, wants %d", got.Lvl, wants.Lvl)
	}
	if got.LvlStr != wants.LvlStr {
		t.Errorf("Missmatch. Got %s, wants %s", got.LvlStr, wants.LvlStr)
	}
}

type acleRes struct {
	ID UserID `json:"acle_user_id"`
	Lvl Permission `json:"acle_user_level"`
	LvlStr string `json:"acle_user_level_str"`
}

func TestEnforceJSONBodyParams(t *testing.T) {
	usr := &User{
		ID: "andersfylling",
		Permission: PermissionLvlDev,
	}

	r := map[string]interface{}{}
	r["random"] = 0
	validateEnforcer(t, r, usr, &acleRes{})

	r[ACLE_uid] = "anders"
	validateEnforcer(t, r, usr, &acleRes{
		ID: usr.ID,
	})

	r[ACLE_uid] = 7
	validateEnforcer(t, r, usr, &acleRes{
		ID: usr.ID,
	})

	r[ACLE_ulvl] = 7
	validateEnforcer(t, r, usr, &acleRes{
		ID: usr.ID,
		Lvl: usr.Permission,
	})
}