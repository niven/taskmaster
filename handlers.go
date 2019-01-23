package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/niven/taskmaster/config"
	. "github.com/niven/taskmaster/data"
	"github.com/niven/taskmaster/db"
)

var conf *oauth2.Config
var state string

func init() {

	conf = &oauth2.Config{
		ClientID:     "406866902910-omkqfc94h59m45a3120j6k6duic3masd.apps.googleusercontent.com",
		ClientSecret: config.EnvironmentVars["TASKMASTER_OAUTH_CLIENT_SECRET"],
		RedirectURL:  "http://taskmaster.org:5000/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

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

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

// User is a retrieved and authenticated user
type GoogleUser struct {
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

	var user GoogleUser

	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		fmt.Println("error:", err)
	}

	session.Set("user-id", user.Email)
	session.Set("user-name", user.Name)
	err = session.Save()
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{"message": "Error while saving session. Please try again."})
		return
	}

	c.HTML(http.StatusOK, "index.tmpl.html", nil)

}

func indexHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)

	if !found {
		log.Printf("User doesn't exist, creating")
		userName := session.Get("user-name")
		if userName == nil {
			userName = "No Name"
		}
		err := db.CreateMinion(userEmail, userName.(string))
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		db.LoadMinion(userEmail, &minion)
	}

	domains, err := db.GetDomainsForMinion(minion)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	err = Update(minion)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	// get all tasks for each domain: everything pending (for today/this week) & today's task
	pendingTasks, err := db.GetPendingTasksForMinion(minion)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.HTML(http.StatusOK, "index.tmpl.html", gin.H{
		"minion":  minion,
		"domains": domains,
		"pending": pendingTasks,
	})

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

func setupHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		c.AbortWithError(http.StatusBadRequest, errors.New("User authenticated but not found"))
	}

	domains, err := db.GetDomainsForMinion(minion)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.HTML(http.StatusOK, "setup.tmpl.html", gin.H{
		"minion":  minion,
		"domains": domains,
	})
}
