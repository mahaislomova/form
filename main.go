package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

var router = mux.NewRouter()

func indexPage(w http.ResponseWriter, r *http.Request) {
	msg := getMsg(w, r, "message")
	var u = &User{}
	u.Errors = make(map[string]string)
	if msg != "" {
		u.Errors["message"] = msg
		render(w, "signin", u)
	} else {
		u := &User{}
		render(w, "signin", u)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("uname")
	pass := r.FormValue("password")
	u := &User{Username: name, Password: pass}
	redirect := "/"
	if name != "" && pass != "" {
		if b, uuid := userExists(u); b == true {
			setSession(&User{Uuid: uuid}, w)
			redirect = "/example"
		} else {
			setMsg(w, "message", "Please signup or enter a valid username and password!")
		}
	} else {
		setMsg(w, "message", "Username or Password field are empty!")
	}
	http.Redirect(w, r, redirect, 302)
}

func logout(w http.ResponseWriter, r *http.Request) {
	clearSession(w, "session")
	http.Redirect(w, r, "/", 302)
}

func examplePage(w http.ResponseWriter, r *http.Request) {
	uuid := getUuid(r)
	u := getUserFromUuid(uuid)
	if uuid != "" {
		render(w, "internal", u)
	} else {
		setMsg(w, "message", "Please login first!")
		http.Redirect(w, r, "/", 302)
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		u := &User{}
		u.Errors = make(map[string]string)
		u.Errors["username"] = getMsg(w, r, "username")
		u.Errors["password"] = getMsg(w, r, "password")
		render(w, "signup", u)
	case "POST":
		if n := checkUser(r.FormValue("userName")); n == true {
			setMsg(w, "username", "User already exists. Please enter a unique username!")
			http.Redirect(w, r, "/signup", 302)
			return
		}
		u := &User{
			Uuid: Uuid(),
			Username: r.FormValue("userName"),
			Password: r.FormValue("password"),
		}
		result, err := govalidator.ValidateStruct(u)
		if err != nil {
			e := err.Error()
			if re := strings.Contains(e, "Username"); re == true {
				setMsg(w, "username", "Please enter a valid Username!")
			}
			if re := strings.Contains(e, "Password"); re == true {
				setMsg(w, "password", "Please enter a Password!")
			}
		}
		if len(r.FormValue("userName")) <= 4 {
			setMsg(w, "username", "The username must be at least 5 characters long!")
			http.Redirect(w, r, "/signup", 302)
			return
		}
		if !validPassword(r.FormValue("password")) {
			setMsg(w, "password", "The password must be at least 8 characters long, contain at least 1 number, 1 uppercase and 1 lowercase character")
			http.Redirect(w, r, "/signup", 302)
			return
		}
		if r.FormValue("password") != r.FormValue("cpassword"){
			setMsg(w, "password", "The passwords you entered do not Match!")
			http.Redirect(w, r, "/signup", 302)
			return
		}

		if result == true {
			u.Password = encryptPass(u.Password)
			saveData(u)
			http.Redirect(w, r, "/", 302)
			return
		}
		http.Redirect(w, r, "/signup", 302)
	}
}

func render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.ParseGlob("*.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tmpl.ExecuteTemplate(w, name, data)
}

func main() {
	govalidator.SetFieldsRequiredByDefault(true)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/", indexPage)
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("POST")
	router.HandleFunc("/example", examplePage)
	router.HandleFunc("/signup", signup).Methods("POST", "GET")
	http.Handle("/", router)
	http.ListenAndServe(":8000", nil)
}