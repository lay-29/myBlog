package services

import (
	"MyBlog/models"
	"encoding/json"
	"fmt"
	"net/http"
)

type ServiceManager struct {
}

func NewServiceManager() (*ServiceManager, error) {
	return &ServiceManager{}, nil
}

func (sm *ServiceManager) Start() {
	mux := sm.SetRouter()
	http.ListenAndServe(":8080", mux)

}

func (sm *ServiceManager) SetRouter() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.HandleFunc("/user/", UserEntryHandler)

	return mux
}
func writeJson(w http.ResponseWriter, resp models.ApiResp) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Println(err)
	}
}
