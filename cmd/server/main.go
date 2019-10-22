package main

import (
	"log"
	"net/http"

	"url-shortener/redirect"

	"github.com/julienschmidt/httprouter"
)

func main() {

	host := "localhost"
	port := "6379"
	password := ""
	db := 0

	redirecter := redirect.NewRedirecter(host+":"+port, password, db)

	router := httprouter.New()
	router.POST("/", redirecter.AddRedirect)
	router.GET("/:link", redirecter.Redirect)

	log.Fatal(http.ListenAndServe(":8080", router))
}
