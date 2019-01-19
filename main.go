package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func isAuthorized(c *gin.Context) bool {
	session := sessions.Default(c)
	v := session.Get("user-id")
	return v != nil
}

// AuthorizeRequest is used to authorize a request for a certain end-point group.
func AuthorizeRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAuthorized(c) {
			c.Next()
		} else {
			c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{"message": "Please login."})
			c.Abort()
		}
	}
}

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

var conf *oauth2.Config
var state string
var store = cookie.NewStore([]byte("secret"))

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

func init() {

	clientSecret := os.Getenv("TASKMASTER_OAUTH_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("$TASKMASTER_OAUTH_CLIENT_SECRET must be set")
	}

	conf = &oauth2.Config{
		ClientID:     "406866902910-omkqfc94h59m45a3120j6k6duic3masd.apps.googleusercontent.com",
		ClientSecret: clientSecret,
		RedirectURL:  "http://taskmaster.org:5000/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

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
