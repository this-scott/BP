package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var serveVersion int = 1

// Serve Content
func basicHandler(w http.ResponseWriter, r *http.Request) {
	// ../ is going to have to be reorganzied
	t, err := template.ParseFiles("static/serve" + string(rune(serveVersion)) + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
	fmt.Println("Started basic server")
}

// Accept new content
func contentHandler(w http.ResponseWriter, r *http.Request) {
	// validating this is going to suck
	// creating a webhook so nothing is sent back
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// validating
	if r.Header.Get("Key") == os.Getenv("authkey") {
		//write new file
	} else {
		err := fmt.Errorf("bad key in content handler: %s", r.Header.Get("Key"))
		fmt.Println(err.Error())
	}

}

// ship to log server
func sendToLog() {

}

func main() {
	fmt.Println("Hello World")
	err := godotenv.Load(".env")
	//Something about this line brings me joy
	if err != nil {
		panic(err.Error())
	}

	//need to handle specified pages but in that case I need an actual landing page
	//scoop.com/livelylanding
	http.HandleFunc("/", basicHandler)
	http.HandleFunc("/update", contentHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
