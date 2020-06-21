# newsie

newsie is a command line tool and [pacman hook](https://wiki.archlinux.org/index.php/pacman#Hooks) for reading Arch Linux newsfeed posts, inspired by [informant](https://github.com/bradford-smith94/informant) and Gentoo's [eselect news module](https://wiki.gentoo.org/wiki/Eselect). It is written in Go.

## Installation

It can be installed via the Arch User Repository.

```bash
yay -S newsie
```
...or built from source here.

```bash
go get
go build main.go
```

## Usage

**browse**: Browse through each unread post until you've read all unread posts or quit via *n*. You can optionally specify the *-a* or *--all* options to browse all posts (both read and unread).


```bash
newsie browse [-a | --all]
```

**clear**: Mark all posts as read.


```bash
newsie clear
```
**fetch**: Returns the number of posts that you haven't read yet. You can optionally specify the *-p* or *--prompt* options if you want to be given an option to **browse** any unread posts. This is the mechanism by which the pacman hook works.


```bash
newsie fetch [-p | --prompt]
```
**ls**: Prints a detailed list of posts that you haven't read yet or, if you specify the *-a* or *--all* options, all posts (both read and unread). Unread posts appear in red when you specify the *--all* option.


```bash
newsie ls [-a | --all]
```
**read**: Read the first post in the queue. If you don't specify the *-n* or *--number* options, the latest post will be displayed. If you specify *-n* or *--number*, the post matching the provided number will be displayed. The post number is consistent with the numbering from the **ls** and **ls** *--all* commands.


```bash
newsie read [-n <post_number> | --number <post_number>]
```


## pacman
The pacman hook is enabled by default if you install via the AUR. It will abort any upgrades/installs (i.e. *-S* or *-Syu*) if you have any unread items. You can disable it by symlinking the existing hook at */usr/share/libalpm/hooks/* to */dev/null*.

## License
[MIT](https://choosealicense.com/licenses/mit/)
