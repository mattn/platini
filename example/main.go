package main

import (
	"fmt"
	"github.com/mattn/platini"
	"net/http"
	"strings"
)

type Pet struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type PetReq struct {
	Id int64 `json:"id"`
}

var pets = []Pet{{1, "Linda_pp", "犬"}, {2, "supermomonga", "ももんが"}}

func list() ([]Pet, error) {
	return pets, nil
}

func get(req *PetReq) (*Pet, error) {
	for _, pet := range pets {
		if req.Id == pet.Id {
			return &pet, nil
		}
	}
	return nil, fmt.Errorf("user not found: id=%d", req.Id)
}

func add(w http.ResponseWriter, r *http.Request, req *Pet) (*Pet, error) {
	if strings.TrimSpace(req.Kind) == "" {
		return nil, fmt.Errorf("invalid request: kind is require")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("invalid request: name is require")
	}
	req.Id = int64(len(pets)) + 1
	pets = append(pets, *req)
	return req, nil
}

func main() {
	platini.Get("/users/:Id", get)
	platini.Get("/users/", list)
	platini.Post("/users/", add)
	platini.Handle("/", http.FileServer(http.Dir("static")))
	http.ListenAndServe(":8080", platini.DefaultMux)
}
