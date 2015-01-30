package main

import (
	"github.com/mattn/platini"
	"net/http"
)

type Pet struct {
	Id   int
	Name string
	Kind string
}

type PetReq struct {
	Id int
}

func parsedHandler1(req *PetReq) (*Pet, error) {
	//return db.Get(req.Id)
	println("Id:", req.Id)
	return &Pet{1, "Linda_pp", "犬"}, nil
}

func parsedHandler2(w http.ResponseWriter, r *http.Request, req *Pet) (*Pet, error) {
	// w, r はCookie読み書きするのに必要な時もある…(闇
	//return db.Put(req)
	return &Pet{1, "Linda_pp", "犬"}, nil
}

func main() {
	platini.HandleFunc("GET", "/:Id", parsedHandler1) // いい感じにパースしてくれる
	platini.HandleFunc("POST", "/", parsedHandler2)   // ハンドラを登録してウェブページを表示させる
	http.ListenAndServe(":8080", nil)
}
