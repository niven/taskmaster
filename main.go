package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/niven/taskmaster/config"
)

func init() {

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)
}

var store = cookie.NewStore([]byte("secret"))

func setupRouting(router *gin.Engine) {

	router.Static("/static", "static")
	router.Static("/favicon.ico", "static/favicon.ico")

	router.GET("/", indexHandler)

	router.GET("/welcome", welcomeHandler)
	router.GET("/auth", authHandler)

	authorized := router.Group("/")
	authorized.Use(AuthorizeRequest())
	{
		authorized.GET("/today", overviewHandler)
		authorized.GET("/setup", setupHandler)
	}

	domain := router.Group("/domain")
	domain.Use(AuthorizeRequest())
	{
		domain.POST("/new", domainNewHandler)
		domain.GET("/edit/:domain_id", domainEditHandler)
	}

	task := router.Group("/task")
	task.Use(AuthorizeRequest())
	{
		task.POST("/new", taskNewHandler)
		task.POST("/done", taskDoneHandler)
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
