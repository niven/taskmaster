package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

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
			welcomeHandler(c)
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

	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		errorHandler(c, fmt.Sprintf("Invalid session state: %s !+ %s", retrievedState, c.Query("state")), nil)
		return
	}

	tok, err := conf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		errorHandler(c, "", err)
		return
	}

	client := conf.Client(oauth2.NoContext, tok)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		errorHandler(c, "", err)
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
		errorHandler(c, "Error while saving session. Please try again.", err)
		return
	}

	c.HTML(http.StatusOK, "index.tmpl.html", nil)

}

func indexHandler(c *gin.Context) {
	if isAuthorized(c) {
		overviewHandler(c)
	} else {
		welcomeHandler(c)
	}
}

func overviewHandler(c *gin.Context) {

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
			errorHandler(c, "", err)
			return
		}
		db.LoadMinion(userEmail, &minion)
	}

	domains, err := db.GetDomainsForMinion(minion)
	if err != nil {
		errorHandler(c, "", err)
		return
	}

	err = Update(minion)
	if err != nil {
		errorHandler(c, "", err)
		return
	}

	// get all tasks for each domain: everything pending (for today/this week) & today's task
	pendingTasks, err := db.GetPendingTasksForMinion(minion)
	if err != nil {
		errorHandler(c, "", err)
		return
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
		errorHandler(c, "User authenticated but not found", nil)
		return
	}

	domains, err := db.GetDomainsForMinion(minion)
	if err != nil {
		errorHandler(c, "User authenticated but not found", nil)
		return
	}

	c.HTML(http.StatusOK, "setup.tmpl.html", gin.H{
		"minion":  minion,
		"domains": domains,
	})
}

func domainNewHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		errorHandler(c, "User authenticated but not found", nil)
		return
	}

	domainName := c.DefaultPostForm("name", "Unnamed Deck")
	db.CreateNewDomain(minion, domainName)

	setupHandler(c)
}

func errorHandler(c *gin.Context, message string, err error) {

	c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
		"message": message,
		"error":   err,
	})

}

func domainEditHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		errorHandler(c, "User authenticated but not found", nil)
	}

	domainID, err := strconv.Atoi(c.Param("domain_id"))
	if err != nil || domainID < 0 {
		errorHandler(c, "Invalid domain ID", err)
		return
	}

	domain, err := db.GetDomainByID(uint32(domainID))
	if err != nil || domain.Owner != minion.ID {
		errorHandler(c, "Domain not found", err)
		return
	}

	tasks, err := db.GetTasksForDomain(domain)
	if err != nil {
		// return not found to avoid leaking domain IDs. Not that it matters here, but general principle
		errorHandler(c, "Domain not found", err)
		return
	}
	/*
		Create a map of tasks and counts by name, so they can be rendered grouped "Laundry x4" is better than 4 line items
		This is a bit annoying maybe, since this is a virtual deck of cards it's nice to have each card be a record/object
		for shuffling and assigning purposes, but since there can be duplicates it's not nice to see a big list of similar
		stuff.

		For editing the deck we group them up, but then saving is an issue when there are ones that are already assigned or completed.
		One option is to just reset everything (but that is disruptive), or only reset when deleting cards. Adding is always safe of course.
	*/
	countedTasks := make(map[string]int)
	for _, task := range tasks {
		count, exists := countedTasks[task.Name]
		if exists {
			countedTasks[task.Name] = count + 1
		} else {
			countedTasks[task.Name] = 1
		}
	}

	c.HTML(http.StatusOK, "domain.tmpl.html", gin.H{
		"minion": minion,
		"domain": domain,
		"tasks":  countedTasks,
	})

}
