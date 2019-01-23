package main

import (
	"fmt"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/niven/taskmaster/config"
)

var store = cookie.NewStore([]byte("secret"))

func setupRouting(router *gin.Engine) {

	router.Static("/static", "static")
	router.Static("/favicon.ico", "static/favicon.ico")

	router.GET("/welcome", welcomeHandler)
	router.GET("/auth", authHandler)

	authorized := router.Group("/")
	authorized.Use(AuthorizeRequest())
	{
		authorized.GET("/", indexHandler)
		authorized.GET("/setup", setupHandler)
	}

}

func main() {

	err := config.ReadEnvironmentVars()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	router := gin.New()

	router.Use(sessions.Sessions("tm", store))
	router.Use(gin.Logger())

	router.LoadHTMLGlob("templates/*.tmpl.html")

	setupRouting(router)

	router.Run(":" + config.EnvironmentVars["PORT"])
}
