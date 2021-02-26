package database

import (
	"errors"
	"log"

	"github.com/poodlenoodle42/Hacken-Backend/datastructures"
)

func getTask(taskID int) (datastructures.Task, error) {
	var task datastructures.Task
	err := db.QueryRow("SELECT id,Name,Author,Description,Group_id FROM Tasks WHERE id = ?", taskID).Scan(
		&task.ID, &task.Name, &task.Author, &task.Description, &task.GroupID)
	return task, err
}

//IsUserAllowedToAccessTask checks if a user is allowed to access a given task
func IsUserAllowedToAccessTask(token string, taskID int) (bool, error) {
	task, err := getTask(taskID)
	if err != nil {
		log.Println("IsUserAllowedToAccessTask: " + err.Error())
		return false, err
	}
	i, err := isUserInGroup(token, task.GroupID)
	if err != nil {
		log.Println("IsUserAllowedToAccessTask: " + err.Error())
		return false, err
	}
	return i, nil
}

//GetSubtasksForTasks checks if user is allowed to view task and if so returns all subtasks for a task
func GetSubtasksForTasks(taskID int, token string) ([]datastructures.Subtask, error) {
	task, err := getTask(taskID)
	if err != nil {
		log.Println("GetSubtasksForTasks: " + err.Error())
		return nil, err
	}
	isUserAllowedToAccess, err := IsUserAllowedToAccessTask(token, task.GroupID)
	if err != nil {
		return nil, err
	}
	if !isUserAllowedToAccess {
		err := errors.New("User not allowed to view Group details")
		return nil, err
	}

	subtaskRows, err := db.Query("SELECT id,Points,Tasks_id,Name FROM `Subtasks` WHERE Tasks_id = ?", taskID)
	if err != nil {
		log.Println("GetSubtasksForTasks: " + err.Error())
		return nil, err
	}
	defer subtaskRows.Close()
	subtasks := make([]datastructures.Subtask, 0, 10)
	for subtaskRows.Next() {
		var subtask datastructures.Subtask
		err := subtaskRows.Scan(&subtask.ID, &subtask.Points, &subtask.TaskID, &subtask.Name)
		if err != nil {
			log.Println("GetSubtasksForTasks: " + err.Error())
			return nil, err
		}
		subtasks = append(subtasks, subtask)
	}
	return subtasks, nil
}

//AddTask adds a task to a group and returns the fully populated task and an error
func AddTask(task datastructures.Task) (datastructures.Task, error) {
	res, err := db.Exec("INSERT INTO Tasks(Name,Author,Description,Group_id) VALUES (?,?,?,?)",
		task.Name, task.Author, task.Description, task.GroupID)
	if err != nil {
		log.Printf("AddTask:" + err.Error())
		return task, err
	}
	id, err := res.LastInsertId()
	task.ID = int(id)
	if err != nil {
		log.Printf("AddTask:" + err.Error())
		return task, err
	}
	return task, nil

}
