package handels

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/poodlenoodle42/Hacken-Backend/database"
	"github.com/poodlenoodle42/Hacken-Backend/datastructures"
)

func GetTasks(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	token = strings.TrimSpace(token)
	vars := mux.Vars(r)
	groupIDstring := vars["groupID"]
	groupID, err := strconv.Atoi(groupIDstring)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ex, err := database.DoesGroupExists(groupID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if !ex {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Group does not exist"))
		return
	}
	tasks, err := database.GetTasksForGroup(token, groupID)
	if err != nil {
		if err.Error() == "User not allowed to view Group details" {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "text/json")
	json.NewEncoder(w).Encode(tasks)
}

func GetSubtasks(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	token = strings.TrimSpace(token)
	vars := mux.Vars(r)
	taskIDstring := vars["taskID"]
	taskID, err := strconv.Atoi(taskIDstring)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	subtasks, err := database.GetSubtasksForTasks(taskID, token)
	if err != nil {
		if err.Error() == "User not allowed to view Group details" {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "text/json")
	json.NewEncoder(w).Encode(subtasks)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	token = strings.TrimSpace(token)
	vars := mux.Vars(r)
	taskIDstring := vars["taskID"]
	taskID, err := strconv.Atoi(taskIDstring)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	allowed, err := database.IsUserAllowedToAccessTask(token, taskID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("User is not allowed to view task"))
		return
	}
	http.ServeFile(w, r, DataDir+fmt.Sprintf("/tasks/%d.pdf", taskID))
}

func NewTask(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	token := r.Header.Get("token")
	token = strings.TrimSpace(token)
	vars := mux.Vars(r)
	groupIDstring := vars["groupID"]
	groupID, err := strconv.Atoi(groupIDstring)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	file, _, err := r.FormFile("pdf")
	defer file.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	taskInfoJSON := r.FormValue("task")
	if taskInfoJSON == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No task info"))
		return
	}
	var taskInfo datastructures.Task
	err = json.Unmarshal([]byte(taskInfoJSON), &taskInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	admin, err := database.IsUserAdminOfGroup(token, groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if !admin {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("User is not admin of Group"))
		return
	}
	taskInfo.GroupID = groupID
	taskInfo.Author = token
	task, err := database.AddTask(taskInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	f, err := os.OpenFile(DataDir+fmt.Sprintf("/tasks/%d.pdf", task.ID), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()
	io.Copy(f, file)
	json.NewEncoder(w).Encode(task)
}
