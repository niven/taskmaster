package main

import (
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// Minion is someone who performs tasks in Domains
type Minion struct {
	ID   uint32
	Name string
}

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

// User is a retrieved and authenticated user
type User struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
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

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()

	router.Use(sessions.Sessions("tm", store))
	router.Use(gin.Logger())

	router.LoadHTMLGlob("templates/*.tmpl.html")

	setupRouting(router)

	router.Run(":" + port)
}
