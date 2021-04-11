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

var mapVar = make(map[string]string)
var quiccFile string

func main() {

	// zeroth, make sure that the file name is specified in the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	// check if file exist or not first
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
	data, err := os.ReadFile(quiccFile)
	if err != nil {
		// file does exist but something is wrong, don't continue
		// fmt.Println("something went wrong while reading the file")
		log.Fatal(err)
	}

	if err = json.Unmarshal(data, &mapVar); err != nil {
		// failed to parse json to map, don't continue
		log.Fatal(err)
	}

	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/add/", additionHandler)

	// it's important that we check our http server is alive or not
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func additionHandler(w http.ResponseWriter, r *http.Request) {
	// here, we get all the post method stuff
	// let's process UwU

	// guard clauses (early return) go brrr
	if r.Method != "POST" {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	key := r.FormValue("key")
	link := r.FormValue("link")

	_, ok := mapVar[key]
	if ok {
		fmt.Fprintf(w, "Key '%s' already exists!", key)
	} else {
		mapVar[key] = link
		fmt.Fprintf(w, "Key '%s' is now registered to redirect to '%s'", key, link)
		saveLinks()
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// okay, so we have received some path in our link,
	// let's check if it is valid or not (by checking in our map)

	key := r.URL.Path[1:]
	if len(key) == 0 {
		http.ServeFile(w, r, "index.html")
		return
	}

	link, ok := mapVar[key]
	if ok {
		http.Redirect(w, r, link, http.StatusFound)
	} else {
		fmt.Fprintf(w, "'%s' is not a valid redirect key!", key)
	}
}

func saveLinks() {
	var file *os.File
	var data []byte
	var err error

	if data, err = json.Marshal(mapVar); err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(quiccFile, data, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
}
