package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type UserCreateRequest struct {
	Id       int    `json:"id"`
	UserName string `json:"user_name"`
}

type UserCreateResponse struct {
	Token string
}

type UserGetResponse struct {
	Name string
}

type UserUpdateRequest struct {
	Name string
}

var MYSQL string = "mysql"
var DB string = "docker:docker@tcp(mysql_host:3306)/game_user"

func init_mysql() {
	db, err := sql.Open(MYSQL, DB)
	if err != nil {
		log.Printf(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("Connection failed")
	} else {
		log.Printf("Connection Successful")
	}
}

func userCreateRequest(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(MYSQL, DB)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	NameIns, err := db.Prepare("INSERT INTO users(user_name) VALUES( ? )") // ? = placeholder
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer NameIns.Close()
	reqBody, err := ioutil.ReadAll(r.Body) // []uint8 byte stream
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	userCreReq := &UserCreateRequest{}
	err = json.Unmarshal(reqBody, &userCreReq)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	res, err := NameIns.Exec(userCreReq.UserName)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	lastId, err := res.LastInsertId() // inserted row
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Printf("ID = %d\n", lastId)
	log.Printf("%d Created User!", http.StatusCreated)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!\n")
	fmt.Println("Endpoint Hit: homePage")
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/user/create", userCreateRequest).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	init_mysql()
	handleRequests()
	// userCreateResponse := &UserCreateResponse{}
	// post_handle := func(w http.ResponseWriter, r *http.Request) {
	// 	// user_name, err := r.GetBody()
	// 	// if err != nil {
	// 	// 	log.Printf("%s",err.Error)
	// 	// }
	// 	userCreateRequest := &UserCreateRequest{}
	// 	// err = json.NewDecoder(user_name).Decode(userCreateRequest)
	// 	err := json.NewDecoder(r.Body).Decode(userCreateRequest)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// 	user_id := userCreateRequest.Name + "fugafuga"
	// 	userCreateResponse.Token = "ggggg"
	// 	// writing operation to sql
	// 	log.Printf("%s\n%s\n%s", userCreateRequest.Name, user_id, userCreateResponse.Token)
	// }
	// // http.HandleFunc("/user/create", post_handle)
	// http.HandleFunc("/", post_handle)
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
