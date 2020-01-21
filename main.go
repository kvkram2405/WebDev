package main

import (
	"database/sql"
	"log"
	"net/http"
       "fmt"
"github.com/gorilla/mux"
	_ "github.com/go-sql-driver/mysql"
)

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "root@123"
	dbName := "gosql"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}


func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
}

func web (w http.ResponseWriter, r *http.Request) {
      fmt.Fprintf(w, "Welcome to my website!")
 }


func main() {
r := mux.NewRouter()
	log.Println("Server started on: http://localhost:3000")
	r.HandleFunc("/", Index)
r.HandleFunc("/web", web)

 r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        title := vars["title"]
        page := vars["page"]

        fmt.Fprintf(w, "You've requested the book: %s on page %s\n", title, page)
    })

	http.ListenAndServe(":3000", r)
}

