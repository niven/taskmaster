package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/niven/taskmaster/config"
	. "github.com/niven/taskmaster/handlers"
)

func init() {

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)
}

var store = cookie.NewStore([]byte("secret"))

func setupRouting(router *gin.Engine) {

	router.Static("/static", "./static")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")

	router.GET("/", IndexHandler)

	router.GET("/welcome", WelcomeHandler)
	router.GET("/auth", AuthHandler)

	authorized := router.Group("/")
	authorized.Use(AuthorizeRequest())
	{
		authorized.GET("/today", OverviewHandler)
		authorized.GET("/setup", SetupHandler)
	}

	domain := router.Group("/domain")
	domain.Use(AuthorizeRequest())
	{
		domain.POST("/new", DomainNewHandler)
		domain.GET("/edit/:domain_id", DomainEditHandler)
	}

	task := router.Group("/task")
	task.Use(AuthorizeRequest())
	{
		task.POST("/new", TaskNewHandler)
		task.POST("/done", TaskDoneHandler)
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
