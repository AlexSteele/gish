package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/AlexSteele/deref"
	"github.com/andygrunwald/go-trending"
	"github.com/google/go-github/github"
)

const TrendingUsersHelp = `trending-users
		Prints the top 25 trending users' usernames.`

func TrendingUsers() {
	trend := trending.NewTrending()
	users, err := trend.GetDevelopers(trending.TimeToday, "")
	if err != nil {
		log.Fatal(err)
	}
	for no, user := range users {
		if no > 24 {
			break
		}
		fmt.Println(user.DisplayName)
	}
}

const TrendingReposHelp = `trending-repos
		Prints the top 25 trending repos in the format 'username/reponame (language - stars)'`

func TrendingRepos() {
	trend := trending.NewTrending()
	projects, err := trend.GetProjects(trending.TimeToday, "")
	if err != nil {
		log.Fatal(err)
	}
	for no, project := range projects {
		if no > 24 {
			break
		}
		fmt.Printf("%s (%s - %d stars)\n", project.Name, project.Language, project.Stars)
	}
}

const UserSummaryHelp = `user-summary 'login'
		Prints information about the user with the given login.'`

func UserSummary(login string) {
	client := github.NewClient(nil)
	user, _, err := client.Users.Get(login)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Name:         %s\n", deref.String(user.Name))
	fmt.Printf("Bio:          %s\n", deref.String(user.Bio))
	fmt.Printf("Company:      %s\n", deref.String(user.Company))
	fmt.Printf("Location:     %s\n", deref.String(user.Location))
	fmt.Printf("Email:        %s\n", deref.String(user.Email))
	fmt.Printf("Public Repos: %d\n", deref.Int(user.PublicRepos))
	fmt.Printf("Followers:    %d\n", deref.Int(user.Followers))
	fmt.Printf("Following:    %d\n", deref.Int(user.Following))
}

const RepoSummaryHelp = `repo-summary 'name'
		Prints information about the repo with the given name, 
		where 'name' follows the format 'username/reponame'`

func RepoSummary(name string) {
	args := strings.Split(name, "/")
	if len(args) < 2 {
		log.Fatalf("Must give user/repo combination.")
	}
	user, reponame := args[0], args[1]
	client := github.NewClient(nil)
	repo, _, err := client.Repositories.Get(user, reponame)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Description: %s\n", deref.String(repo.Description))
	fmt.Printf("Language:    %s\n", deref.String(repo.Language))
	fmt.Printf("Created at:  %v\n", *repo.CreatedAt)
	fmt.Printf("Hompage:     %s\n", deref.String(repo.Homepage))
	fmt.Printf("Pushed at:   %v\n", *repo.PushedAt)
	fmt.Printf("Forks:       %d\n", deref.Int(repo.ForksCount))
	fmt.Printf("Open Issues: %d\n", deref.Int(repo.OpenIssuesCount))
	fmt.Printf("Stargazers:  %d\n", deref.Int(repo.StargazersCount))
	fmt.Printf("Subscribers: %d\n", deref.Int(repo.SubscribersCount))
	fmt.Printf("Watchers:    %d\n", deref.Int(repo.WatchersCount))
}

const ViewFileHelp = `view-file 'location'
		Prints the contents of the file at 'location' to stdout, 
		where location follows the format 'username/reponame/filename'`

func ViewFile(location string) {
	args := strings.SplitN(location, "/", 3)
	if len(args) < 3 {
		fmt.Println(log.Flags())
		log.Fatal("Must give user/repo/path combination.")
	}
	user, repo, path := args[0], args[1], args[2]
	client := github.NewClient(nil)
	fileContent, dirContent, _, err := client.Repositories.GetContents(user, repo, path, nil)
	if err != nil {
		log.Fatal(err)
	}
	if fileContent != nil {
		content, err := fileContent.Decode()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(content))
	} else {
		for _, f := range dirContent {
			fmt.Println(deref.String(f.Path))
		}
	}
}

const ReadmeHelp = `readme 'repo'
		Prints the contents of the given repo's README.md to stdout.
		'repo' should be in the form 'username/reponame'. readme is 
		a convenience wrapper for 'view-file repo/README.md'.`

func Readme(repo string) {
	ViewFile(repo + "/README.md")
}

const SearchReposHelp = `search-repos 'query'
		Search for repos matching the given query.
		Prints results in 'username/reponame - description' form. `

func SearchRepos(query string) {
	client := github.NewClient(nil)
	opt := github.SearchOptions{ListOptions: github.ListOptions{PerPage: 10}}
	res, _, err := client.Search.Repositories(query, &opt)
	if err != nil {
		log.Fatal(err)
	}
	for _, repo := range res.Repositories {
		fmt.Printf("%s/%s - %s\n", deref.String(repo.Owner.Login),
			deref.String(repo.Name), deref.String(repo.Description))
	}
}

const SearchUsersHelp = `search-users 'login'
		Display the result of searching for users matching the given login.`

