package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Get uploaded file
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Save uploaded file to ./uploads/
		os.MkdirAll("uploads", 0755)
		savePath := "uploads/" + handler.Filename
		dst, err := os.Create(savePath)
		if err != nil {
			http.Error(w, "Could not save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)

		//  Vulnerability: Execute .sh files automatically (very dangerous)
		if len(handler.Filename) > 3 && handler.Filename[len(handler.Filename)-3:] == ".sh" {
			fmt.Println("[!] Executing uploaded shell script...")
			cmd := exec.Command("bash", savePath)
			cmd.Run()
		}

		// Respond to user
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<p>File uploaded successfully: <b>%s</b></p>", handler.Filename)
		return
	}

	// Display upload form
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>Upload Portal</title></head>
		<body style="font-family: sans-serif; padding: 40px;">
			<h2>Upload your profile image</h2>
			<form method="POST" enctype="multipart/form-data">
				<input type="file" name="file" /><br><br>
				<input type="submit" value="Upload" />
			</form>
		</body>
		</html>
	`))
}

func main() {
	fmt.Println("[*] Server listening on http://localhost:8080/upload")
	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":8080", nil)
}
