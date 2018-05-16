package main

import (
	"net/http"
	"fmt"
	"database/sql"
	"log"
	"encoding/json"
	"strconv"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type User struct {
	Id 		int
	Name 	string
	Phone 	string
}

func main(){

	initDb()

	defer db.Close()

	router := httprouter.New()

	router.GET("/api/users", getUsers)
	router.GET("/api/users/:id", getUser)
	router.POST("/api/users", addUser)
	router.DELETE("/api/users/:id", deleteUser)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func initDb(){

	var err error
	db, err = sql.Open("sqlite3", "testApi.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS people("id" INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,"name" TEXT, "phone" TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

}

func getUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {

	var strName string

	if len(r.URL.RawQuery) > 0 {
		strName = r.URL.Query().Get("id")

		if strName == "" {
			w.WriteHeader(400)
			return
		}
	}

	users, err := read(strName)

	fmt.Fprintf(w, "%v", users)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err = json.NewEncoder(w).Encode(strName); err != nil {
		w.WriteHeader(500)
	}
}

func getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	fmt.Fprintf(w,"Hell0")
}

func addUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {

	user := new(User)
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil || user.Name == "" {
		w.WriteHeader(400)
		return
	}

	if _, err := insert(user); err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
}

func deleteUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	id, ok := getID(w, ps)
	if !ok {
		return
	}
	if err := remove(id); err != nil {
		w.WriteHeader(500)
	}
	w.WriteHeader(204)
}

func insert(user *User) (sql.Result, error) {
	return db.Exec("INSERT INTO people (name, phone) VALUES ($1, $2)", user.Name, user.Phone)
}

func getID(w http.ResponseWriter, ps httprouter.Params) (int, bool) {

	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(400)
		return 0, false
	}
	return id, true
}

func read(id string) ([]User, error) {

	var rows *sql.Rows
	var err error

	if id != "" {
		rows, err = db.Query("SELECT * FROM people WHERE id LIKE $1 ORDER BY id",
			"%"+id+"%")
	} else {
		rows, err = db.Query("SELECT * FROM people")
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users = make([]User, 0)
	var record User

	for rows.Next() {
		if err = rows.Scan(&record.Id, &record.Name, &record.Phone); err != nil {
			fmt.Println(err)
			continue
		}
		users = append(users, record)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func remove(id int) error {

		_, err := db.Exec("DELETE FROM people WHERE id = $1", id)

		if err != nil{
			return err
		}

	return nil
}