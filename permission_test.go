package aclsrv

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"testing"
)

func TestJWTParser(t *testing.T) {
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

func TestJWTVerification(t *testing.T) {
	keysData := []byte(`{"keys":[{"alg":"RS256","e":"AQAB","kid":"Ws8tZ0zxCaLDRE+rsOKTG60BqYsYL8AOx/9ZdlnHcuM=","kty":"RSA","n":"hKl65YhybDOfqYyVMcxYQYW7o-UPhl73JDFkxWRQRtgB_Ic-DvprzyM4XJU2gIcOKA_4mN_JofQBHr9u20CLTTPHparWeGd4LfpWv1oeXgWqlcDtUxovVmdvxV3FT43rTVAQ6sEB8kugLMIv_6qxXQ6gKCBdRiTZwM389Q2x-wSMJGR49yGWAN9QlT9gchuH1Tox5BnFz7PvqhUY3mx09g9j7wu_isfFku0tlHvftWhP_FRNczgLV1DR1ejlBiWv_ciGL2E7SwbxE8j-Hi1cWHrINq7imfT97B5dBtt4TLasfkltNgskWxVsgrLOrW8dd3RvctHi2MCr16wCJOQZOw","use":"sig"},{"alg":"RS256","e":"AQAB","kid":"Xlc6uCpdIH3W01dJSAPIhi7FctEL652E7LJi2osU/X4=","kty":"RSA","n":"zYgIa4-f38zOmJw1k4BTaD8gyEImj2zuJd2z8XM4gVFPZAACVy9d16ca_odsq_DGvZNWO11diI-SvWigmw1XiGnNsU2IbFYyYN9JrUiElcG5Xe67GEM-juVqEqyNN5FusrgEphzMdeyw1_fFdKqTQDZcDwLNqCpbGbMkbpMRV1pWCQWoOkHknlfhqyP5Mhbbf6ESwmlWe8hQD7TfMAZUVzOeANCWP4sgGG3l3N_I1wgOEi6AxJtEKl42JdHtFVAQeZ9vbXDKLDs8X63_ZWYGTjBue_FLkmcY9ZgaE0_J82ovaI2J26rIU8ukzF2HEP753UumVBdmGe9w_N1tzToBJw","use":"sig"}]}`)

	set, err := jwk.Parse(keysData)
	if err != nil {
		t.Error(err)
	}

	tokenString := "eyJraWQiOiJXczh0WjB6eENhTERSRStyc09LVEc2MEJxWXNZTDhBT3hcLzlaZGxuSGN1TT0iLCJhbGciOiJSUzI1NiJ9.eyJhdF9oYXNoIjoiV0VlRlB0U1J0MU95TC1ZMzlDZVJVdyIsInN1YiI6ImY3NzliNWFiLTQ0ZGYtNGIyOS05OGM5LWFjNjkwZjIxZjQwNSIsImF1ZCI6IjM1anVvdmdoY3I4bmMzbTVuYmo4NmljMzBnIiwiY29nbml0bzpncm91cHMiOlsidXNlciJdLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwidG9rZW5fdXNlIjoiaWQiLCJhdXRoX3RpbWUiOjE1NDU1ODcxNTUsImlzcyI6Imh0dHBzOlwvXC9jb2duaXRvLWlkcC51cy1lYXN0LTEuYW1hem9uYXdzLmNvbVwvdXMtZWFzdC0xX0FNZm9wbVA2ZSIsImNvZ25pdG86dXNlcm5hbWUiOiJhbmRlcnNmeWxsaW5nIiwiZXhwIjoxNTQ1NTkwNzU1LCJpYXQiOjE1NDU1ODcxNTUsImVtYWlsIjoiYW5meWwxOEBzdHVkZW50LnNkdS5kayJ9.Zn7IXChvzF79XGKhG0tyHZLb72lEoYweq9YhTg6pqfbwzj1Qak0Iy_6ThhEhvVR-0zEij0ZzDIlA5ZYTA5D84Hc4exRcrNXa0fUgLrY-QUJNK-jKsKZ1-NU25EVOLesJG8MnaxhenmgR4DVFJ5xU_rwTyxP5MiqomQ101A_qgkmwPVA-Gi_Pdqb0NM1WdWEFiomLyCcDcU3kAtuP5WEDlgt7yNQy4uzBcy-uALaHd1yio723yc06rf7PM0iksTy__nNchf_TUMh1yuct2JmK_F2iZaks5hVGmcWn3yW-xZ0FSZhErLsebDNWCv_rVn1LRccXb3Xsy6qWAAmWvckNrg"

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("unable to convert kid to string")
		}

		if key := set.LookupKeyID(kid); len(key) == 1 {
			return key[0].Materialize()
		}

		return nil, errors.New("not found")
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["cognito:username"] != "andersfylling" {
			t.Errorf("incorrect username. Got %s, wants %s", claims["cognito:username"], "andersfylling")
		}
	} else if err.Error() != "Token is expired" {
		t.Error(err)
	}

}
