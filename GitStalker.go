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

	for _, key := range keyword {
		items, err := repos(key)
		if err != nil {
			log.Fatalln(err)
		}

		for _, fullname := range items {
			os.Chdir(pwd)

			repoInfo := strings.Split(fullname, "/")
			repoUser := repoInfo[0]
			repoName := repoInfo[1]

			if err := os.Mkdir(key, os.ModePerm); err != nil && !os.IsExist(err) {
				log.Fatalln(err)
			}
			os.Chdir(key)

			if err := os.Mkdir(repoUser, os.ModePerm); err != nil && !os.IsExist(err) {
				log.Fatalln(err)
			}
			os.Chdir(repoUser)

			exitCode := 0
			_, err = os.Stat(repoName)
			if err == nil {
				os.Chdir(repoName)
				log.Printf("repo '%s' already exists, try update.\n", repoInfo)
				c := command("git", "pull")
				if err := c.Run(); err != nil {
					log.Println(err)
				}
				continue
			}

			url := fmt.Sprintf("https://github.com/%s/%s.git", repoUser, repoName)
			log.Printf("git clone %s", url)
			c := command("git", "clone", url)
			if err := c.Run(); err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						exitCode = status.ExitStatus()
					}
				}
				log.Println(err)
				log.Printf("exit code %d\n", exitCode)
			}

		}
	}

}
