package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
)

var redisClient redis.Client

func main() {

	redisClient = *connectToRedis()

	router := httprouter.New()
	//TODO: router.GET("/:link/info", displayInfo)
	router.POST("/", addRedirect)
	router.GET("/", indexPage)
	router.GET("/:link", redirect)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func generateRandomLink() string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	b := make([]byte, 4)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)

}

func indexPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//TODO: add index page to create shorturl
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Index page"))
}

func redirect(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	target, err := redisClient.Get(ps.ByName("link")).Result()

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

func addRedirect(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	generated := false
	shortLink := r.FormValue("shortlink")
	website := r.FormValue("website")

	dialAdress := strings.TrimPrefix(website, "http://")
	dialAdress = strings.TrimPrefix(dialAdress, "https://")
	dialAdress += ":http"

	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", dialAdress, timeout)

	if err != nil {
		//TODO: add invalid website page
		fmt.Println(dialAdress)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Status code: 400 Invalid website, not created"))
		return
	}

	if !strings.HasPrefix("https://", website) || !strings.HasPrefix("http://", website) {
		website = "http://" + website
	}

	if shortLink == "" {
		shortLink = generateRandomLink()
		generated = true
	}

	ex, err := redisClient.SetNX(shortLink, website, 0).Result()

	if err != nil {
		panic(err)
	}
	if ex == false && generated == false {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Status code: 409 " + shortLink + " already exists in database, choose another shortlink"))
	} else if ex == false && generated == true {
		for ex != true { // generate random link while it won't be available it database
			shortLink = generateRandomLink()
			ex, err = redisClient.SetNX(shortLink, website, 0).Result()
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Status code: 201 " + website + " added succesfully as " + shortLink))

	} else {
		//TODO: add created page
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Status code: 201 " + website + " added succesfully as " + shortLink))
	}

}

func connectToRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(pong)

	return client
}
