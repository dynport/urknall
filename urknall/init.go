package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"path/filepath"
	"strings"

	"github.com/dynport/dgtk/github"
)

type initProject struct {
	Dir string `cli:"arg required"`
}

func loadBase() (*base, error) {
	wd, e := os.Getwd()
	if e != nil {
		return nil, e
		logger.Fatal(e)
	}
	return &base{BaseDir: wd}, nil
}

func githubClient() *http.Client {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return github.NewHttpClient(token)
	}
	return &http.Client{}
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

func loadContent(cl *http.Client, url string, i interface{}) error {
	logger.Printf("loading content from %q", url)
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

func exampleFiles() ([]*content, error) {
	cl := githubClient()
	rsp, e := cl.Get("https://api.github.com/repos/dynport/urknall/contents/examples")
	defer rsp.Body.Close()
	if e != nil {
		return nil, e
	}
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("expected status 2xx, got %q", rsp.Status)
	}
	decoder := json.NewDecoder(rsp.Body)

	contents := []*content{}
	e = decoder.Decode(&contents)
	if e != nil {
		return nil, e
	}
	return contents, e
}

func (init *initProject) Run() error {
	dir, e := filepath.Abs(init.Dir)
	if e != nil {
		return e
	}

	_, e = os.Stat(dir)
	switch {
	case os.IsNotExist(e):
		if e = os.Mkdir(dir, 0755); e != nil {
			return e
		}
	case e != nil:
		return e
	}

	contents, e := exampleFiles()
	if e != nil {
		return e
	}

	for _, c := range contents {
		localPath := dir + "/" + c.Name
		switch {
		case strings.HasPrefix(c.Name, "cmd_") || c.Name == "main.go":
			_, e := os.Stat(localPath)
			if e == nil {
				logger.Printf("file %q exists", c.Name)
				continue
			}

			e = c.Load()
			if e != nil {
				return e
			}

			content, e := c.DecodedContent()
			if e != nil {
				return e
			}
			e = writeFile(localPath, content)
			if e != nil {
				return e
			}
			logger.Printf("saving file %q to %q", c.Name, localPath)
		}
	}
	return nil
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

var b64 = base64.StdEncoding
