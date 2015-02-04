package main

import (
	"fmt"
	"github.com/mattn/platini"
	"net/http"
	"strings"
)

type Pet struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type PetReq struct {
	Id int `json:"id"`
}

func list() ([]Pet, error) {
	return []Pet{{1, "Linda_pp", "犬"}, {2, "supermomonga", "ももんが"}}, nil
}

func get(req *PetReq) (*Pet, error) {
	switch req.Id {
	case 1:
		return &Pet{1, "Linda_pp", "犬"}, nil
	case 2:
		return &Pet{2, "supermomonga", "モモンガ"}, nil
	default:
		return nil, fmt.Errorf("user not found: id=%d", req.Id)
	}
}

func add(w http.ResponseWriter, r *http.Request, req *Pet) (*Pet, error) {
	if strings.TrimSpace(req.Kind) == "" {
		return nil, fmt.Errorf("invalid request: kind is require")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("invalid request: name is require")
	}
	return &Pet{1, "Linda_pp", "犬"}, nil
}

func main() {
	platini.Get("/users/:Id", get)
	platini.Get("/users/", list)
	platini.Post("/users/", add)
	platini.Handle("/", http.FileServer(http.Dir("static")))
	http.ListenAndServe(":8080", platini.DefaultMux)
}
