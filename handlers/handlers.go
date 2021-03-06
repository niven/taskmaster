package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/niven/taskmaster/config"
	. "github.com/niven/taskmaster/data"
	"github.com/niven/taskmaster/db"
	"github.com/niven/taskmaster/logic"
)

var conf *oauth2.Config
var state string

func init() {

	conf = &oauth2.Config{
		ClientID:     "406866902910-omkqfc94h59m45a3120j6k6duic3masd.apps.googleusercontent.com",
		ClientSecret: config.EnvironmentVars["TASKMASTER_OAUTH_CLIENT_SECRET"],
		RedirectURL:  config.EnvironmentVars["BASE_URL"] + "auth",
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
			WelcomeHandler(c)
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

func AuthHandler(c *gin.Context) {

	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		ErrorHandler(c, fmt.Sprintf("Invalid session state: %s !+ %s", retrievedState, c.Query("state")), nil)
		return
	}

	tok, err := conf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		ErrorHandler(c, "", err)
		return
	}

	client := conf.Client(oauth2.NoContext, tok)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		ErrorHandler(c, "", err)
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
		ErrorHandler(c, "Error while saving session. Please try again.", err)
		return
	}

	c.HTML(http.StatusOK, "index.tmpl.html", nil)

}

func IndexHandler(c *gin.Context) {
	if isAuthorized(c) {
		OverviewHandler(c)
	} else {
		WelcomeHandler(c)
	}
}

func OverviewHandler(c *gin.Context) {

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
			ErrorHandler(c, "", err)
			return
		}
		db.LoadMinion(userEmail, &minion)
	}

	domains := db.GetDomainsForMinion(minion)

	err := logic.Update(minion)
	if err != nil {
		ErrorHandler(c, "", err)
		return
	}

	// get all tasks for each domain: everything pending (for today/this week) & today's task
	pendingTaskAssignments := db.AssignmentRetrieveForMinion(minion, false)

	// split in Today, This Week, Overdue
	now := time.Now()
	today, this_week, overdue := logic.SplitTaskAssignments(pendingTaskAssignments, now)

	c.HTML(http.StatusOK, "index.tmpl.html", gin.H{
		"minion":    minion,
		"domains":   domains,
		"pending":   today,
		"this_week": this_week,
		"overdue":   overdue,
		"today":     now.Format("Monday January 2"),
	})

}

func WelcomeHandler(c *gin.Context) {

	state = randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()

	c.HTML(http.StatusOK, "welcome.tmpl.html", gin.H{
		"login_url": getLoginURL(state),
	})
}

func SetupHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		ErrorHandler(c, "User authenticated but not found", nil)
		return
	}

	domains := db.GetDomainsForMinion(minion)

	c.HTML(http.StatusOK, "setup.tmpl.html", gin.H{
		"minion":  minion,
		"domains": domains,
	})
}

func TaskDoneHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		ErrorHandler(c, "User authenticated but not found", nil)
		return
	}

	paramTaskAssignmentID, presentTaskAssignmentID := c.GetPostForm("task_assignment_id")
	paramReturnTask, presentReturnTask := c.GetPostForm("return_task")
	if !presentTaskAssignmentID || !presentReturnTask {
		ErrorHandler(c, "Missing parameters", nil)
		return
	}

	taskAssignmentID, err := strconv.Atoi(paramTaskAssignmentID)
	if err != nil {
		ErrorHandler(c, "Invalid task assignment ID", err)
		return
	}

	assignment := db.AssignmentRetrieve(int64(taskAssignmentID))
	if assignment == nil {
		ErrorHandler(c, "No such assignment", err)
		return
	}

	if paramReturnTask == "true" {
		assignment.Status = DoneAndAvailable
	} else {
		assignment.Status = DoneAndStashed
	}

	db.AssignmentUpdate(*assignment)

	c.JSON(http.StatusOK, nil)
}

func TaskNewHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		ErrorHandler(c, "User authenticated but not found", nil)
		return
	}

	paramCount, presentCount := c.GetPostForm("count")
	paramDomainID, presentDomainID := c.GetPostForm("domain_id")
	if !presentCount || !presentDomainID {
		ErrorHandler(c, "Missing parameters", nil)
		return
	}

	count, err := strconv.Atoi(paramCount)
	if err != nil {
		ErrorHandler(c, "Invalid count", err)
		return
	}
	domainID, err := strconv.Atoi(paramDomainID)
	if err != nil {
		ErrorHandler(c, fmt.Sprintf("Invalid Domain ID: '%s'", paramDomainID), err)
		return
	}
	domain, err := db.GetDomainByID(uint32(domainID))
	if err != nil || domain.Owner != minion.ID {
		ErrorHandler(c, "Domain not found", err)
		return
	}

	name := c.DefaultPostForm("name", "Unnamed Task")
	var weekly bool
	if c.DefaultPostForm("weekly", "false") == "false" {
		weekly = false
	} else {
		weekly = true
	}

	task := Task{
		Name:     name,
		DomainID: uint32(domainID),
		Weekly:   weekly,
		Count:    uint32(count),
	}

	err = db.CreateNewTask(task)
	if err != nil {
		ErrorHandler(c, "Error creating new task", err)
		return
	}

	// add the domain ID to the params so we can chain to another handler
	// A redirect is kind of weird, and DomainEdit doesn't accept a POST anyway
	c.Params = append(c.Params, gin.Param{Key: "domain_id", Value: paramDomainID})
	DomainEditHandler(c)

}

func DomainNewHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		ErrorHandler(c, "User authenticated but not found", nil)
		return
	}

	domainName := c.DefaultPostForm("name", "Unnamed Deck")
	db.CreateNewDomain(minion, domainName)

	SetupHandler(c)
}

func DomainDeleteHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		ErrorHandler(c, "User authenticated but not found", nil)
		return
	}

	domainID, err := strconv.Atoi(c.Param("domain_id"))
	if err != nil || domainID < 0 {
		ErrorHandler(c, "Invalid domain ID", err)
		return
	}

	domain, err := db.GetDomainByID(uint32(domainID))
	if err != nil || domain.Owner != minion.ID {
		ErrorHandler(c, "Domain not found", err)
		return
	}

	db.DomainDelete(domain)

	domains := db.GetDomainsForMinion(minion)

	c.HTML(http.StatusOK, "setup.tmpl.html", gin.H{
		"minion":  minion,
		"domains": domains,
	})
}

func ErrorHandler(c *gin.Context, message string, err error) {

	c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
		"message": message,
		"error":   err,
	})

}

func DomainEditHandler(c *gin.Context) {

	session := sessions.Default(c)
	userEmail := session.Get("user-id").(string)
	var minion Minion
	found := db.LoadMinion(userEmail, &minion)
	if !found {
		ErrorHandler(c, "User authenticated but not found", nil)
	}

	domainID, err := strconv.Atoi(c.Param("domain_id"))
	if err != nil || domainID < 0 {
		ErrorHandler(c, "Invalid domain ID", err)
		return
	}

	domain, err := db.GetDomainByID(uint32(domainID))
	if err != nil || domain.Owner != minion.ID {
		ErrorHandler(c, "Domain not found", err)
		return
	}

	tasks, err := db.GetTasksForDomain(domain)
	if err != nil {
		// return not found to avoid leaking domain IDs. Not that it matters here, but general principle
		ErrorHandler(c, "Domain not found", err)
		return
	}
	// setup menu needs the list
	domains := db.GetDomainsForMinion(minion)

	c.HTML(http.StatusOK, "domain.tmpl.html", gin.H{
		"minion":  minion,
		"domain":  domain,
		"domains": domains,
		"daily":   TaskFilter(tasks, func(t Task) bool { return !t.Weekly }),
		"weekly":  TaskFilter(tasks, func(t Task) bool { return t.Weekly }),
	})

}
