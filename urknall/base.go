package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type templates map[string]*content

func (t templates) exists(name string) bool {
	_, exist := t[name]
	return exist
}

func (t templates) names() []string {
	names := []string{}
	for n := range t {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func upstreamFiles(repo, path string) ([]*content, error) {
	cl := githubClient()
	url := "https://api.github.com/repos/" + repo + "/contents/" + path
	rsp, err := cl.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("loading url=%s. expected status 2xx, got %q", url, rsp.Status)
	}
	decoder := json.NewDecoder(rsp.Body)

	contents := []*content{}
	return contents, decoder.Decode(&contents)
}

func allUpstreamTemplates(repo, path string) (tmpls templates, e error) {
	tmpls = templates{}
	contents, e := upstreamFiles(repo, path)
	if e != nil {
		return nil, e
	}

	for _, c := range contents {
		if strings.HasPrefix(c.Name, "tpl_") && strings.HasSuffix(c.Name, ".go") {
			name := c.Name[4 : len(c.Name)-3]
			tmpls[name] = c
		}
	}
	return tmpls, nil
}
