# Task Master
Task Master helps you get all your chores done!

When it comes to chores, there are infrequent ones you forget about and those you hate. The end result is that some things never get done.

The solution seems obvious: a fixed chore schedule. This also doesn't work because sometimes you don't have the time and then things pile up. Even worse, it feels like you have no control over what you do.

Task Master solves this in a fun way: every day you get a Task from a (virtual) deck of cards that has 1 task for you to do. This can be either something you do today, or this week. Weekly tasks are things like cleaning the fridge you might not get around to during the week. After you have done a task you can choose to keep the card or return it to the deck. This allows you to withhold tasks you don't like doing and increases the chances you get the ones you do.

The deck consists of chores based on the number of people and the types of tasks. Examples are a couple living together doing housekeeping chores, a club sharing a workspace or anything else. It's up to you!

# Technology

I've already made a basic version in JavaScript that uses local storage and is kind of fun. This version has the aim of getting some experience using Heroku, doing a bit of CI and some Go since that has been a while.


## What happens where?

The code is hosted on github. When I push Travis CI runs tests and a code coverage report which is then uploaded to CodeCov.io

The thing then runs on Heroku

### prerequisites

- accounts: gitub, heroku, travis-ci.org, codecov.io (latter 2 through github)
- Heroku, Git, Go, Postgres (+ CLI tools)

### initial setup

#### Go dependencies

(fish shell)
set -x PATH $PATH $GOPATH/bin/

web framework:
go get -u github.com/gin-gonic/gin

i18n text stuff:
go get golang.org/x/text/language
go get golang.org/x/text/message
go get -u golang.org/x/text/cmd/gotext

so we can vendor things, and Heroku needs this for a buildpack
https://gocodecloud.com/blog/2016/03/29/go-vendoring-beginner-tutorial/
go get -u github.com/kardianos/govendor

Postgres driver and stuff:
go get github.com/lib/pq

Other stuff: google-oauth, various gin contrib

#### Steps

- create a new repo on github
- write some code + tests
- govendor init
	- git add -A vendor
	- git commit -am "Setup Vendor"
- govendor fetch github.com/gin-gonic/gin	
- push to github
- add the repo to travis CI / CodeCov
- add a .travis.yml file with Codecov coverage upload
- push to github, check Travis if everything builds
- heroku apps:create
	- unnamed app since app names are global
	- I got: https://peaceful-everglades-27897.herokuapp.com/ | https://git.heroku.com/peaceful-everglades-27897.git
- git remote -v
	heroku	https://git.heroku.com/peaceful-everglades-27897.git (fetch)
	heroku	https://git.heroku.com/peaceful-everglades-27897.git (push)
	origin	git@github.com:niven/taskmaster (fetch)
	origin	git@github.com:niven/taskmaster (push)	
- Tell Heroku we are a Go app
	heroku buildpacks:set heroku/go
- git push heroku master
- add a Procfile
	contents = web: taskmaster
	
#### Setting up OAUTH with Google

https://support.google.com/cloud/answer/6158849?hl=en
https://console.developers.google.com/iam-admin/projects
Create a new project (in this case: taskmaster)
https://console.developers.google.com/home/dashboard?project=taskmaster

Much mucking about

client ID:
406866902910-omkqfc94h59m45a3120j6k6duic3masd.apps.googleusercontent.com
client secret:
[stored elsewhere]

Set a domain in your /etc/hosts file
127.0.0.1	taskmaster.org
Then enter that one as the Authorized Domains for your OAuth stuff
Note: the UI is garbage, you need to hit enter to add new items to a list and then click a save button


#### Running locally

set -x DATABASE_URL postgres://localhost/taskmaster\?sslmode=disable
set -x PORT 5000
set -x TASKMASTER_OAUTH_CLIENT_SECRET ...

go run main.go taskmaster.go handlers.go

and/or

go install
heroku local


#### Setting up a database

First, do it locally and test stuff. Then later dump the local dir and export it to a Heroku psql instance

Set it up with heroku
heroku addons:create heroku-postgresql:hobby-dev
This creates DATABASE_URL, which we should also set locally
Use 
	heroku pg:psql
To query the remote db

Set it locally (fish shell):
set -x DATABASE_URL postgres://taskmaster

run table definitions (and for testing insert the test data)
first table, for users:
createdb taskmaster
psql taskmaster
(TODO: don't store, or store emails/names encrypted)

go get github.com/lib/pq

##### Table Definitions

Moved to database/update_NNN.sql
Automatically run with

	go run database/db_manage.go

That will pick up any changes to the db newer than the update point in the "version" table.
If there is no version, update_000.sql and any higher will run, effectiveky doing a clean db setup

##### Test Data

INSERT INTO minions (id, email, name) VALUES (1, 'gru@minions.com', 'Gru');
INSERT INTO domains (owner, name) VALUES (1, 'Tree House');
INSERT INTO tasks (domain_id, name, weekly) VALUES (1, 'Remove leaves', false), (1, 'Wash window', true);

#### drop tables for testing
DROP TABLE IF EXISTS task_assignments;
DROP TYPE IF EXISTS enum_status;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS minion_domain;
DROP TABLE IF EXISTS domains;
DROP TABLE IF EXISTS minions;
DROP TABLE IF EXISTS version;


# Ideas

Might be nice ot have a domain like tm.interdictor.org or somehting at least.

Depending on locale weekends might not be Sat+Sun