func SearchUsers(login string) {
	client := github.NewClient(nil)
	opt := github.SearchOptions{ListOptions: github.ListOptions{PerPage: 10}}
	res, _, err := client.Search.Users(login, &opt)
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range res.Users {
		fmt.Println(deref.String(user.Login))
	}
}

const OpenHelp = `open 'entity'
		Open the page associated with the given entity (i.e. user, repo, or file) in your browser.
		Entity should be in 'username/reponame/filename' format, where 'reponame'
		and 'filename' are optional`

func Open(entity string) {
	url := "https://www.github.com/"
	args := strings.SplitN(entity, "/", 3)
	if len(args) > 0 {
		user := args[0]
		url += user
	}
	if len(args) > 1 {
		repo := args[1]
		url += "/" + repo
	}
	if len(args) > 2 {
		path := args[2]
		url += "/blob/master/" + path
	}

	cmd := exec.Command("open", url)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

const UrlHelp = `url 'entity'
		Prints the URL of the page associated with the given entity (i.e. user, repo. or file).
		If the entity is a repo, prints a clonable (via SSH) URL`

func Url(entity string) {
	url := "https://www.github.com/"
	args := strings.SplitN(entity, "/", 3)
	if len(args) > 0 {
		user := args[0]
		url += user
	}
	if len(args) > 1 {
		repo := args[1]
		url += "/" + repo
	}
	if len(args) == 2 {
		url += ".git"
	} else if len(args) > 2 {
		path := args[2]
		url += "/blob/master/" + path
	}
	fmt.Println(url)
}

const HelpHelp = `help [command]
		Prints usage info for the given command or, if none is given, prints general usage.`

var HelpBlurbs = map[string]string{
	"trending-users": TrendingUsersHelp,
	"trending-repos": TrendingReposHelp,
	"user-summary":   UserSummaryHelp,
	"repo-summary":   RepoSummaryHelp,
	"view-file":      ViewFileHelp,
	"readme":         ReadmeHelp,
	"search-users":   SearchUsersHelp,
	"search-repos":   SearchReposHelp,
	"open":           OpenHelp,
	"url":            UrlHelp,
	"help":           HelpHelp,
}

func Help(command ...string) {
	if len(command) == 0 {
		var sorted []string
		for _, v := range HelpBlurbs {
			sorted = append(sorted, v)

		}
		sort.Strings(sorted)
		fmt.Print(Usage)
		for _, v := range sorted {
			fmt.Println("\t", v)
		}
	} else if help, ok := HelpBlurbs[command[0]]; ok {
		fmt.Println(help)
	} else {
		fmt.Println(UnrecognizedCommandHelp)
	}
}

// func Repl() {
// 	in := bufio.NewReader(os.Stdin)
// 	for {
// 		fmt.Print("gish> ")
// 		raw, _ := in.ReadString('\n')
// 		trimmed := strings.Trim(raw, "\n \t")
// 		args := strings.Split(trimmed, " ")
// 		if len(args) == 0 || len(args[0]) == 0 {
// 			continue
// 		}
// 		name := args[0]
// 		if name == "exit" {
// 			fmt.Println("Goodbye!")
// 			return
// 		}
// 	}
// }

var UnrecognizedCommandHelp = "Unrecognized command. Try '" + os.Args[0] + " help' to see all commands."

var Usage = "Gish - Browse GitHub from the command line.\n\n" +
	"Usage: " + os.Args[0] + " 'command'.\n\n"

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		fmt.Print(Usage)
		return
	}

	switch os.Args[1] {
	case "trending-users":
		TrendingUsers()
	case "trending-repos":
		TrendingRepos()
	case "user-summary":
		if len(os.Args) < 3 {
			log.Fatal("Must give user.")
		}
		UserSummary(os.Args[2])
	case "repo-summary":
		if len(os.Args) < 3 {
			log.Fatal("Must give repo.")
		}
		RepoSummary(os.Args[2])
	case "view-file":
		if len(os.Args) < 3 {
			log.Fatal("Must give file.")
		}
		ViewFile(os.Args[2])
	case "readme":
		if len(os.Args) < 3 {
			log.Fatal("Must give repo.")
		}
		Readme(os.Args[2])
	case "search-users":
		if len(os.Args) < 3 {
			log.Fatal("Must give user.")
		}
		SearchUsers(os.Args[2])
	case "search-repos":
		if len(os.Args) < 3 {
			log.Fatal("Must give search term.")
		}
		SearchRepos(os.Args[2])
	case "open":
		if len(os.Args) < 3 {
			Open("")
		} else {
			Open(os.Args[2])
		}
	case "url":
		if len(os.Args) < 3 {
			Url("")
		} else {
			Url(os.Args[2])
		}
	case "help":
		Help(os.Args[2:]...)
	default:
		log.Fatalf(UnrecognizedCommandHelp)
	}
}
