package main

import (
	"bank"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()

	bank.ReadAllUserData()

	e.POST("/api/user", bank.GetAccessToken)

	tokenCheck := middleware.JWT([]byte("secret"))
	var router bank.Router
	router = &bank.RouterImpl{}

	e.POST("/api/deposit", router.Deposit, tokenCheck)
	e.POST("/api/withdraw", router.Withdraw, tokenCheck)
	e.GET("/api/check", router.CheckBalance, tokenCheck)
	e.DELETE("/api/user", router.DeleteUser, tokenCheck)

	e.Logger.Fatal(e.Start(":1323"))
}
