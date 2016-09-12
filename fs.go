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

func (f *file) get(ctxt *gish, path string) (*file, error) {
	if len(path) == 0 || path == "." {
		return f, nil
	}
	if strings.HasPrefix(path, "/") {
		return f.root().get(ctxt, path[1:])
	}
	if strings.HasPrefix(path, "./") {
		return f.get(ctxt, path[1:])
	}
	if strings.HasPrefix(path, "..") {
		if f.above == nil {
			return nil, ErrBadPath
		}
		if path == ".." {
			return f.above, nil
		}
		if path[2] == '/' {
			return f.above.get(ctxt, path[2:])
		}
		return nil, ErrBadPath
	}
	if f.dir && !f.populated {
		populate(ctxt, f)
	}
	for _, sub := range f.below {
		if strings.HasPrefix(path, sub.name) {
			if len(path) == len(sub.name) {
				return &sub, nil
			}
			if path[len(sub.name)] == '/' {
				return sub.get(ctxt, path[len(sub.name)+1:])
			}
		}
	}
	if f.name == "users" {
		user, _, err := ctxt.client.Users.Get(path)
		if err != nil {
			return nil, err
		}
		userFile := file{name: deref.String(user.Login), above: f}
		f.below = append(f.below, userFile)
		return &userFile, nil
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

func (f *file) path() string {
	var path string
	for curr := f; curr.above != nil; curr = curr.above {
		path = curr.name + "/" + path
	}
	path = "/" + path
	return path
}

func (f *file) url() string {
	return "https://www.github.com" + f.path()
}

func (f *file) isRoot() bool {
	return f.above == nil
}

func (f *file) isRepo() bool {
	return !f.dir
}

func populate(ctxt *gish, f *file) error {
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
		//
	}
	f.populated = true
	return nil
}
