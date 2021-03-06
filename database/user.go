package database

import (
	"errors"
	"log"
)

func isUserInGroup(token string, groupID int) (bool, error) {
	var isInGroup bool
	err := db.QueryRow("SELECT exists (SELECT * FROM Group_has_Users WHERE User_Token = ? AND Group_id = ?)",
		token, groupID).Scan(&isInGroup)
	return isInGroup, err
}

//IsUserAdminOfGroup checks if a user with a given token is a admin of the given group
func IsUserAdminOfGroup(token string, groupID int) (bool, error) {
	var isAdmin int = 0
	err := db.QueryRow("SELECT IsAdmin FROM Group_has_Users WHERE User_Token = ? AND Group_id = ?",
		token, groupID).Scan(&isAdmin)
	if isAdmin == 1 {
		return true, err
	}
	return false, err
}

func doesUserExists(token string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT exists (SELECT * FROM User WHERE Token = ?)",
		token).Scan(&exists)
	return exists, err
}

//AddUser adds a usertoken if no other user with this token exists
func AddUser(token string) error {
	ex, err := doesUserExists(token)
	if err != nil {
		log.Println("AddUser: " + err.Error())
		return err
	}
	if ex {
		return errors.New("User already exists")
	}
	_, err = db.Exec("INSERT INTO User(Token) VALUES (?)", token)
	if err != nil {
		log.Println("AddUser: " + err.Error())
		return err
	}
	return nil
}

func AddUserToGroup(token string, groupID int) error {
	groupExists, err := DoesGroupExists(groupID)
	if err != nil {
		log.Println("AddUser: " + err.Error())
		return err
	}
	if !groupExists {
		return errors.New("Group does not exists")
	}
	userInGroup, err := isUserInGroup(token, groupID)
	if err != nil {
		log.Println("AddUser: " + err.Error())
		return err
	}
	if userInGroup {
		return errors.New("User already in group")
	}
	_, err = db.Exec("INSERT INTO Group_has_Users(Group_id,User_Token,IsAdmin) VALUES (?,?,?)",
		groupID, token, 0)

	if err != nil {
		log.Println("AddUser: " + err.Error())
		return err
	}
	return nil
}
