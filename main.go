package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"xkcd/xkcd"
)

var (
	build  = flag.Bool("build", false, "usage: xkcd -build")
	search = flag.String("search", "", "usage: xkcd -search=<term>")
	usage  = errors.New("usage: xkcd -build\nxkcd -search=<term>\nxkcd -build -search=<term>")
)

func main() {
	flag.Parse()
	if len(flag.Args()) > 0 {
		log.Fatal(usage)
	}

	if !*build && *search == "" {

		log.Fatal(usage)
	}
	var comics []xkcd.Comic
	var err error
	if *build {
		comics, err = xkcd.HandleComics(displayProgress)
		if err != nil {
			log.Fatal(err)
		}
		err = xkcd.SaveIndex(comics, "index.json")
		if err != nil {
			log.Fatal(err)
		}
	}
	if *search != "" && comics == nil {
		comics, err = xkcd.LoadIndex("index.json")
		if errors.Is(err, xkcd.ErrIndexNotFound) {
			comics, err = xkcd.HandleComics(displayProgress)
			if err != nil {
				log.Fatal(err)
			}
			err = xkcd.SaveIndex(comics, "index.json")
			if err != nil {
				log.Fatal(err)
			}
		} else if err != nil {
			log.Fatal(err)
		}
	}
	if *search != "" {
		results := xkcd.Search(comics, *search)
		for _, r := range results {
			fmt.Println(r.String())
		}
	}
}

func displayProgress(current, maximum int) {
	fmt.Printf("\rdownloaded %d/%d", current, maximum)
}
