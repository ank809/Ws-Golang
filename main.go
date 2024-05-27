package main

import (
	"fmt"
	"net/http"
)

func main() {
	go broadCastMessage()
	http.HandleFunc("/ws", webserver)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Server is running on port 8081")
	}
}
