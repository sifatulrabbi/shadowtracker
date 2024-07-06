package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/*", handleReq)
	if err := http.ListenAndServe("0.0.0.0:3000", mux); err != nil {
		fmt.Println(err)
	}
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	fmt.Println("new req received")
	fmt.Fprintf(w, "Hello World!\n")
}
