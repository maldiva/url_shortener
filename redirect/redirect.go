package redirect

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
)

type Redirecter struct {
	redis.Client
}

func NewRedirecter(address, password string, database int) Redirecter {

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       database,
	})

	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	return Redirecter{*client}
}

func (redir *Redirecter) Redirect(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	target, err := redir.Get(ps.ByName("link")).Result()
	if err != nil && err != redis.Nil {
		panic(err)
	}

	fmt.Println(target)

	if err == redis.Nil {
		// TODO: add page for not found status
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Redirect link not found"))
	} else {
		http.Redirect(w, r, target, 301)
	}
}

func (redir *Redirecter) AddRedirect(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	generated := false
	shortLink := r.FormValue("shortlink")
	website := r.FormValue("website")

	website = strings.TrimPrefix(website, "http://")
	website = strings.TrimPrefix(website, "https://")
	website += ":http"

	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", website, timeout)
	if err != nil {
		//TODO: add invalid website page
		fmt.Println(website)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Status code: 400 Invalid website, not created"))
		return
	}

	if shortLink == "" {
		shortLink = generateRandomLink()
		generated = true
	}

	ex, err := redir.SetNX(shortLink, website, 0).Result()
	if err != nil {
		panic(err)
	}

	if ex == false && generated == false {

		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Status code: 409 " + shortLink + " already exists in database, choose another shortlink"))

	} else if ex == false && generated == true {

		for ex != true { // generate random link while it won't be available it database
			shortLink = generateRandomLink()
			ex, err = redir.SetNX(shortLink, website, 0).Result()
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Status code: 201 " + website + " added succesfully as " + shortLink))

	} else {

		//TODO: add created page
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Status code: 201 " + website + " added succesfully as " + shortLink))

	}

}

func generateRandomLink() string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	b := make([]byte, 4)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)

}
