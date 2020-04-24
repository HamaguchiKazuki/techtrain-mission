package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
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
var SigningKey = []byte("secret")

func createToken(id int64, userName string) string {

	// headerのセット
	token := jwt.New(jwt.SigningMethodHS256)

	// claimsのセット
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = id
	claims["name"] = userName

	// 電子署名
	// tokenString, _ := token.SignedString([]byte(os.Getenv("SIGNINGKEY")))
	tokenString, err := token.SignedString(SigningKey)
	if err != nil {
		log.Printf(err.Error())
	}

	// JWTを返却
	return tokenString
}

func AuthMiddleware(next http.Handler) http.Handler {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return SigningKey, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	return jwtMiddleware.Handler(next)
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

func userCreateHandler(w http.ResponseWriter, r *http.Request) {
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

func userGetHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open(MYSQL, DB)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	userCreReq := &UserCreateRequest{}
	err = db.QueryRow("SELECT id, user_name FROM users WHERE id = ?", 2).Scan(&userCreReq.Id, &userCreReq.UserName)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	log.Println(userCreReq.Id)
	log.Println(userCreReq.UserName)

}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/user/create", userCreateHandler).Methods("POST")
	myRouter.Handle("/user/get", AuthMiddleware(http.HandlerFunc(userGetHandler))).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	initMysql()
	handleRequests()
}
