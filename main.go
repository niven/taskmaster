package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/niven/taskmaster/config"
)

// Domain is a name for something that has tasks and chores
type Domain struct {
	ID    uint32
	Owner uint32
	Name  string
}

// Task is a chore you do
type Task struct {
	ID          uint32
	DomainID    uint32
	Name        string
	Weekly      bool
	Description string
}

var store = cookie.NewStore([]byte("secret"))

func setupRouting(router *gin.Engine) {

	router.Static("/static", "static")
	router.Static("/favicon.ico", "static/favicon.ico")

	router.GET("/", indexHandler)

	router.GET("/welcome", welcomeHandler)
	router.GET("/auth", authHandler)

	authorized := router.Group("/domains")
	authorized.Use(AuthorizeRequest())
	{
		authorized.GET("/", domainHandler)
	}

}

func main() {

	router := gin.New()

	router.Use(sessions.Sessions("tm", store))
	router.Use(gin.Logger())

	router.LoadHTMLGlob("templates/*.tmpl.html")

	setupRouting(router)

	router.Run(":" + config.EnvironmentVars["PORT"])
}
