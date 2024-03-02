package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	regex  = regexp.MustCompile(`\bhttps:\/\/cdn.\b.*\.(?i)(jpg|jpeg|png|webp)`)
	client = http.Client{}
	usage  = `
SYNOPSIS
	toonily-dl [flag]... <URL>
FLAGS
	-h
	    Print help message
	-c
	    Indicate the chapter's list to download.
	    Example:
		    toonily-dl -c 3         :: download only chapter 3
		    toonily-dl -c 13:       :: download chapters starting with 13 until the end
		    toonily-dl -c 213:321   :: download chapters starting with 213 to 321
		    toonily-dl -c :3210     :: download chapters up to 3210
`
)

type Book struct {
	Title       string
	Description string
	Cover       string
	Chapters    []string
}

func main() {
	flagC := flag.String("c", "", "Indicate the chapters's list to download")
	flagH := flag.Bool("h", false, "Print help message")
	flagD := flag.Bool("d", false, "Save description")
	flag.Parse()

	if *flagH {
		fmt.Println(usage)
		os.Exit(0)
	}

	url := os.Args[len(os.Args)-1]
	if url == "" {
		fmt.Printf("invalid argument: URL\nexample: toonily-dl <URL>\n")
		os.Exit(1)
	}

	chapterRange := [2]int{}
	for idx, v := range strings.Split(*flagC, ":") {
		n, err := strconv.Atoi(v)
		switch {
		case v == "" && idx == 0:
		case v == "" && idx == 1:
			n = 1 << 20
		case err != nil, idx > 1, (idx == 1 && n < chapterRange[0]):
			fmt.Println(usage)
			os.Exit(1)
		}
		chapterRange[idx] = n
	}

	scanner := bufio.NewScanner(bytes.NewReader(Wget(url)))
	scanner.Split(bufio.ScanLines)

	book := Book{}
	var chapterSection bool
	var summarySection bool

	for scanner.Scan() {
		line := scanner.Text()

		if chapterSection {
			// 9 <== `<a href="`
			// 3 <== `/">`
			link := line[9 : len(line)-3]
			n, err := strconv.Atoi(link[strings.LastIndex(link, "-")+1:])
			if err != nil {
				fmt.Printf("Unexpected chapter %q is about to be downloaded\n", link[strings.LastIndex(link, "/")+1:])
			}
			between := chapterRange[0] <= n && n <= chapterRange[1]
			exact := chapterRange[0] == n && chapterRange[1] == 0
			if between || exact || *flagC == "" || err != nil {
				book.Chapters = append(book.Chapters, link)
			}
			chapterSection = false
			continue
		}
		chapterSection = strings.Contains(line, "<li class=\"wp-manga-chapter")

		if summarySection {
			book.Description = line[4 : len(line)-4]
			summarySection = false
			continue
		}
		if book.Description == "" {
			summarySection = strings.Contains(line, "summary__content")
		}
		if book.Cover == "" {
			if sIdx := strings.Index(line, `data-src="`); sIdx > 0 {
				sIdx += 10 // 10 <== `data-src="`
				if eIdx := strings.Index(line[sIdx:], `"`); eIdx > 0 {
					book.Cover = line[sIdx : sIdx+eIdx]
				}
			}
		}

		if book.Title == "" {
			if strings.HasPrefix(line, "<title>") {
				// 12 <== `<title>Read `
				// 24 <== ` Manga - Toonily</title>`
				book.Title = line[12 : len(line)-24]
				book.Title = strings.ReplaceAll(book.Title, "&#8217;", "'")
				book.Title = strings.ReplaceAll(book.Title, "&#8230;", "...")
			}
		}
	}

	MakeDir(book.Title)
	os.Chdir(book.Title)
	DownloadCover(book.Cover)

	if *flagD {
		SaveTextToFile(book.Description, "description.txt")
	}

	fmt.Println("Download:", book.Title)
	for _, url := range book.Chapters {
		sl := strings.Split(url, "/")
		slen := len(sl)
		chapter := fmt.Sprintf("chapter-%03s", sl[len(sl)-1][8:]) // 8 <== `chapter-`
		if !strings.HasPrefix(sl[slen-1], "chapter-") {
			chapter = sl[slen-1] // unexpected chapter
		}
		MakeDir(chapter)
		os.Chdir(chapter)

		fmt.Println("Working on", chapter)
		page := bufio.NewScanner(bytes.NewReader(Wget(url)))
		page.Split(bufio.ScanLines)

		var wg sync.WaitGroup
		for page.Scan() {
			img := regex.FindString(page.Text())
			if len(img) == 0 {
				continue
			}
			wg.Add(1)
			go func(wg *sync.WaitGroup, img string) {
				defer wg.Done()
				if err := DownloadFile(img); err != nil {
					fmt.Println("[error]", err)
				}
			}(&wg, img)
		}
		wg.Wait()
		os.Chdir("../")

		if cnt, err := os.ReadDir(chapter); err == nil {
			fmt.Println("Downloaded", len(cnt), "file(s)")
		}
	}

}

func MakeDir(s string) {
	if err := os.Mkdir(s, os.ModePerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			fmt.Printf("cannot create directory %q\n%v\n", s, err)
			os.Exit(1)
		}
	}
}

func Wget(url string) []byte {
	r, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer r.Body.Close()

	if r.StatusCode >= 400 {
		fmt.Printf("[%d] %s\n", r.StatusCode, url)
	}

	data, _ := io.ReadAll(r.Body)
	return data
}

func DownloadFile(url string) error {
	filename := FilenameFromURL(url)
	if filename == "" {
		return errors.New("invalid URL:" + url)
	}
	if _, err := os.Stat(filename); err == nil {
		// file already present, skip it.
		return nil
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authority", "cdn.toonily.com")
	req.Header.Set("Referer", "https://toonily.com/")

	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, r.Body)
	return err
}

func FilenameFromURL(url string) string {
	if !strings.HasPrefix(url, "http") {
		return ""
	}
	sl := strings.Split(url, "/")
	if l := len(sl); l > 2 {
		return sl[l-1]
	}
	return ""
}

func DownloadCover(url string) {
	if _, err := os.Stat("cover.jpg"); os.IsNotExist(err) {
		if err = DownloadFile(url); err != nil {
			fmt.Println("cannot download cover image ==> ", err)
			return
		}
		if err = os.Rename(FilenameFromURL(url), "cover.jpg"); err != nil {
			fmt.Println(err)
		}
	}
}

func SaveTextToFile(txt, dst string) {
	if _, err := os.Stat(dst); err == nil {
		return
	}

	file, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	txt = strings.ReplaceAll(txt, "s/&#8217;", "'")
	txt = strings.ReplaceAll(txt, "&#8230;", "...")
	if _, err = file.WriteString(txt); err != nil {
		fmt.Println(err)
	}
}
