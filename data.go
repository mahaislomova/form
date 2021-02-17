package main

import (
	"database/sql"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Uuid     string	`valid:"required,uuidv4"`
	Username string	`valid:"required,alphanum"`
	Password string	`valid:"required"`
	Errors   map[string]string `valid:"-"`
}

func saveData(u *User) error {
	var db, _ = sql.Open("sqlite3", "users.sqlite3")
	defer db.Close()
	db.Exec("create table if not exists users (uuid text not null unique, username text not null unique, password text not null, primary key(uuid))")
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into users (uuid, username, password) values ( ?, ?, ?)")
	_, err := stmt.Exec(u.Uuid, u.Username, u.Password)
	tx.Commit()
	return err
}

func userExists(u *User) (bool, string)  {
	var db, _ = sql.Open("sqlite3", "users.sqlite3")
	defer db.Close()
	var ps, uu string
	q, err := db.Query("select uuid, password from users where username = '" + u.Username +"'")
	if err != nil {
		return false, ""
	}
	for q.Next() {
		q.Scan(&uu, &ps)
	}
	pw := bcrypt.CompareHashAndPassword([]byte(ps), []byte(u.Password))
	if uu != "" && pw == nil {
		return true, uu
	}
	return false, ""
}

func validPassword(passw string) bool {
	if len(passw) <= 7 {
		return false
	}
	num := `[0-9]{1}`
	a_z := `[a-z]{1}`
	A_Z := `[A-Z]{1}`

	if b, err := regexp.MatchString(num, passw); !b || err != nil {
		return false
	}
	if b, err := regexp.MatchString(a_z, passw); !b || err != nil {
		return false
	}
	if b, err := regexp.MatchString(A_Z, passw); !b || err != nil {
		return false
	}
	return true
}

func checkUser(user string) bool {
	var db, _ = sql.Open("sqlite3", "users.sqlite3")
	defer db.Close()
	var un string
	q, err := db.Query("select username from users where username = '" + user + "'")
	if err != nil {
		return false
	}
	for q.Next(){
		q.Scan(&un)
	}
	if un == user {
		return true
	}
	return false
}

func getUserFromUuid(uuid string) *User {
	var db, _ = sql.Open("sqlite3", "users.sqlite3")
	defer db.Close()
	var uu, un, pass string
	q, err := db.Query("select * from users where uuid = '" + uuid + "'")
	if err != nil {
		return &User{}
	}
	for q.Next(){
		q.Scan(&uu, &un, &pass)
	}
	return &User{Username: un, Password: pass}
}

func encryptPass(password string) string {
	pass := []byte(password)
	hashpw, _ := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	return string(hashpw)
}

func Uuid() string {
	u := uuid.NewV4() //this function was broken in one of satori's  PRs
	id := uuid.Must(u, nil)
	return id.String()
}