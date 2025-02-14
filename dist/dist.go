package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var serveVersion int = 1

// Serve Content
func basicHandler(w http.ResponseWriter, r *http.Request) {
	// ../ is going to have to be reorganzied
	t, err := template.ParseFiles("static/serve.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
	fmt.Println("Started basic server")
}

// Accept new content
func contentHandler(w http.ResponseWriter, r *http.Request) {
	// A webhook is just an api which sends nothing back
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// validating. There is definitely a better way to validate this
	if r.Header.Get("Key") == os.Getenv("authkey") {
		err := r.ParseMultipartForm(10 << 20) //allocates 10mb file size
		if err != nil {
			//using this line because we're not returning anything to content server
			fmt.Println(err.Error())
			return
		}

		//write new file(s)
		for fname := range r.MultipartForm.File {
			file, header, err := r.FormFile(fname)
			if err != nil {
				fmt.Println(err)
				//fmt.Errorf("Error retrieving file %d", http.StatusBadRequest)
				return
			}
			defer file.Close()

			outFile, err := os.Create("./static/" + header.Filename)
			if err != nil {
				fmt.Println(err.Error())
				//fmt.Errorf("Error saving file %d", http.StatusInternalServerError)
				return
			}
			defer outFile.Close()

			io.Copy(outFile, file)
			fmt.Printf("Uploaded: %s\n", header.Filename)
		}
	} else {
		err := fmt.Errorf("bad key in content handler: %s", r.Header.Get("Key"))
		fmt.Println(err.Error())
	}

}

// ship to log server
func sendToLog() {

}

// func init() {
// 	//check to see if distribution content is updated
// }

func main() {
	//EVERYTHING BEFORE TODO SHOULD BE WRAPPED IN THE INIT FUNCTION BUT GO'S SCOPE DOESN'T WORK LIKE THAT

	//TODO: SETUP FOR HTTPS
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
