package main

import (
	"fmt"
	"github.com/snackbag/compass/v2"
	"strings"
)

var server *compass.Server

type User struct {
	Name     string
	Password string // DO NOT SAVE PASSWORDS LIKE THIS!!!!!!!!!
}

var users = make(map[string]User)

func main() {
	server = compass.NewServer(compass.NewStandardConfiguration())

	//
	// Page that changes its content based on whether you are logged in or not
	//
	server.AddRoute("/", func(request compass.Request) compass.Response {
		session, ok := request.GetSession(server)
		if !ok {
			return compass.ServeFile("index.html", "index.html")
		}

		id, err := compass.SessionGet[string](session, "name")
		if err != nil {
			return compass.InternalError("corrupted session, is missing name")
		}

		user, ok := users[id]
		if !ok {
			session.MustDestroy()
			return compass.Text("Your session contains weird data. Try again")
		}

		return compass.HTML(fmt.Sprintf("<html><p>You are logged in as %s; password %s. <a href=\"/logout\">Log out.</a></html>", user.Name, user.Password))
	})

	//
	// Separate page that is only available to logged in people
	//
	server.AddRoute("/special", func(request compass.Request) compass.Response {
		session, ok := request.GetSession(server)
		if !ok {
			return compass.Redirect("/", false)
		}

		return compass.Text(fmt.Sprintf("Hi %s!", compass.SessionGetOrDefault[string](session, "name", "fallback")))
	})

	//
	// Page that shows a login/register dialog
	//
	routeLogin := server.AddRoute("/login", func(request compass.Request) compass.Response {
		// if already logged in, do not allow new login
		_, ok := request.GetSession(server)
		if ok {
			return compass.Redirect("/", false)
		}

		if request.Method == "get" {
			return compass.ServeFile("login.html", "login.html")
		} else if request.Method == "post" {
			err := request.Http.ParseForm()
			if err != nil {
				return compass.InternalError(fmt.Sprintf("failed to parse form: %s", err))
			}

			action := request.Http.FormValue("action")
			username := strings.ToLower(request.Http.FormValue("username"))
			password := request.Http.FormValue("password")

			switch action {
			case "Login":
				return handleLogin(username, password)
			case "Register":
				return handleRegister(username, password)
			default:
				return compass.Text("Invalid action")
			}
		}

		return compass.InternalError("somehow triggered illegal method")
	})
	routeLogin.AllowedMethods = append(routeLogin.AllowedMethods, "post")

	//
	// Route that logs you out
	//
	server.AddRoute("/logout", func(request compass.Request) compass.Response {
		session, ok := request.GetSession(server)
		if !ok {
			return compass.Text("You must be logged in to log out")
		}

		session.MustDestroy()
		return compass.Redirect("/", false)
	})

	server.MustRun()
}

func handleRegister(username string, password string) compass.Response {
	_, ok := users[username]
	if ok {
		return compass.Text("User with same name is already registered!")
	}

	users[username] = User{
		Name:     username,
		Password: password,
	}

	session, err := makeSession(username)
	if err != nil {
		return *err
	}

	resp := compass.Redirect("/", false)
	resp.SetSession(session)
	return resp
}

func handleLogin(username string, password string) compass.Response {
	user, ok := users[username]
	if !ok {
		return compass.Text("Username or password incorrect")
	}

	if user.Password != password {
		return compass.Text("Username or password incorrect")
	}

	session, err := makeSession(username)
	if err != nil {
		return *err
	}

	resp := compass.Redirect("/", false)
	resp.SetSession(session)
	return resp
}

func makeSession(username string) (*compass.Session, *compass.Response) {
	session, err := server.CreateSession()
	if err != nil {
		resp := compass.InternalError(err.Error())
		return session, &resp
	}

	tx := session.BeginTx()
	tx.Set("name", username)
	tx.Commit()

	return session, nil
}
