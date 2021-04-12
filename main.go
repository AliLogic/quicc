package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var urlMap = make(map[string]string)
var quiccFile string

func main() {

	// zeroth, make sure that the file name is specified in the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	// check if file exists or not
	if quiccFile = os.Getenv("QUICC_FILE"); len(quiccFile) == 0 {
		log.Fatal("Environment variable QUICC_FILE is empty!")
	}

	_, err = os.Stat(quiccFile)
	if os.IsNotExist(err) {
		log.Panic(err)

		var file *os.File
		var data []byte

		if file, err = os.Create(quiccFile); err != nil {

			log.Fatal(err)
		}
		defer file.Close()

		data, err = json.Marshal("_")
		file.WriteString(string(data))
	}

	// if it does, then continue to read the file
	var data []byte
	if data, err = os.ReadFile(quiccFile); err != nil {
		// file does exist but something is wrong, don't continue
		// fmt.Println("something went wrong while reading the file")
		log.Fatal(err)
	}

	if err = json.Unmarshal(data, &urlMap); err != nil {
		// failed to parse json to map, don't continue
		log.Fatal(err)
	}

	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/add/", additionHandler)
	http.HandleFunc("/delete/", deletionHandler)

	// it's important that we check our http server is alive or not
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func deletionHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST method
	if r.Method != "POST" {
		return
	}

	// failed to parse form? panic!
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	key := r.FormValue("key")

	// key validation
	if _, exists := getKey(key); !exists {
		fmt.Fprintf(w, "Key '%s' is not registered!", key)
		return
	}

	// delete key
	delete(urlMap, key)
	fmt.Fprintf(w, "Key '%s' is deleted!", key)
	saveLinks()
}

func additionHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST method
	if r.Method != "POST" {
		return
	}

	// failed to parse form? panic1
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	key := r.FormValue("key")
	link := r.FormValue("link")

	// key validation
	if _, exists := getKey(key); exists {
		fmt.Fprintf(w, "Key '%s' already exists!", key)
		return
	}

	// key registered!
	urlMap[key] = link
	fmt.Fprintf(w, "Key '%s' is now registered to redirect to '%s'", key, link)
	saveLinks()
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// User didn't ask for a redirect? Serve them the homepage
	key := r.URL.Path[1:]
	if len(key) == 0 {
		http.ServeFile(w, r, "index.html")
		return
	}

	// key validation before we redirect
	if link, exists := getKey(key); exists {
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	fmt.Fprintf(w, "'%s' is not a valid redirect key!", key)
}

func saveLinks() {
	var file *os.File
	var data []byte
	var err error

	if data, err = json.Marshal(urlMap); err != nil {
		log.Fatal(err)
	}

	if err = ioutil.WriteFile(quiccFile, data, 0); err != nil {
		log.Fatal(err)
	}
	defer file.Close()
}

func getKey(key string) (string, bool) {
	link, exists := urlMap[key]
	return link, exists
}
