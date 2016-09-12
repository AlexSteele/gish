package main

import (
	"errors"
	"strings"

	"github.com/AlexSteele/deref"
	"github.com/andygrunwald/go-trending"
)

type file struct {
	name      string
	below     []file
	above     *file
	dir       bool
	populated bool
}

var ErrBadPath = errors.New("Bad path\n")

func (f *file) get(path string) (*file, error) {
	if len(path) == 0 || path == "." {
		return f, nil
	}
	if strings.HasPrefix(path, "./") {
		return f.get(path[1:])
	}
	if strings.HasPrefix(path, "..") {
		if f.above == nil {
			return nil, ErrBadPath
		}
		if path == ".." {
			return f.above, nil
		}
		if path[2] == '/' {
			return f.above.get(path[2:])
		}
		return nil, ErrBadPath
	}
	if strings.HasPrefix(path, "/") && !f.isRoot() {
		return f.root().get(path[1:])
	}
	for _, sub := range f.below {
		if strings.HasPrefix(path, sub.name) {
			if len(path) == len(sub.name) {
				return &sub, nil
			}
			if path[len(sub.name)] == '/' {
				return sub.get(path[len(sub.name):])
			}
		}
	}
	return nil, ErrBadPath
}

func (f *file) root() *file {
	curr := f
	for curr.above != nil {
		curr = curr.above
	}
	return curr
}

func (f *file) isRoot() bool {
	return f.above == nil
}

func (f *file) isRepo() bool {
	return !f.dir
}

func populate(f *file, ctxt *gish) error {
	if f.populated {
		return nil
	}
	switch f.name {
	case "trending":
		projects, err := ctxt.trend.GetProjects(trending.TimeToday, "")
		if err != nil {
			return err
		}
		for no, project := range projects {
			if no > 24 {
				break
			}
			// projStr := fmt.Sprintf("%d: %s (written in %s with %d ★ )",
			// 	no, project.Name, project.Language, project.Stars)
			projStr := project.Name
			f.below = append(f.below, file{name: projStr, dir: false})
		}
	case "users":
		users, _, err := ctxt.client.Users.ListAll(nil)
		if err != nil {
			return err
		}
		for _, user := range users {
			// userStr := fmt.Sprintf("%s - Followers: %d",
			// 	deref.String(user.Login), deref.Int(user.Followers))
			userStr := deref.String(user.Login)
			f.below = append(f.below, file{name: userStr})
		}
		// 		case "info":
		// 			user, _, err := ctxt.client.Users.Get(pathItems[1])
		// 			if err != nil {
		// 				fmt.Fprintf(os.Stderr, err.Error())
		// 			}
		// 			fmt.Printf("%s - Followers: %d\n", deref.String(user.Login), deref.Int(user.Followers))
	}
	f.populated = true
	return nil
}
