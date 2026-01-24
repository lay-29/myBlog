package services

import (
	"MyBlog/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func UserEntryHandler(w http.ResponseWriter, r *http.Request) {
	subPath := r.URL.Path[len("/user/"):]
	fmt.Println("URL: ", r.URL)
	fmt.Println("SubPath: ", subPath)
	switch subPath {
	case "login":
		UserLoginHandleFunc(w, r)
	case "logout":
		UserLogoutHandleFunc(w, r)
	case "register":
		UserRegisterHandleFunc(w, r)
	}
}

func UserRegisterHandleFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Func: ", "UserRegisterHandleFunc", "Meth：", r.Method)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Println("UserLoginHandleFunc: ", "StatusMethodNotAllowed")
		return
	}
	//user1 := models.User{Name: "Louis", Email: "louis@gmail.com", Password: "123456"}
	//userJson, _ := json.Marshal(user1)
	//fmt.Println(string(userJson))

	//bodystr, _ := io.ReadAll(r.Body)
	//fmt.Println(string(bodystr))

	user := models.User{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		fmt.Println("UserRegisterHandleFunc: ", "StatusBadRequest")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("User:\n\tName:", user.Name, "\n\tEmail:", user.Email)

	if err := user.RegisterUser(); err != nil {
		fmt.Println("UserRegisterHandleFunc: ", "StatusInternalServerError", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJson(w, models.ApiResp{Code: 0, Msg: "ok", Data: user})
}

func UserLogoutHandleFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Func: ", "UserLogoutHandleFunc")
}

func UserLoginHandleFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Func: ", "UserLoginHandleFunc")

}
