package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type UserCreateRequest struct {
	Id       int64  `json:"id"`
	UserName string `json:"user_name"`
}

type UserCreateResponse struct {
	Token string `json:"token"`
}

type UserGetResponse struct {
	Name string
}

type UserUpdateRequest struct {
	Name string
}

var MYSQL string = "mysql"
var DB string = "docker:docker@tcp(mysql_host:3306)/game_user"

func createToken(id int64, userName string) string {

	// headerのセット
	token := jwt.New(jwt.SigningMethodHS256)

	// claimsのセット
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = id
	claims["name"] = userName

	// 電子署名
	// tokenString, _ := token.SignedString([]byte(os.Getenv("SIGNINGKEY")))
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		log.Printf(err.Error())
	}

	// JWTを返却
	return tokenString
}

func initMysql() {
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
	defer r.Body.Close()
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
	userCreReq.Id = lastId //int64 -> int or int -> int64 huummmm

	fmt.Fprintf(w, "%d Created User!\n", http.StatusCreated)
	userCreRes := &UserCreateResponse{}
	userCreRes.Token = createToken(userCreReq.Id, userCreReq.UserName)
	json.NewEncoder(w).Encode(userCreRes)

}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/user/create", userCreateRequest).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	initMysql()
	handleRequests()
}
