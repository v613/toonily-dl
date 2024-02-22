package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	regex  = regexp.MustCompile(`\bhttps:\/\/cdn.\b.*\.jpg`)
	client = http.Client{}
)

func main() {
	url := os.Args[1]
	if url == "" {
		fmt.Printf("invalid argument: URL\nexample: toonily-dl <URL>\n")
		os.Exit(1)
	}
	scanner := bufio.NewScanner(bytes.NewReader(Wget(url)))
	scanner.Split(bufio.ScanLines)

	var title string
	var chapters []string
	var chapterSection bool
	var cover string

	for scanner.Scan() {
		line := scanner.Text()

		if chapterSection {
			// 9 <== `<a href="`
			// 3 <== `/">`
			chapters = append(chapters, line[9:len(line)-3])
			chapterSection = false
			continue
		}
		chapterSection = strings.Contains(line, "<li class=\"wp-manga-chapter")
		if cover == "" {
			if sIdx := strings.Index(line, `data-src="`); sIdx > 0 {
				sIdx += 10 // 10 <== `data-src="`
				if eIdx := strings.Index(line[sIdx:], `"`); eIdx > 0 {
					cover = line[sIdx : sIdx+eIdx]
				}
			}
		}

		if title == "" {
			if strings.HasPrefix(line, "<title>") {
				// 12 <== `<title>Read `
				// 24 <== ` Manga - Toonily</title>`
				title = line[12 : len(line)-24]
			}
		}
	}

	MakeDir(title)
	os.Chdir(title)
	DownloadCover(cover)

	fmt.Println("Download:", title)
	for _, url := range chapters {
		sl := strings.Split(url, "/")
		chapter := sl[len(sl)-1]
		MakeDir(chapter)
		os.Chdir(chapter)

		fmt.Println("Working on", chapter)
		page := bufio.NewScanner(bytes.NewReader(Wget(url)))
		page.Split(bufio.ScanLines)

		for page.Scan() {
			img := regex.FindString(page.Text())
			if len(img) == 0 {
				continue
			}
			if err := DownloadFile(img); err != nil {
				fmt.Println("[error]", err)
			}
		}
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