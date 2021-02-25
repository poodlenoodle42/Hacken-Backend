package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/poodlenoodle42/Hacken-Backend/config"
	"github.com/poodlenoodle42/Hacken-Backend/database"
	"github.com/poodlenoodle42/Hacken-Backend/handels"
)

func main() {
	config := config.ReadConfig("config/config.yaml")

	f, err := os.OpenFile("log.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)

	database.InitDB(config.DBName, config.DBUser, config.DBPassword)
	defer database.CloseDB()

	r := mux.NewRouter().StrictSlash(true)
	//Use for unautherized route

	s := r.PathPrefix("/auth").Subrouter()
	s.Use(handels.AuthToken)
	s.HandleFunc("/groups", handels.GetGroups).Methods("GET")
	s.HandleFunc("/{groupID}/tasks", handels.GetTasks).Methods("GET")
	s.HandleFunc("/tasks/{taskID}/subtasks", handels.GetSubtasks).Methods("GET")
	fmt.Println("Started serving")
	err = http.ListenAndServe(":8080", s)
	if err != nil {
		log.Panic(err)
	}
	//End

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
