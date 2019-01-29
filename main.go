package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/niven/taskmaster/config"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)

	message.SetString(language.AmericanEnglish, "%s went to %s.", "%s is in %s.")
}

var store = cookie.NewStore([]byte("secret"))

func setupRouting(router *gin.Engine) {

	p := message.NewPrinter(language.BritishEnglish)
	p.Printf("There are %v flowers in our garden.\n", 1500)

	p = message.NewPrinter(language.AmericanEnglish)
	p.Printf("%s went to %s.", "Peter", "England")

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
