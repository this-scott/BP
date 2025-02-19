package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var serveVersion int = 0

func latestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//I don't have to worry about writing back because my crazy format on dist
	//I do need to worry about sending another request tho
	sendLatest(serveVersion)
}

// having this be called latest and say serveVersion doesn't get confusing at all
func sendLatest(latest int) error {
	//this was so much nicer on Postman
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	req, err := http.NewRequest(http.MethodPut, os.Getenv("distserver")+"/update", &b)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Key", os.Getenv("authkey"))
	//multipart form content
	req.Header.Set("Content-Type", writer.FormDataContentType())

	directory := "versions/serve" + strconv.Itoa(latest)

	// Is this memory inefficient: yes. I will fix it later
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	// Add each file to the multipart form data
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(directory, file.Name())
			fileContent, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer fileContent.Close()

			// Create a form file part for each file
			part, err := writer.CreateFormFile("files", file.Name())
			if err != nil {
				return err
			}

			// Copy the file content into the form file part
			_, err = io.Copy(part, fileContent)
			if err != nil {
				return err
			}
		}
	}

	// Closing the writer
	writer.Close()

	// this part's actually kind of simple lol
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func getLatestVersionNum(directory string, folder string) (int, error) {
	highestNum := -1
	// Getting all files that fit
	pattern := regexp.MustCompile(folder + `(\d+)`)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skipping subdirs
		if info.IsDir() && path != directory {
			matches := pattern.FindStringSubmatch(info.Name())
			if len(matches) > 1 {
				num, err := strconv.Atoi(matches[1])
				if err != nil {
					return nil
				}

				if num > highestNum {
					highestNum = num
				}
			}
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return -1, fmt.Errorf("error walking directory: %v", err)
	}

	//one more check
	if highestNum == -1 {
		return -1, fmt.Errorf("no folders matching pattern '%s<number>' found", folder)
	}

	return highestNum, nil
}

func main() {
	fmt.Println("Versioning is Go")
	// good old file serving

	latest, _ := getLatestVersionNum("versions", "serve")
	fmt.Println(latest)
	serveVersion = latest
	http.HandleFunc("/latest", latestHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
