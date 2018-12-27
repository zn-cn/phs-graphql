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
	"github.com/labstack/echo"
)

type Error struct {
	ErrMsg  string `json:"err_msg"`
	ErrCode int    `json:"err_code"`
}

func getJWTUserID(p graphql.ResolveParams) string {
	return p.Context.Value(constant.JWTContextKey).(jwt.MapClaims)["user_id"].(string)
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

/****************************************** 分割线 ****************************************/

// ErrorRes ErrorResponse
type ErrorRes struct {
	Status int    `json:"status"`
	ErrMsg string `json:"err_msg"`
}

// DataRes DataResponse
type DataRes struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

// RetError response error, wrong response
func retError(c echo.Context, code, status int, errMsg string) error {
	return c.JSON(code, ErrorRes{
		Status: status,
		ErrMsg: errMsg,
	})
}

// RetData response data, correct response
func retData(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, DataRes{
		Status: 200,
		Data:   data,
	})
}
