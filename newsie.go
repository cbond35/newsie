package main

import (
	"fmt"
	"os"

	"github.com/cbbond/newsie-go/archnews"

	"github.com/docopt/docopt-go"
)

// handleLs prints all the items returned by Ls.
func handleLs(all bool) {
	items := archnews.Ls(all)

	for _, item := range items {
		fmt.Println(item)
	}
}

// handleFetch prints the message associated with Fetch (or not if prompt)
// and returns the number of unread items used by the pacman hook.
func handleFetch(prompt bool) int {
	status, msg := archnews.Fetch(prompt)

	fmt.Print(msg)
	return status
}

// main parses the docopt options and calls the appropriate newsie function.
func main() {

	usage := `newsie
	
Usage:
	newsie (-h | --help)
	newsie --version
	newsie browse [-a | --all]
	newsie clear
	newsie fetch [-p | --prompt]
	newsie ls [-a | --all]
	newsie read [-n <post_no> | --number <post_no>]

Commands:
	browse - Browse through unread (if no options are used) or all (if -a or --all
	options are used) posts. Quits when all posts have been viewed or if enters "n".
		
	clear - Marks all posts as read.
	
	fetch - Fetch the number of unread posts. If the -p or --prompt options
	are used, the user will be asked whether or not they would like to browse
	the posts.
		
	ls - Lists all unread posts. If the -a or --all options are used, all posts
	(regardless if they have been read or not) are listed. 
		
	read - Read the first post in the queue. The -n and --number options 
	specify which post number to read, consistent with the numbering from
	the 'ls' and 'ls --all' commands.

Options:
	-h  --help                       Show usage.
	-p  --prompt                     Prompt user to browse any unread posts.
	-a  --all                        Increase scope to all posts (read and unread).
	-n  --number                     Number of post to read (default = 1).`

	args, _ := docopt.ParseDoc(usage)

	a, _ := args.Bool("-a")
	all, _ := args.Bool("--all")

	p, _ := args.Bool("-p")
	prompt, _ := args.Bool("--prompt")

	n, _ := args.Bool("-n")
	number, _ := args.Bool("--number")

	archnews.New()
	status := 0

	if browse, _ := args.Bool("browse"); browse {
		archnews.Browse(a || all)
	} else if clear, _ := args.Bool("clear"); clear {
		archnews.Clear()
	} else if fetch, _ := args.Bool("fetch"); fetch {
		status = handleFetch(p || prompt)
	} else if ls, _ := args.Bool("ls"); ls {
		handleLs(a || all)
	} else if read, _ := args.Bool("read"); read && !(n || number) {
		prettyPost, _ := archnews.Read(1)
		fmt.Println(prettyPost)
	} else if read, _ := args.Bool("read"); read {
		postNum, err := args.Int("<post_no>")
		prettyPost, err := archnews.Read(postNum)

		if err != nil {
			fmt.Println("Invalid post number.")
		} else {
			fmt.Println(prettyPost)
		}
	}

	os.Exit(status)
}
