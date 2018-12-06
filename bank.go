package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

type AllUser struct {
	Users        []User `json:"users"`
	userNameList []interface{}
}

type User struct {
	Name  string `json:"name"`
	Money int    `json:"money"`
	Key   string `json:"key"`
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

func updateUser(name string, money int) bool {
	user, find := allUser.findUser(name)

	if !find {
		return false
	}

	user.Money += money

	return true
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
		return c.String(http.StatusOK, "User not found.")
	}

	return c.String(http.StatusOK, user.Key)

}

//POST /api/deposit?name="yourname"
func deposit(c echo.Context) error {
	name := c.QueryParam("name")
	deposit, _ := strconv.Atoi(c.FormValue("money"))

	updateUser(name, deposit)

	return c.String(http.StatusOK, "ok")
}

//GET /api/check?name="yourname"
func checkBalance(c echo.Context) error {
	name := c.QueryParam("name")
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

	updateUser(name, withdraw)

	return c.String(http.StatusOK, "ok")
}

// DELETE /api/user?name="yourname"
func deleteUser(c echo.Context) error {
	name := c.QueryParam("name")
	allUser.removeUser(name)
	return c.String(http.StatusOK, "ok")
}

func main() {
	e := echo.New()

	readAllUserData()

	e.POST("/api/user", getAccessToken)
	e.POST("/api/deposit", deposit)
	e.POST("/api/withdraw", withdraw)
	e.GET("/api/check", checkBalance)
	e.DELETE("/api/user", deleteUser)

	e.Logger.Fatal(e.Start(":1323"))
}
