package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Container struct {
	Id         int    `json:"id"`
	Host_id    int    `json:"host_id"`
	Name       string `json:"name"`
	Image_name string `json:"image_name"`
	Host_name  string `json:"host_name"`
}
type Host struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	Name       string `json:"name"`
	Ip_address string `json:"ip_address"`
}

var sqliteDatabase *sql.DB

func main() {
	handleRequests()
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/hosts", displayHosts).Methods("GET")
	myRouter.HandleFunc("/containers", enterNewContainer).Methods("PUT")
	myRouter.HandleFunc("/containers", displayContainers).Methods("GET")
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}
func enterNewContainer(w http.ResponseWriter, r *http.Request) {
	var c Container
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&c)
	if err != nil {
		fmt.Println(err)
	}
	sqliteDatabase, _ := sql.Open("sqlite3", "./aqua.db") // Open the created SQLite File
	log.Println("tyring to open aqua db")
	log.Println("Inserting container record ...")
	rows, err := sqliteDatabase.Query("SELECT COUNT(*) FROM containers")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var id int

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
	}
	statement, err := sqliteDatabase.Prepare("INSERT INTO containers (id, host_id, name, image_name) values (?,?,?,?)") // Prepare statement.
	if err != nil {
		log.Fatalln(err.Error())
	}
	statement.Exec(id, c.Host_id, RandomString(10), c.Image_name)

	row, err := sqliteDatabase.Query("SELECT hosts.name as host_name,containers.id,containers.host_id ,containers.name, containers.image_name from hosts inner join containers on hosts.id= containers.host_id WHERE containers.id=?", id)
	if err != nil {
		log.Fatal(err)
	}
	ContainerRowLoop(row, w)
	defer row.Close()
	defer sqliteDatabase.Close()
}
func ContainerRowLoop(row *sql.Rows, w http.ResponseWriter) {
	for row.Next() {
		var id1 int
		var host_id int
		var name string
		var image_name string
		var host_name string
		row.Scan(&host_name, &id1, &host_id, &name, &image_name)
		b, err := json.Marshal(Container{id1, host_id, name, image_name, host_name})
		if err != nil {
			fmt.Println(err)

			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write(b)
	}
}
func displayContainers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key1 := q.Get("id")
	Query := "SELECT hosts.name as host_name,containers.id,containers.host_id ,containers.name, containers.image_name from hosts inner join containers on hosts.id= containers.host_id"
	if key1 == "" {
		key1 = q.Get("host_id")
		if key1 != "" {
			Query += " WHERE containers.host_id = " + key1
		}
	} else {
		Query += " WHERE containers.id = " + key1
	}
	sqliteDatabase, _ := sql.Open("sqlite3", "./aqua.db") // Open the created SQLite File
	log.Println("tyring to open aqua db")
	row, err := sqliteDatabase.Query(Query)
	if err != nil {
		log.Fatal(err)
	}
	ContainerRowLoop(row, w)

	defer sqliteDatabase.Close()
	defer row.Close()
}
func displayHosts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key1 := q.Get("id")
	Query := "SELECT * FROM hosts"
	if key1 != "" {
		Query += " WHERE id =" + key1
	}
	sqliteDatabase, _ := sql.Open("sqlite3", "./aqua.db") // Open the created SQLite File
	log.Println("tyring to open aqua db")
	row, err := sqliteDatabase.Query(Query)
	if err != nil {
		log.Fatal(err)
	}
	HostRowLoop(row, w)
	defer row.Close()

	defer sqliteDatabase.Close()

}
func HostRowLoop(row *sql.Rows, w http.ResponseWriter) {
	for row.Next() { // Iterate and fetch the records from result cursor
		var id int
		var uuid string
		var name string
		var ip_address string
		row.Scan(&id, &uuid, &name, &ip_address)
		b, err := json.Marshal(Host{id, uuid, name, ip_address})
		if err != nil {
			fmt.Println(err)

			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write(b)
	}
}
