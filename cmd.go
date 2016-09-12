package main

import (
	"fmt"
	"os"
)

type Cmd struct {
	Help string
	Run  func(*gish, ...string)
}

var cmds = map[string]Cmd{
	"ls":   lsCmd,
	"cd":   cdCmd,
	"less": lessCmd,
	"open": openCmd,
}

var (
	lsCmd = Cmd{
		Help: "Browse contents.",
		Run:  ls,
	}
	cdCmd = Cmd{
		Help: "Move around the Gish fs.",
		Run:  cd,
	}
	lessCmd = Cmd{
		Help: "View information about a repo.",
		Run:  less,
	}
	openCmd = Cmd{
		Help: "Open the relevant page in your browser.",
		Run:  open,
	}
)

func ls(ctxt *gish, args ...string) {
	var dir *file
	if len(args) == 0 {
		dir = ctxt.fs
	} else {
		var err error
		dir, err = ctxt.fs.get(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}
	}
	if !dir.dir {
		fmt.Println(dir.name)
		return
	}
	if !dir.populated {
		populate(dir, ctxt)
	}

	for _, f := range dir.below {
		fmt.Println(f.name)
	}
}
func cd(ctxt *gish, args ...string) {
	if len(args) == 0 {
		return
	}
	dir, err := ctxt.fs.get(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	ctxt.fs = dir
}

func less(ctxt *gish, args ...string) {
	if len(args) == 0 {
		return
	}
	file, err := ctxt.fs.get(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	if !file.isRepo() {
		fmt.Fprintf(os.Stderr, "'less' can only be called on repos.")
		return
	}
	// TODO: Ensure user exists
	user, repo := file.above.name, file.name
	readme, _, err := ctxt.client.Repositories.GetReadme(user, repo, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	content, err := readme.GetContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	fmt.Println(content)

}

func open(ctxt *gish, args ...string) {
	// TODO: Impl
	return
}

func help(cmdName string) {
	if len(cmdName) == 0 {
		fmt.Println("gish - Browse GitHub like it's your filesystem.\n")
		for name, cmd := range cmds {
			fmt.Printf("\t%s - %s\n\n", name, cmd.Help)
		}
		fmt.Println("\nType 'help cmd' to get help with a specific command.")
	} else {
		cmd, ok := cmds[cmdName]
		if !ok {
			fmt.Println(BadCmd)
			return
		}
		fmt.Println(cmd.Help)
	}
}
