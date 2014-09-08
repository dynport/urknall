package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dynport/dgtk/github"
)

var b64 = base64.StdEncoding

func githubClient() *http.Client {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return github.NewHttpClient(token)
	}
	return &http.Client{}
}

func loadContent(cl *http.Client, url string, i interface{}) error {
	rsp, e := cl.Get(url)
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	return json.Unmarshal(b, &i)
}

type content struct {
	Name     string `json:"name"`     // "Makefile",
	Path     string `json:"path"`     // "github/Makefile",
	Sha      string `json:"sha"`      // "1a91023ae6c2b830090d615304098ac957453ae2",
	Size     int    `json:"size"`     // 21,
	Url      string `json:"url"`      // "https://api.github.com/repos/dynport/dgtk/contents/github/Makefile?ref=master",
	HtmlUrl  string `json:"html_url"` // "https://github.com/dynport/dgtk/blob/master/github/Makefile",
	GitUrl   string `json:"git_url"`  // "https://api.github.com/repos/dynport/dgtk/git/blobs/1a91023ae6c2b830090d615304098ac957453ae2",
	Type     string `json:"type"`     // "file",
	Content  string `json:"content"`
	Encoding string `json:"encoding"` // "base64",
}

func (c *content) DecodedContent() ([]byte, error) {
	return b64.DecodeString(c.Content)
}

func (c *content) Load() error {
	return loadContent(githubClient(), c.Url, c)
}

func writeFile(p string, content []byte) error {
	f, e := os.OpenFile(p, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0644)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write(content)
	return e
}
