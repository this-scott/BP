package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var serveVersion int = 0

// Serve Content
func basicHandler(w http.ResponseWriter, r *http.Request) {
	// ../ is going to have to be reorganzied
	t, err := template.ParseFiles("/static/serve.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
	//fmt.Println("Started basic server")
}

// Accept new content
// ngl if the content server(literally exists just to serve) is just going to be my laptop I should store old versions in the log server
func contentHandler(w http.ResponseWriter, r *http.Request) {
	// A webhook is just an api which sends nothing back
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// validating. There is definitely a better way to validate this
	if r.Header.Get("Key") == os.Getenv("authkey") {

		//R storing the files within itself as r.MultipartForm instead of returning as an array is evil and akward
		err := r.ParseMultipartForm(10 << 20) //allocates 10mb file size
		if err != nil {
			//using this line because we're not returning anything to content server
			fmt.Println(err.Error())
			return
		}
		//checking version. Perfectly ok with rollbacks
		versionValue, exists := r.Form["sversion"]
		if !exists || len(versionValue) == 0 {
			fmt.Println("Version not found")
			return
		} else {
			//serveVersion here isn't the global var above
			serveVersion, err := strconv.Atoi(versionValue[0])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("Changing to Version:", serveVersion)
		}

		fmt.Println(r.MultipartForm.File)
		//write new file(s)
		for f, headers := range r.MultipartForm.File {
			fmt.Println(f)
			for _, header := range headers {
				file, err := header.Open() // Open each file
				if err != nil {
					fmt.Println("Error opening file:", err)
					continue
				}
				defer file.Close()

				//create a blank copy with the same name as the file
				outFile, err := os.Create("/static/" + header.Filename)
				if err != nil {
					fmt.Println("Error creating file:", err.Error())
					//fmt.Errorf("Error saving file %d", http.StatusInternalServerError)
					return
				}
				defer outFile.Close()

				//copy file contents in
				io.Copy(outFile, file)
				fmt.Printf("Uploaded: %s\n", outFile.Name())
			}
		}
	} else {
		err := fmt.Errorf("bad key in content handler: %s", r.Header.Get("Key"))
		fmt.Println(err.Error())
	}

}

func init() {
	//LMAO I DIDN'T REALIZE INIT IS CALLED AUTOMATICALLY I WAS GOING TO THROW IT IN FRONT OF MAIN

	staticDir := "/static"
	// Ensure /static exists
	err := os.MkdirAll(staticDir, 0755)
	if err != nil {
		fmt.Println("Error creating /static directory:", err)
		return
	}
	//COMMENTING THIS OUT FOR DOCKER
	// err := godotenv.Load(".env")
	// //Something about this line brings me joy
	// if err != nil {
	// 	panic(err.Error())
	// }

	//the internet says I should send my integers raw ;P
	//body := bytes.NewBufferString(strconv.Itoa(serveVersion))

	//sends request to content server to check for updates -> updates will be sent to update url
	//this is going to be a hook to because I don't see a point writing the above content handler unless I want to wrap the change content function(lowkey probably should)
	//WAIT I THOUGHT ABOUT THIS WITH FRESH EYES AND THIS IS NOT GOOD OML
	requestURL := fmt.Sprintf("http://%s:%d/latest", os.Getenv("version_server"), 8080)

	startTime := time.Now()
	timeout := 10 * time.Second

	for {
		// Wrapping checking with log server into a 30 second connection phase
		if time.Since(startTime) > timeout {
			fmt.Println("client: request failed after 10 seconds of retries")
			break
		}

		//would send body but apparently its weird to with get requests.
		// Also we're not saving content here on docker up so we boot with nothing
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		if err != nil {
			fmt.Printf("client: could not create request: %s\n", err)
			os.Exit(1)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			if os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused") {
				fmt.Println("versioning server connection refused: retrying in 2 seconds")
				time.Sleep(2 * time.Second)
				continue
			} else {
				fmt.Printf("client: error making http request: %s\n", err)
			}
		}
		defer res.Body.Close()
	}
}

func main() {
	//TODO: SETUP FOR HTTPS
	fmt.Println("Dist is Go")
	//need to handle specified pages but in that case I need an actual landing page
	//scoop.com/livelylanding
	http.HandleFunc("/", basicHandler)
	http.HandleFunc("/update", contentHandler)

	// Serve static files from /static
	fs := http.FileServer(http.Dir("/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
