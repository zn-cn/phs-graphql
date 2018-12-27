package controller

import (
	"config"
	"constant"
	"io/ioutil"
	"net/http"
	"util/token"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/graphql-go/graphql"
	jsoniter "github.com/json-iterator/go"
)

type Error struct {
	ErrMsg  string `json:"err_msg"`
	ErrCode int    `json:"err_code"`
}

func getJWTUserID(p graphql.ResolveParams) string {
	return p.Context.Value(constant.JWTContextKey).(jwt.MapClaims)["userID"].(string)
}

func getJWTToken(auth map[string]interface{}) string {
	return token.GetJWTToken(auth, config.Conf.Security.Secret, constant.JWTExpire)
}

func validateJWT(tokenStr string) (jwt.MapClaims, bool) {
	return token.ValidateJWT(constant.JWTAuthScheme, tokenStr, config.Conf.Security.Secret)
}

func loadJSONData(r *http.Request, to interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return jsoniter.Unmarshal(body, to)
}

func resJSONData(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var resData []byte
	if data != nil {
		resData, _ = jsoniter.Marshal(data)
	}
	w.Write(resData)
}

func resJSONError(w http.ResponseWriter, code int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data := Error{
		ErrMsg: errMsg,
	}
	resData, _ := jsoniter.Marshal(data)
	w.Write(resData)
}
