package bank

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// AllUser ...
type AllUser struct {
	Users        map[string]*User
	userNameList []interface{}
}

// User ...
type User struct {
	Name  string `json:"name"`
	Money int    `json:"money"`
}

type Response struct {
	Message string `json:"message"`
}

type PayloadMoney struct {
	Money int `json:"money"`
}

type Router interface {
	Deposit(echo.Context) error
	Withdraw(echo.Context) error
	CheckBalance(echo.Context) error
	DeleteUser(echo.Context) error
}

type RouterImpl struct {
	VerifyUser func(echo.Context, string) bool
}

var allUser AllUser

func (users *AllUser) addUser(u User) {
	users.Users[u.Name] = &u
}

func (users *AllUser) removeUser(name string) bool {

	if _, ok := users.Users[name]; ok {
		delete(users.Users, name)
		return true
	}
	return false
}

func updateUser(name string, money int) (int, error) {
	user, find := allUser.findUser(name)

	if !find {
		return http.StatusNotFound, errors.New("user not found")
	}

	if user.Money+money < 0 {
		return http.StatusOK, errors.New("not enough money")
	}

	user.Money += money
	writeUserData(user)

	return http.StatusOK, nil
}

func (users *AllUser) findUser(name string) (*User, bool) {

	if target, ok := users.Users[name]; ok {
		return target, true
	}

	return &User{}, false

}

func writeUserData(user *User) {

	fileName := "user/" + user.Name + ".json"

	byteValue, _ := ioutil.ReadFile(fileName)

	temp := User{}

	json.Unmarshal(byteValue, &temp)

	temp.Money = user.Money

	byteWrite, _ := json.Marshal(&temp)

	ioutil.WriteFile(fileName, byteWrite, 0644)

}

func ReadAllUserData(folder string) {

	allUser = AllUser{Users: make(map[string]*User)}

	byteValue, _ := ioutil.ReadFile(folder + "users.json")

	json.Unmarshal(byteValue, &(allUser.userNameList))

	for _, userName := range allUser.userNameList {

		fileName := folder + fmt.Sprint(userName) + ".json"
		userValue, _ := ioutil.ReadFile(fileName)

		var user User
		json.Unmarshal(userValue, &user)
		allUser.addUser(user)
	}

}

func formattedMoney(money int) (int, error) {
	// money, err := strconv.Atoi(moneyStr)

	if money < 0 {
		return 0, errors.New("invalid syntax of money")
	}
	return money, nil
}

//POST /api/user
func GetAccessToken(c echo.Context) error {
	name := c.FormValue("name")
	user, find := allUser.findUser(name)

	if find == false {
		return c.JSON(http.StatusNotFound, Response{Message: "user not found"})
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.Name
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})

}

//POST /api/deposit?name="yourname"
func (router *RouterImpl) Deposit(c echo.Context) (err error) {
	name := c.QueryParam("name")
	payload := new(PayloadMoney)
	if err = c.Bind(payload); err != nil {
		return
	}
	deposit, err := formattedMoney(payload.Money)

	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
	}

	if verified := router.VerifyUser(c, name); !verified {
		return c.JSON(http.StatusUnauthorized, Response{Message: "token not allowed."})
	}

	statusCode, err := updateUser(name, deposit)

	if err != nil {
		return c.JSON(statusCode, Response{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, Response{Message: "save success"})

}

//GET /api/check?name="yourname"
func (router *RouterImpl) CheckBalance(c echo.Context) (err error) {
	name := c.QueryParam("name")

	if verified := router.VerifyUser(c, name); !verified {
		return c.JSON(http.StatusUnauthorized, Response{Message: "token not allowed."})
	}

	user, find := allUser.findUser(name)

	if find == false {
		return c.JSON(http.StatusNotFound, Response{Message: "user not found"})
	}

	return c.JSON(http.StatusOK, PayloadMoney{Money: user.Money})
}

// POST/api/withdraw?name="yourname"
func (router *RouterImpl) Withdraw(c echo.Context) (err error) {
	name := c.QueryParam("name")
	payload := new(PayloadMoney)
	if err = c.Bind(payload); err != nil {
		return
	}

	withdraw, err := formattedMoney(payload.Money)
	withdraw *= -1

	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
	}

	if verified := router.VerifyUser(c, name); !verified {
		return c.JSON(http.StatusUnauthorized, Response{Message: "token not allowed."})
	}

	statusCode, err := updateUser(name, withdraw)

	if err != nil {
		return c.JSON(statusCode, Response{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, Response{Message: "withdraw success"})

}

// DELETE /api/user?name="yourname"
func (router *RouterImpl) DeleteUser(c echo.Context) (err error) {
	name := c.QueryParam("name")

	if verified := router.VerifyUser(c, name); !verified {
		return c.JSON(http.StatusUnauthorized, Response{Message: "token not allowed."})
	}

	if allUser.removeUser(name) {
		return c.JSON(http.StatusOK, Response{Message: "remove success"})
	}
	return c.JSON(http.StatusNotFound, Response{Message: "user not found"})

}

func VerifyUser(c echo.Context, name string) bool {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	tokenName := claims["name"].(string)

	if tokenName != name {
		return false
	}

	return true
}
