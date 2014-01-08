package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func handleError(w http.ResponseWriter, msg string, statusCode int) {
	log.Print(msg)
	http.Error(w, msg, statusCode)
}

// add binary package, name, version, arch
func addPackage(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		handleError(w, "only post allowed", http.StatusMethodNotAllowed)
		return
	}

	e := req.ParseMultipartForm(2048)
	if e != nil {
		handleError(w, "failed parse request", http.StatusInternalServerError)
		return
	}

	form := req.MultipartForm
	if form == nil {
		handleError(w, "request contains no uploaded files", http.StatusInternalServerError)
		return
	}
	defer form.RemoveAll()

	osFile, err := os.Create(fmt.Sprintf("/tmp/packages/%s", req.FormValue("file")))
	if err != nil {
		handleError(w, "failed to open output file", http.StatusInternalServerError)
		return
	}
	defer osFile.Close()

	formFile, _, err := req.FormFile("data")
	if err != nil {
		handleError(w, "failed to get request form file", http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	_, e = io.Copy(osFile, formFile)
	if e != nil {
		handleError(w, "failed to write file", http.StatusInternalServerError)
		return
	}
}

// get binary package (by name, version and arch).
func getPackage(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		handleError(w, "only get allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(req.URL.Path[1:], "/")

	if len(parts) != 5 {
		log.Print(parts)
		handleError(w, "wrong path given /get/<pkg>/<version>/<arch>/[data|checksum]", http.StatusBadRequest)
		return
	}

	pkg, version, arch, action := parts[1], parts[2], parts[3], parts[4]

	filename := fmt.Sprintf("%s.%s.%s.bpkg", pkg, version, arch)
	if _, err := os.Stat("/tmp/packages/" + filename); os.IsNotExist(err) {
		handleError(w, fmt.Sprintf("package %s not known", filename), http.StatusNotFound)
		return
	}

	fh, err := os.Open("/tmp/packages/" + filename)
	if err != nil {
		handleError(w, fmt.Sprintf("failed to read package %s", filename), http.StatusInternalServerError)
		return
	}
	defer fh.Close()

	w.Header().Set("Content-Type", "application/octet-stream")

	switch action {
	case "data":
		_, err = io.Copy(w, fh)
		if err != nil {
			handleError(w, fmt.Sprintf("failed to send package %s", filename), http.StatusInternalServerError)
			return
		}

	case "checksum":
		hash := sha256.New()
		_, err := io.Copy(hash, fh)
		if err != nil {
			handleError(w, fmt.Sprintf("failed to checksum package %s", filename), http.StatusInternalServerError)
			return
		}
		h := make([]byte, 256)
		hex.Encode(h, hash.Sum(nil))
		for n := 0; n < len(h); {
			x, err := w.Write(h[n:])
			if err != nil {
				handleError(w, "", http.StatusInternalServerError)
				return
			}
			n += x
		}

	default:
		handleError(w, "wrong action given", http.StatusBadRequest)
	}
}

func main() {
	http.HandleFunc("/add", addPackage)
	http.HandleFunc("/get/", getPackage)

	http.ListenAndServe(":8080", nil)
}
