package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "root@123"
	dbName := "test"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

type Employee struct {
	Id         int
	Username   string
	Password   string
	Created_at string
}

var tmpl = template.Must(template.ParseGlob("form/*"))

func dashboard(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Dashboard", nil)
}
func Index(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM users ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}
	emp := Employee{}
	res := []Employee{}
	for selDB.Next() {
		var id int
		var username, password, created_at string
		err = selDB.Scan(&id, &username, &password, &created_at)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Username = username
		emp.Password = password
		emp.Created_at = created_at
		res = append(res, emp)
	}
	tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()
}
func Show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM users WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
	emp := Employee{}
	for selDB.Next() {
		var id int
		var username, password, created_at string
		err = selDB.Scan(&id, &username, &password, &created_at)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Username = username
		emp.Password = password
		emp.Created_at = created_at
	}
	tmpl.ExecuteTemplate(w, "Show", emp)
	defer db.Close()
}
func Registration(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Registration", nil)
}
func LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Login", nil)
}
func Register(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Password changed to hash
		hash, _ := HashPassword(password) // ignore error for the sake of simplicity
		t := time.Now()
		t.Format(time.RFC3339)
		insUForm, err := db.Prepare("INSERT INTO users (username,password, created_at) values (?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insUForm.Exec(username, hash, t)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 15)
	return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func login(w http.ResponseWriter, r *http.Request) {
	// Authentication goes here
	// ...
	// Compare user password and hash
	db := dbConn()
	if r.Method == "POST" {
		usernameFlogin := r.FormValue("username")
		passwordFlogin := r.FormValue("password")
		selDB, err := db.Query("SELECT * FROM users WHERE username=?", usernameFlogin)
		if err != nil {
			panic(err.Error())
		}
		for selDB.Next() {
			var id int
			var username, password, created_at string
			err = selDB.Scan(&id, &username, &password, &created_at)
			if err != nil {
				panic(err.Error())
			}
			hash1 := password                                 // Password from Database
			match := CheckPasswordHash(passwordFlogin, hash1) // Compare Password from Db and Login page
			if match == true {
				http.Redirect(w, r, "/dashboard", 301)
			} else {
				http.Redirect(w, r, "/registration", 301)
			}
		}
		defer db.Close()
	}
}
func logout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 301)
}
func main() {
	r := mux.NewRouter()
	log.Println("Server started on: http://localhost:3000")
	r.HandleFunc("/dashboard", dashboard)
	r.HandleFunc("/", Index)
	r.HandleFunc("/auth", LoginPage)
	r.HandleFunc("/login", login)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		title := vars["title"]
		page := vars["page"]
		fmt.Fprintf(w, "You've requested the book: %s on page %s\n", title, page)
	})
	r.HandleFunc("/show", Show)
	r.HandleFunc("/registration", Registration)
	r.HandleFunc("/register", Register)
	http.ListenAndServe(":3000", r)
}
