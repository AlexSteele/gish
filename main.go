package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlexSteele/deref"
	"github.com/andygrunwald/go-trending"
	"github.com/google/go-github/github"
)

const (
	help   = "gish - Browse GitHub like it's your filesystem."
	badCmd = "Unrecognized command. (Hint: Type :help)"
)

type state struct {
	dir    string
	client *github.Client
	trend  *trending.Trending
}

func initState() *state {
	return &state{
		dir:    "/",
		client: github.NewClient(nil),
		trend:  trending.NewTrending(),
	}
}

func (s *state) fullPath(path string) string {
	if len(path) == 0 {
		return s.dir
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	if s.dir == "/" {
		return s.dir + path
	}
	return s.dir + "/" + path
}

func split(toSplit, splitter string) []string {
	split := strings.Split(toSplit, splitter)
	if len(split) > 0 &&
		len(split[len(split)-1]) == 0 {
		split = split[0 : len(split)-1]
	}
	if len(split) > 0 &&
		len(split[0]) == 0 {
		split = split[1:len(split)]
	}
	return split
}

func cd(ctxt *state, args ...string) string {
	if len(args) == 0 || args[0] == "." {
		return ""
	}
	path := args[0]
	if strings.HasPrefix(path, "..") {
		pathItems := split(ctxt.dir, "/")
		if len(pathItems) == 0 {
			return ""
		}
		newPathItems := pathItems[:len(pathItems)-1]
		ctxt.dir = "/" + strings.Join(newPathItems, "/")
		if len(path) > 2 {
			if path[2] == '/' {
				if len(path) > 3 {
					return cd(ctxt, path[3:])
				}
			} else {
				return cd(ctxt, path[2:])
			}
		}
		return ""
	}
	if strings.HasPrefix(path, ".") {
		if len(path) == 1 {
			return ""
		}
		if path[1] == '/' {
			if len(path) > 2 {
				return cd(ctxt, args[0][2:])
			}
			return ""
		} else {
			path = path[1:]
		}
	}
	fullPath := ctxt.fullPath(path)
	err := validatePath(fullPath)
	if err != nil {
		return err.Error()
	}
	ctxt.dir = fullPath
	return ""
}

func validatePath(path string) error {
	if path == "/trending" {
		return nil
	}

	if strings.HasPrefix(path, "/users") {
		if len(path) > len("/users") {
			rest := path[len("/users"):]
			split := split(rest, "/")
			if len(split) != 2 || split[1] != "info" || split[1] != "repos" {
				return errors.New("Bad path")
			}
		}
		return nil
	}

	return errors.New("Bad path.")
}

func ls(ctxt *state, args ...string) string {
	pathArg := ""
	if len(args) > 0 {
		pathArg = args[0]
	}
	fullPath := ctxt.fullPath(pathArg)
	pathItems := split(fullPath, "/")
	if len(pathItems) == 0 {
		return "trending\n" +
			"users"
	}
	switch pathItems[0] {
	case "trending":
		if len(pathItems) > 1 {
			break
		}
		projects, err := ctxt.trend.GetProjects(trending.TimeToday, "")
		if err != nil {
			return err.Error()
		}
		var res []string
		for no, project := range projects {
			if no > 24 {
				break
			}
			projStr := fmt.Sprintf("%d: %s (written in %s with %d ★ )\n",
				no, project.Name, project.Language, project.Stars)
			res = append(res, projStr)
		}
		return strings.Join(res, "\n")
	case "users":
		if len(pathItems) < 2 {
			users, _, err := ctxt.client.Users.ListAll(nil)
			if err != nil {
				return err.Error()
			}
			var res []string
			for _, user := range users {
				userStr := fmt.Sprintf("%s - Followers: %d\n",
					deref.String(user.Login), deref.Int(user.Followers))
				res = append(res, userStr)
			}
			return strings.Join(res, "\n")
		}
		if len(pathItems) == 2 {
			return "info\n" +
				"repos"
		}
		if len(pathItems) == 3 {
			switch pathItems[2] {
			case "info":
				user, _, err := ctxt.client.Users.Get(pathItems[1])
				if err != nil {
					return err.Error()
				}
				return fmt.Sprintf("%s - Followers: %d\n", deref.String(user.Login), deref.Int(user.Followers))
			case "repos":
				repos, _, err := ctxt.client.Repositories.List(pathItems[1], nil)
				if err != nil {
					return err.Error()
				}
				var res []string
				for no, repo := range repos {
					if no > 24 {
						break
					}
					repoStr := fmt.Sprintf("%s: %s (written in %s)\n",
						deref.String(repo.Name), deref.String(repo.Description), deref.String(repo.Language))
					res = append(res, repoStr)
				}
				return strings.Join(res, "\n")
			}
		}
	}

	return "Bad path. Try 'ls /'."
}

func cat(ctxt *state, args ...string) string {
	pathArg := ""
	if len(args) > 0 {
		pathArg = args[0]
	}
	fullPath := ctxt.fullPath(pathArg)
	pathItems := split(fullPath, "/")
	if len(pathItems) == 4 &&
		pathItems[0] == "users" &&
		pathItems[2] == "repos" {

		user := pathItems[1]
		repo := pathItems[3]

		readme, _, err := ctxt.client.Repositories.GetReadme(user, repo, nil)
		if err != nil {
			return err.Error()
		}
		content, err := readme.GetContent()
		if err != nil {
			return err.Error()
		}
		return content
	}
	return "Bad arg. " +
		"Args to cat should be of the form 'cat /users/repos/{repoName}'. " +
		"Try 'cat /users/torvalds/linux'."
}

var cmds = map[string]func(*state, ...string) string{
	"cd":  cd,
	"ls":  ls,
	"cat": cat,
}

func main() {
	in := bufio.NewReader(os.Stdin)
	ctxt := initState()
	for {
		fmt.Print("gish> ")
		raw, _ := in.ReadString('\n')
		trimmed := strings.Trim(raw, "\n \t")
		args := split(trimmed, " ")
		if len(args) == 0 {
			continue
		}
		cmd := args[0]
		switch cmd {
		case ":help":
			fmt.Println(help)
		default:
			handler, ok := cmds[cmd]
			if !ok {
				fmt.Println(badCmd)
				continue
			}
			var res string
			if len(args) > 1 {
				res = handler(ctxt, args[1:]...)
			} else {
				res = handler(ctxt)
			}
			fmt.Println(res)
		}
	}
}
