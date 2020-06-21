package archnews

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/cbbond/newsie-golang/archparser"
	"github.com/cbbond/newsie-golang/termstyle"

	"github.com/mmcdole/gofeed"
)

// List of posts from the newsfeed.
var posts []*gofeed.Item

// Number of unread posts.
var numUnread = 0

// Location of cache in the filesystem.
var cacheFolder string

// Fully qualified cache name.
var cache string

// $USER or $SUDO_USER.
var userEnv string

// Maps hashed post titles to whether or not they've been read.
var cacheMap = make(map[string]bool)

// Link to Arch newsfeed posts.
var link = "https://www.archlinux.org/feeds/news/"

// locateCache locates the cache, depending on whether or not
// newsie is run with sudo. Pacman hooks are run under sudo so
// we need to take special care to find the correct path + permissions.
func locateCache() {
	userEnv = os.Getenv("USER")

	if userEnv != "root" { // Not using sudo.
		cacheFolder = filepath.Join(os.Getenv("HOME"), ".cache/newsie/")
	} else {
		userEnv = os.Getenv("SUDO_USER")
		cacheFolder = filepath.Join(
			"/home/", userEnv, ".cache/newsie/")
	}

	cache = filepath.Join(cacheFolder, "cache")
}

// cacheCheck ensures that the ~/.cache/newsie/ directory exists and creates
// the cache file if necessary.
func cacheCheck() error {
	// Get the cache directory from $HOME or $SUDO_USER.
	locateCache()

	group, err := user.Lookup(userEnv)

	if err != nil {
		return err
	}

	uid, _ := strconv.Atoi(group.Uid)
	gid, _ := strconv.Atoi(group.Gid)

	if _, err := os.Stat(cacheFolder); err != nil {
		if os.IsNotExist(err) {
			// Create the cache directory if it doesn't exist.
			if err := os.MkdirAll(cacheFolder, 0755); err != nil {
				return err
			}
			// Permissions get wacky if above is run as sudo. Make
			// sure cacheFolder + cache belong to $USER or $SUDO_USER.
			if err := syscall.Chown(cacheFolder, uid, gid); err != nil {
				return err
			}
		} else { // More serious, bail.
			return err
		}
	}

	if _, err := os.Stat(cache); os.IsNotExist(err) {
		// Same as above but for cache file.
		if os.IsNotExist(err) {
			if err := ioutil.WriteFile(cache, []byte{}, 0644); err != nil {
				return err
			}
			// Permissions issue w/ sudo again... is there a better way to do this?
			if err := syscall.Chown(cache, uid, gid); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// initPosts marks all the items in the cache as read.
func initCache() error {
	contents, err := ioutil.ReadFile(cache)

	if err != nil {
		return err
	}

	cachedPosts := strings.Split(string(contents), "\n")

	for i := 0; i < len(cachedPosts); i++ {
		cacheMap[cachedPosts[i]] = true
	}
	return nil
}

// hashTitle hashes the given title using SHA-1.
func hashTitle(title string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(title))

	return hex.EncodeToString(sha1.Sum(nil))
}

// cachePost stores read posts in the cache.
func cachePost(title string) error {
	hashedTitle := hashTitle(title)

	f, err := os.OpenFile(
		cache, os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(hashedTitle + "\n")
	return err
}

// countUnread compares the contents of the feed to the cache map
// and sets the number of unread items.
func countUnread() {
	for _, post := range posts {
		if !cacheMap[hashTitle(post.Title)] {
			numUnread++
		}
	}
}

// isRead returns true if the post with the given title has been read.
func isRead(title string) bool {
	return cacheMap[hashTitle(title)]
}

// promptBrowse is called from Fetch if prompt = true. It returns
// true if the user inputs "y", "yes", or hits enter.
func promptBrowse() bool {
	var userChoice string

	fmt.Print("Read them now? [Y/n] ")
	fmt.Scanln(&userChoice)
	userChoice = strings.TrimSpace(
		strings.ToLower(userChoice))

	return userChoice == "y" || userChoice == "yes" || userChoice == ""
}

// New sets up all the necessary data structures for archnews to function
// properly. New must (!) be called before Browse, Clear, Fetch, Ls, or Read.
func New() {
	if err := cacheCheck(); err != nil {
		fmt.Printf("Exiting: %s\n", err.Error())
		os.Exit(1)
	}
	if err := initCache(); err != nil {
		fmt.Printf("Exiting: %s\n", err.Error())
		os.Exit(1)
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(link)

	if err != nil {
		fmt.Printf("Exiting: %s\n", err.Error())
		os.Exit(1)
	}

	posts = feed.Items
	countUnread()
}

// Ls returns a formatted slice of (unread) items. If all is set to true,
// all posts will be returned. Otherwise only unread posts will be added.
func Ls(all bool) []string {
	var lsPosts []string

	for i, post := range posts {
		var listEntry string

		if all || (!all && !isRead(post.Title)) {

			if all && !isRead(post.Title) {
				listEntry = termstyle.StyleText(
					fmt.Sprintf("%-4s", strconv.Itoa(i+1)+"."), []string{"bold", "red"})
				listEntry += termstyle.StyleText(post.Title, []string{"red"})
			} else {
				listEntry = termstyle.StyleText(
					fmt.Sprintf("%-4s", strconv.Itoa(i+1)+"."), []string{"bold"})
				listEntry += post.Title
			}

			lsPosts = append(lsPosts, listEntry)
		}
	}
	return lsPosts
}

// Read returns a formatted post from the feed matching the post number.
// An error is returned if the post number is out of bounds.
func Read(postNum int) (string, error) {
	if postNum < 1 || postNum > len(posts) {
		return "", errors.New("invalid post number")
	}

	post := posts[postNum-1]
	prettyPost := archparser.MakePretty(post)

	if !isRead(post.Title) {
		err := cachePost(post.Title)

		if err == nil {
			numUnread--
		}
	}
	return prettyPost, nil
}

// Browse walks through the news posts, waiting for the user to move to the
// next post, run out of posts, or quit. If the all is set to true, all posts
// in the feed will be shown. Otherwise only unread posts will be shown.
func Browse(all bool) error {
	if !all && numUnread == 0 {
		fmt.Println("You don't have any unread news. Use the -a option to browse all posts.")
		return nil
	}

	userChoice := ""
	idx := 0

	for (userChoice == "y" || userChoice == "yes" || userChoice == "") && idx < len(posts) {
		out, err := exec.Command("clear").Output()

		if err != nil {
			return err
		}
		fmt.Print(string(out))

		currPost := posts[idx]

		if all || !isRead(currPost.Title) {
			prettyPost, _ := Read(idx + 1)
			fmt.Println(prettyPost)

			fmt.Print(termstyle.StyleText("\nContinue? [Y/n] ", []string{"bold"}))
			fmt.Scanln(&userChoice)
		}

		userChoice = strings.TrimSpace(strings.ToLower(userChoice))
		idx++
	}

	return nil
}

// Fetch returns the number of unread items and a status message about them.
// If prompt is true, the user will be prompted to browse the items. An
// empty mesage is returned if prompt is true.
func Fetch(prompt bool) (int, string) {
	if numUnread == 0 {
		return numUnread, "No news is good news.\n"
	}

	msg := termstyle.StyleText(
		fmt.Sprintf("* You have %d unread item(s).\n", numUnread), []string{"yellow"})

	if prompt {
		fmt.Print(msg)

		if yes := promptBrowse(); yes {
			Browse(false)
			return numUnread, ""
		}
	}

	cmd := termstyle.StyleText("newsie browse", []string{"bold", "green"})
	msg += fmt.Sprintf("Use %s to view them.\n", cmd)

	return numUnread, msg
}

// Clear stores all outstanding news items in the cache.
func Clear() int {
	numCleared := 0

	for _, post := range posts {
		if !isRead(post.Title) {
			err := cachePost(post.Title)

			if err != nil {
				fmt.Printf("Exiting: %s\n", err.Error())
				os.Exit(1)
			}
			numCleared++
		}
	}
	return numCleared
}
