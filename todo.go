package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/tabwriter"
)

var (
	cwd, _      = os.Getwd()
	dir         = flag.String("dir", "./", "directory to find todo's in")
	reIsTodo    = regexp.MustCompile(`(//|#|\*) (?i)todo.*`)
	reTodo      = regexp.MustCompile(`(?i)todo`)
	reName      = regexp.MustCompile(`\(([^\)]+)\)`)
	defaultName = "Unknown"
	raw         [][]byte
	wg          sync.WaitGroup
)

type todo struct {
	Name, Desc string
}

func newTodo(content string) *todo {
	var name, desc string

	idxs := reTodo.FindStringIndex(content)
	matches := reName.FindAllStringIndex(content, -1)

	if len(matches) > 0 {
		if matches[0][0]-idxs[1] < 2 {
			name = content[matches[0][0]+1 : matches[0][1]-1]
		}
	}

	if name != "" {
		desc = content[matches[0][1]:]
	} else {
		desc = content[idxs[1]:]
		name = defaultName
	}

	desc = strings.Trim(desc, " ")

	return &todo{
		Name: name,
		Desc: desc,
	}
}

func parse(path string, file os.FileInfo, err error) error {
	wg.Add(1)

	go func() {
		defer wg.Done()
		if file.IsDir() {
			return
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return
		}

		matches := reIsTodo.FindAll(data, -1)
		raw = append(raw, matches...)
	}()

	return nil
}

func main() {
	flag.Parse()

	if *dir == "./" {
		*dir = cwd
	}

	err := filepath.Walk(*dir, parse)
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprint(tabWriter, "Name\t")
	fmt.Fprint(tabWriter, "Description\n")
	for _, match := range raw {
		td := newTodo(string(match))
		fmt.Fprint(tabWriter, td.Name+"\t")
		fmt.Fprint(tabWriter, td.Desc+"\n")
	}
	tabWriter.Flush()
}
