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

#### prerequisites

- accounts: gitub, heroku, travis-ci.org, codecov.io (latter 2 through github)
- Heroku, Git, Go, Postgres (+ CLI tools)

### initial setup

- install Go stuff
	- go get -u github.com/gin-gonic/gin
		- web framework
	- set -x PATH $PATH $GOPATH/bin/

- create a new repo on github
- write some code + tests
- go get -u github.com/kardianos/govendor
	- so we can vendor things, and Heroku needs this for a buildpack
	- https://gocodecloud.com/blog/2016/03/29/go-vendoring-beginner-tutorial/
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

go run main.go

and/or

go install
heroku local


# Ideas

Might be nice ot have a domain like tm.interdictor.org or somehting at least.
