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
	"github.com/gorilla/sessions"
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

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

type Employee struct {
	Id         int
	Username   string
	Password   string
	Created_at string
}

var tmpl = template.Must(template.ParseGlob("form/*"))

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
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

func Register(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		t := time.Now()
		t.Format(time.RFC3339)
		insUForm, err := db.Prepare("INSERT INTO users (username,password, created_at) values (?,?,?)")

		if err != nil {
			panic(err.Error())
		}
		insUForm.Exec(username, password, t)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func web(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to my website!")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 15)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func password(w http.ResponseWriter, r *http.Request) {

	password := "secret"
	hash, _ := HashPassword(password) // ignore error for the sake of simplicity

	fmt.Println("Password:", password)
	fmt.Println("Hash:    ", hash)
	match := CheckPasswordHash(password, hash)
	fmt.Println("Match:   ", match)

}

func secret(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Print secret message
	fmt.Fprintln(w, "The cake is a lie!")
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Authentication goes here
	// ...

	// Set user as authenticated
	session.Values["authenticated"] = true
	session.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)
}

func main() {
	r := mux.NewRouter()
	log.Println("Server started on: http://localhost:3000")
	r.HandleFunc("/", Index)
	r.HandleFunc("/web", web)
	r.HandleFunc("/auth", password)
	r.HandleFunc("/secret", secret)
	r.HandleFunc("/login", login)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		title := vars["title"]
		page := vars["page"]

		fmt.Fprintf(w, "You've requested the book: %s on page %s\n", title, page)
	})
	r.HandleFunc("/show/{id}", Show)
	r.HandleFunc("/registration", Registration)
	r.HandleFunc("/register", Register)
	http.ListenAndServe(":3000", r)
}
