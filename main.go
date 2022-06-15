package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

var (
	fileNameKey  = "fileName"
	operationKey = "operation"
	itemKey      = "item"
	idKey        = "id"
)

func Perform(args Arguments, writer io.Writer) error {
	if err := checkOperationValid(args[operationKey]); err != nil {
		return err
	}
	if err := checkFileNameValid(args[fileNameKey]); err != nil {
		return err
	}

	operation := args[operationKey]
	switch operation {
	case "list":
		return executeList(args, writer)
	case "add":
		return executeAdd(args, writer)
	case "remove":
		return executeRemove(args, writer)
	case "findById":
		return executeFindById(args, writer)
	default:
		return fmt.Errorf("Operation %s not allowed!", operation)
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func parseArgs() Arguments {
	fileNameArg := flag.String("fileName", "", "file name to store user info")
	operationArg := flag.String("operation", "", "operation to do")
	itemArg := flag.String("item", "", "user item to add")
	idArg := flag.String("id", "", "user id to search")
	flag.Parse()

	fmt.Println("arguments:")
	fmt.Println("  fileName", *fileNameArg)
	fmt.Println("  operation", *operationArg)
	fmt.Println("  item", *itemArg)
	fmt.Println("  id", *idArg)
	fmt.Println()

	return Arguments{
		fileNameKey:  sanitizeArg(*fileNameArg),
		operationKey: sanitizeArg(*operationArg),
		itemKey:      *itemArg,
		idKey:        sanitizeArg(*idArg),
	}
}

func sanitizeArg(arg string) string {
	result := strings.ReplaceAll(arg, "«", "")
	result = strings.ReplaceAll(result, "»", "")
	return result
}

func checkOperationValid(operation string) error {
	switch operation {
	case "list", "add", "remove", "findById":
		return nil
	case "":
		return errors.New("-operation flag has to be specified")
	default:
		return fmt.Errorf("Operation %s not allowed!", operation)
	}
}

func checkFileNameValid(fileName string) error {
	if len(fileName) == 0 {
		return errors.New("-fileName flag has to be specified")
	}

	return nil
}

func checkItemValid(item string) error {
	if len(item) == 0 {
		return errors.New("-item flag has to be specified")
	}

	return nil
}

func checkIdValid(id string) error {
	if len(id) == 0 {
		return errors.New("-id flag has to be specified")
	}

	return nil
}

func executeAdd(args Arguments, writer io.Writer) error {
	item := args[itemKey]
	err := checkItemValid(item)
	if err != nil {
		return err
	}

	var user User
	err = json.Unmarshal([]byte(item), &user)
	if err != nil {
		return err
	}

	users := readUsers(args)
	for _, u := range users {
		if u.Id == user.Id {
			writeData([]byte(fmt.Sprintf("Item with id %s already exists", user.Id)), writer)
			return nil
		}
	}

	users = append(users, user)

	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	writeDataToFile(data, args[fileNameKey])

	return nil
}

func executeList(args Arguments, writer io.Writer) error {
	data := readFile(args)
	writeData(data, writer)
	return nil
}

func executeFindById(args Arguments, writer io.Writer) error {
	id := args[idKey]
	if err := checkIdValid(id); err != nil {
		return err
	}

	users := readUsers(args)
	for _, user := range users {
		if user.Id == id {
			data, err := json.Marshal(user)
			if err != nil {
				return err
			}
			writeData(data, writer)
			return nil
		}
	}
	writeData([]byte(""), writer)
	return nil
}

func executeRemove(args Arguments, writer io.Writer) error {
	id := args[idKey]
	err := checkIdValid(id)
	if err != nil {
		return err
	}
	users := readUsers(args)
	filteredUsers := make([]User, 0)
	for _, user := range users {
		if user.Id != id {
			filteredUsers = append(filteredUsers, user)
		}
	}
	if len(filteredUsers) == len(users) {
		writeData([]byte(fmt.Sprintf("Item with id %s not found", id)), writer)
		return nil
	}

	data, err := json.Marshal(filteredUsers)
	if err != nil {
		return err
	}

	writeDataToFile(data, args[fileNameKey])

	return nil
}

func readFile(args Arguments) []byte {
	fileName := args[fileNameKey]
	data, _ := os.ReadFile(fileName)
	return data
}

func readUsers(args Arguments) []User {
	data := readFile(args)
	if len(data) == 0 {
		return []User{}
	}
	var users []User
	err := json.Unmarshal(data, &users)
	if err != nil {
		panic(err)
	}
	return users
}

func writeData(data []byte, writer io.Writer) {
	_, err := writer.Write(data)
	if err != nil && err != io.EOF {
		panic(err)
	}
}

func writeDataToFile(data []byte, fileName string) {
	err := os.WriteFile(fileName, data, 0644)
	if err != nil {
		panic(err)
	}
}
