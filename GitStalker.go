package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nareix/curl"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

type Repository struct {
	FullName string `json:"full_name,omitempty"`
	Language string `json:"language,omitempty"`
}

var (
	username *string  = flag.String("u", "", "github username")
	keyword  []string = []string{"starred", "repos"}
)

func repos(key string) ([]string, error) {
	page := 1
	var items []string
	for {
		url := fmt.Sprintf("https://api.github.com/users/%s/%s?page=%d&per_page=100", *username, key, page)
		log.Println(url)
		err, content := curl.Bytes(url)
		if err != nil {
			return items, err
		}
		var repoList []Repository
		err = json.Unmarshal(content, &repoList)
		if err != nil {
			log.Println((string)(content))
			return items, err
		}
		if len(repoList) == 0 {
			break
		}
		for _, repo := range repoList {
			items = append(items, repo.FullName)
			log.Println(repo.FullName)
		}
		page++
	}

	return items, nil
}

func command(name string, args ...interface{}) *exec.Cmd {
	var a []string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			a = append(a, v)
		case []string:
			a = append(a, v...)
		}
	}
	c := exec.Command(name, a...)
	c.Stderr = os.Stderr
	return c
}

func action(key, fullname string) {
	repoDir := key + "/" + fullname
	repoInfo := strings.Split(fullname, "/")
	repoUser := repoInfo[0]
	repoName := repoInfo[1]

	exitCode := 0
	_, err := os.Stat(repoDir+"/.git")
	if err == nil {
		os.Chdir(repoDir)
		log.Printf("repo '%s' already exists, try update.\n", repoDir)
		c := command("git", "pull")
		if err := c.Run(); err != nil {
			log.Println(err)
		}
		return
	}

	if err := os.MkdirAll(repoDir, os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	url := fmt.Sprintf("https://github.com/%s/%s.git", repoUser, repoName)
	log.Printf("git clone %s into %s", url, repoDir)
	c := command("git", "clone", url, repoDir)
	if err := c.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		log.Printf("close %s failed, exit code %d, err: %v\n", repoDir, exitCode, err)
		_, err = os.Stat(repoDir)
		if err != nil {
			os.RemoveAll(repoDir)
		}
	}
}

func main() {
	flag.Parse()
	if len(*username) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if runtime.GOOS == "windows" {
		path := os.Getenv("PATH")
		path = fmt.Sprintf("%s;%s", `C:\Program Files\Git\bin`, path)
		path = fmt.Sprintf("%s;%s", `C:\Program Files (x86)\Git\bin`, path)
		os.Setenv("PATH", path)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(pwd)

	for _, key := range keyword {
		items, err := repos(key)
		if err != nil {
			log.Fatalln(err)
		}

		for _, fullname := range items {
			os.Chdir(pwd)
			go action(key, fullname)
		}
	}

}
