package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func authHandler(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %s !+ %s", retrievedState, c.Query("state")))
		return
	}

	tok, err := conf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := conf.Client(oauth2.NoContext, tok)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer userinfo.Body.Close()
	data, _ := ioutil.ReadAll(userinfo.Body)
	log.Println("Email body: ", string(data))

	var user User

	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		fmt.Println("error:", err)
	}

	session.Set("user-id", user.Email)
	err = session.Save()
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{"message": "Error while saving session. Please try again."})
		return
	}

	c.HTML(http.StatusOK, "index.tmpl.html", nil)

}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

func welcomeHandler(c *gin.Context) {

	state = randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()

	c.HTML(http.StatusOK, "welcome.tmpl.html", gin.H{
		"login_url": getLoginURL(state),
	})
}

func domainHandler(c *gin.Context) {

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

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()

	router.Use(sessions.Sessions("tm", store))
	router.Use(gin.Logger())

	router.LoadHTMLGlob("templates/*.tmpl.html")

	router.Static("/static", "static")
	router.Static("/favicon.ico", "static/favicon.ico")

	router.GET("/", func(c *gin.Context) {

		if isAuthorized(c) {
			c.HTML(http.StatusOK, "index.tmpl.html", nil)
		} else {
			c.Request.URL.Path = "/welcome"
			router.HandleContext(c)
		}

	})

	router.GET("/welcome", welcomeHandler)
	router.GET("/auth", authHandler)

	authorized := router.Group("/domains")
	authorized.Use(AuthorizeRequest())
	{
		authorized.GET("/", domainHandler)
	}

	router.Run(":" + port)
}
