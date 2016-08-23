package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/andygrunwald/go-trending"
)

const (
	help = "gish - Browse GitHub like your filesystem."
	badCmd = "Unrecognized command. (Hint: Type :help)"
)

func deref(s *string) string {
	if (s == nil) {
		return ""
	}
	return *s
}

func main() {
	in := bufio.NewReader(os.Stdin)
	client := github.NewClient(nil)
	trend := trending.NewTrending()
	for {
		fmt.Print("gish> ")
		raw, _ := in.ReadString('\n')
		trimmed := strings.Trim(raw, "\n \t")
		args := strings.Split(trimmed, " ")
		if len(args) == 0 {
			continue
		}
		cmd := args[0]
		switch (cmd) {
		case ":help":
			fmt.Println(help)
		case "ls":
			switch {
			case strings.HasPrefix(args[1], "/users/"):
				if len(args[1]) < 8 {
					continue
				}
				user := args[1][7:]
				repos, _, err := client.Repositories.List(user, nil)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					continue
				}
				for no, repo := range repos {
					if (no > 24) {
						break
					}
					fmt.Printf("%s: %s (written in %s)\n",
						deref(repo.Name), deref(repo.Description), deref(repo.Language))
				}
			case args[1] == "/trending":
				projects, err := trend.GetProjects(trending.TimeToday, "")
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					continue
				}
				for no, project := range projects {
					if (no > 24) {
						break
					}
					fmt.Printf("%d: %s (written in %s with %d ★ )\n",
						no, project.Name, project.Language, project.Stars)
				}				
			}
		case "cat":
			if (len(args) > 1 && strings.HasPrefix(args[1], "/users/")) {
				userAndRepo := strings.Split(args[1][7:], "/")
				if len(userAndRepo) != 2 {
					continue
				}
				readme, _, err := client.Repositories.GetReadme(userAndRepo[0], userAndRepo[1], nil)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					continue
				}
				content, err := readme.GetContent()
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					continue
				}
				fmt.Println(content)
			}
		default:
			fmt.Println(badCmd)
		}
	}
}
