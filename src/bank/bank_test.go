package bank

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

type responseToken struct {
	Token string `json:"token`
}

func setupTest(t *testing.T) {

	path := "../../user/mock_data/"

	ResetAllUserData(path)
	ReadAllUserData(path)

}

func mockUserVerify(c echo.Context, s string) bool {
	return true
}

func mockRouter() *RouterImpl {
	return &RouterImpl{VerifyUser: mockUserVerify}
}

func ResetAllUserData(folder string) {

	writeByte := []byte(`{"name":"kevin","money":200}`)

	ioutil.WriteFile(folder+"kevin.json", writeByte, 0644)

}

func prepareRequest(name string, request func(q url.Values) *http.Request) (*httptest.ResponseRecorder, echo.Context) {
	e := echo.New()

	q := make(url.Values)
	q.Set("name", name)

	req := request(q)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return rec, c
}

func checkBalanceApi(t *testing.T, name string, expectAmount int) {

	rec, c := prepareRequest("kevin", func(q url.Values) *http.Request {
		return httptest.NewRequest(http.MethodGet, "/api/withdraw?"+q.Encode(), nil)
	})

	router := mockRouter()

	err := router.CheckBalance(c)
	assert.NoError(t, err)

	expected := `{"money":` + fmt.Sprintf("%v", expectAmount) + "}"
	response := rec.Body.String()
	assert.JSONEq(t, response, expected)
}

func TestGetAccessToken(t *testing.T) {

	setupTest(t)

	rec, c := prepareRequest("kevin", func(q url.Values) *http.Request {
		return httptest.NewRequest(http.MethodPost, "/api/user?"+q.Encode(), nil)
	})

	err := GetAccessToken(c)
	assert.NoError(t, err)

	response := rec.Body.String()

	var rs responseToken
	err = json.Unmarshal([]byte(response), &rs)
	if err != nil {
		fmt.Println("error:", err)
	}

	assert.NotEmpty(t, rs.Token)

}

func TestCheckBalance(t *testing.T) {
	setupTest(t)
	checkBalanceApi(t, "kevin", 200)
}

func TestDeposit(t *testing.T) {
	setupTest(t)
	name := "kevin"
	payload := `{"money":10}`

	rec, c := prepareRequest("kevin", func(q url.Values) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/api/deposit?"+q.Encode(), strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return req
	})

	router := &RouterImpl{VerifyUser: mockUserVerify}
	err := router.Deposit(c)
	assert.NoError(t, err)

	expected := `{"message": "save success"}`
	response := rec.Body.String()
	assert.JSONEq(t, response, expected)

	checkBalanceApi(t, name, 210)
}

func TestWithdraw(t *testing.T) {

	setupTest(t)
	name := "kevin"

	payload := `{"money":10}`

	rec, c := prepareRequest(name, func(q url.Values) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/api/withdraw?"+q.Encode(), strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return req
	})

	router := &RouterImpl{VerifyUser: mockUserVerify}
	err := router.Withdraw(c)
	assert.NoError(t, err)

	expected := `{"message": "withdraw success"}`
	response := rec.Body.String()
	assert.JSONEq(t, response, expected)

	checkBalanceApi(t, name, 190)

}

func TestDeleteUser(t *testing.T) {

	setupTest(t)
	name := "kevin"

	rec, c := prepareRequest(name, func(q url.Values) *http.Request {
		return httptest.NewRequest(http.MethodDelete, "/api/user?"+q.Encode(), nil)
	})

	router := &RouterImpl{VerifyUser: mockUserVerify}
	err := router.DeleteUser(c)
	assert.NoError(t, err)

	expected := `{"message": "remove success"}`
	response := rec.Body.String()
	assert.JSONEq(t, response, expected)

}
