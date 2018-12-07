package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// AllUser ...
type AllUser struct {
	Users        []User `json:"users"`
	userNameList []interface{}
}

// User ...
type User struct {
	Name  string `json:"name"`
	Money int    `json:"money"`
	Key   string `json:"key"`
}

type Response struct {
	Message string `json:"message"`
}

var allUser AllUser

func (users *AllUser) addUser(u User) {
	users.Users = append(users.Users, u)
}

func (users *AllUser) removeUser(name string) bool {

	idx, find := users.findUserIdx(name)

	if !find {
		return false
	}

	allUser.Users = append(allUser.Users[:idx], allUser.Users[idx+1:]...)

	return true
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

	return http.StatusOK, nil
}

func (users *AllUser) findUser(name string) (*User, bool) {

	// for _, user := range users.Users {
	// 	if user.Name == name {
	// 		return user, true
	// 	}
	// }

	for i := 0; i < len(users.Users); i++ {
		if users.Users[i].Name == name {
			return &(users.Users[i]), true
		}
	}

	return &User{}, false

}

func (users *AllUser) findUserIdx(name string) (int, bool) {

	for idx, user := range users.Users {
		if user.Name == name {
			return idx, true
		}
	}

	return -1, false

}

func readAllUserData() {

	allUser = AllUser{}

	byteValue, _ := ioutil.ReadFile("user/users.json")

	json.Unmarshal(byteValue, &(allUser.userNameList))

	for _, userName := range allUser.userNameList {

		fileName := "user/" + fmt.Sprint(userName) + ".json"
		userValue, _ := ioutil.ReadFile(fileName)

		var user User
		json.Unmarshal(userValue, &user)
		allUser.addUser(user)
	}

}

//POST /api/user
func getAccessToken(c echo.Context) error {
	name := c.FormValue("name")
	user, find := allUser.findUser(name)

	if find == false {
		return c.JSON(http.StatusOK, Response{Message: "User not found"})
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
func deposit(c echo.Context) error {
	name := c.QueryParam("name")
	deposit, _ := strconv.Atoi(c.FormValue("money"))

	if verified := verifydUser(c, name); !verified {
		return c.JSON(http.StatusUnauthorized, Response{Message: "token not allowed."})
	}

	statusCode, err := updateUser(name, deposit)

	if err != nil {
		return c.JSON(statusCode, Response{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, Response{Message: "save success"})

}

//GET /api/check?name="yourname"
func checkBalance(c echo.Context) error {
	name := c.QueryParam("name")

	if verified := verifydUser(c, name); !verified {
		return c.JSON(http.StatusUnauthorized, Response{Message: "token not allowed."})
	}

	user, find := allUser.findUser(name)

	if find == false {
		return c.String(http.StatusOK, "User not found.")
	}

	return c.String(http.StatusOK, fmt.Sprint(user.Money))
}

// POST/api/withdraw?name="yourname"
func withdraw(c echo.Context) error {
	name := c.QueryParam("name")
	withdraw, _ := strconv.Atoi(c.FormValue("money"))
	withdraw *= -1

	if verified := verifydUser(c, name); !verified {
		return c.JSON(http.StatusUnauthorized, Response{Message: "token not allowed."})
	}

	statusCode, err := updateUser(name, withdraw)

	if err != nil {
		return c.JSON(statusCode, Response{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, Response{Message: "withdraw success"})

}

// DELETE /api/user?name="yourname"
func deleteUser(c echo.Context) error {
	name := c.QueryParam("name")

	if verified := verifydUser(c, name); !verified {
		return c.JSON(http.StatusUnauthorized, Response{Message: "token not allowed."})
	}

	if allUser.removeUser(name) {
		return c.JSON(http.StatusOK, Response{Message: "remove success"})
	}
	return c.JSON(http.StatusNotFound, Response{Message: "user not found"})

}

func verifydUser(c echo.Context, name string) bool {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	tokenName := claims["name"].(string)

	if tokenName != name {
		return false
	}

	return true
}

func main() {
	e := echo.New()

	readAllUserData()

	e.POST("/api/user", getAccessToken)

	tokenCheck := middleware.JWT([]byte("secret"))

	e.POST("/api/deposit", deposit, tokenCheck)
	e.POST("/api/withdraw", withdraw, tokenCheck)
	e.GET("/api/check", checkBalance, tokenCheck)
	e.DELETE("/api/user", deleteUser, tokenCheck)

	e.Logger.Fatal(e.Start(":1323"))
}
