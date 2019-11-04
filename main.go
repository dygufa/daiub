package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/subosito/gotenv"
)

func init() {
	gotenv.Load()
}

type UpdatePayload struct {
	LastStart *time.Time
	Duration  int
}

type Session struct {
	gorm.Model
	UserName         string `gorm:"unique;not null"`
	OrganizationName string
	Duration         int
	LastStart        *time.Time
}

var db *gorm.DB
var dberr error

func main() {
	db, dberr = gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if dberr != nil {
		fmt.Print(dberr)
	}
	defer db.Close()

	db.AutoMigrate(&Session{})

	router := mux.NewRouter()

	router.HandleFunc("/session", getSessions).Queries("organization_name", "{organization_name}").Methods("GET")
	router.HandleFunc("/session", createSession).Methods("POST")
	router.HandleFunc("/session/{id}", updateSession).Methods("PUT")

	port := os.Getenv("PORT")
	if port == "" {
		port = "9015"
	}

	fmt.Printf("Inicializando na porta: %s", port)
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		fmt.Print(err)
	}
}

func getSessions(w http.ResponseWriter, r *http.Request) {
	organizationName := r.FormValue("organization_name")
	var sessions []Session
	db.Where("organization_name = ?", organizationName).Find(&sessions)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func createSession(w http.ResponseWriter, r *http.Request) {
	var newSession Session
	_ = json.NewDecoder(r.Body).Decode(&newSession)
	db.Create(&newSession)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newSession)
}

func updateSession(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	var session Session
	db.First(&session, id)

	var updateData UpdatePayload
	_ = json.NewDecoder(r.Body).Decode(&updateData)

	session.LastStart = updateData.LastStart
	session.Duration = updateData.Duration

	db.Save(&session)
	json.NewEncoder(w).Encode(&session)
}
