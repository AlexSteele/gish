package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/andygrunwald/go-trending"
	"github.com/google/go-github/github"
)

func split(toSplit, by string) []string {
	split := strings.Split(toSplit, by)
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

type gish struct {
	fs     *file
	client *github.Client
	trend  *trending.Trending
}

func fs() *file {
	root := file{name: "/", populated: true, dir: true}
	users := file{name: "users", above: &root, dir: true}
	trending := file{name: "trending", above: &root, dir: true}
	root.below = append(root.below, users, trending)
	return &root
}

const (
	BadCmd = "Unrecognized command. (Hint: Type 'help')\n"
)

func main() {
	if len(os.Args) < 2 {
		help("")
		return
	}
	cmdName := os.Args[1]
	ctxt := &gish{
		fs:     fs(),
		client: github.NewClient(nil),
		trend:  trending.NewTrending(),
	}
	if cmdName == "repl" {
		repl(ctxt)
		return
	}
	if cmdName == "help" {
		if len(os.Args) > 2 {
			help(os.Args[2])
		} else {
			help("")
		}
		return
	}
	cmd, ok := cmds[cmdName]
	if !ok {
		fmt.Println(BadCmd)
		os.Exit(1)
	}
	cmd.Run(ctxt, os.Args[2:]...)
}

func repl(ctxt *gish) {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("gish> ")
		raw, _ := in.ReadString('\n')
		trimmed := strings.Trim(raw, "\n \t")
		args := split(trimmed, " ")
		if len(args) == 0 {
			continue
		}
		name := args[0]
		if name == "exit" {
			return
		}
		if name == "help" {
			help(name)
			continue
		}
		if cmd, ok := cmds[name]; ok {
			cmd.Run(ctxt, args[1:]...)
		} else {
			fmt.Fprintf(os.Stderr, BadCmd)
		}
	}
}
