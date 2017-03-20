// +build ignore

package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func downloadAmalgamationCodes() error {
	site := "https://www.sqlite.org/download.html"
	log.Printf("Scraping %v\n", site)
	doc, err := goquery.NewDocument(site)
	if err != nil {
		return err
	}
	var url string
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		if url == "" && strings.HasPrefix(s.Text(), "sqlite-amalgamation-") {
			url = "https://www.sqlite.org/2017/" + s.Text()
		}
	})
	if url == "" {
		return errors.New("amalgamation code not found")
	}
	log.Printf("Downloading %v\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Extracting %v\n", path.Base(url))
	r, err := zip.NewReader(bytes.NewReader(b), resp.ContentLength)
	if err != nil {
		return err
	}

	for _, zf := range r.File {
		var f *os.File
		switch path.Base(zf.Name) {
		case "sqlite3.c":
			f, err = os.Create("sqlite3-binding.c")
		case "sqlite3.h":
			f, err = os.Create("sqlite3-binding.h")
		case "sqlite3ext.h":
			f, err = os.Create("sqlite3ext.h")
		default:
			continue
		}
		if err != nil {
			return err
		}
		zr, err := zf.Open()
		if err != nil {
			return err
		}

		_, err = io.WriteString(f, "#ifndef USE_LIBSQLITE3\n#define SQLITE_DISABLE_INTRINSIC 1\n")
		if err != nil {
			zr.Close()
			f.Close()
			return err
		}
		_, err = io.Copy(f, zr)
		if err != nil {
			zr.Close()
			f.Close()
			return err
		}
		_, err = io.WriteString(f, "#else // USE_LIBSQLITE3\n// If users really want to link against the system sqlite3 we\n// need to make this file a noop.\n#endif\n")
		if err != nil {
			zr.Close()
			f.Close()
			return err
		}
		zr.Close()
		f.Close()
		log.Printf("Extracted %v\n", filepath.Base(f.Name()))
	}
	return nil
}

func downloadModuleCodes() error {
	modules := []struct {
		file string
		rev  string
	}{
		{
			file: "csv.c",
			rev:  "531a46cbad789fca0aa9db69a0e6c8ac9e68767d",
		},
	}

	for _, module := range modules {
		url := fmt.Sprintf("https://www.sqlite.org/src/raw/ext/misc/%s?name=%s", module.file, module.rev)
		log.Printf("Downloading %v\n", url)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		f, err := os.Create("sqlite3-" + module.file)
		if err != nil {
			resp.Body.Close()
			return err
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			resp.Body.Close()
			f.Close()
			return err
		}
		resp.Body.Close()
		f.Close()
		log.Printf("Saved %v\n", f.Name())
	}
	return nil
}

func main() {
	err := downloadAmalgamationCodes()
	if err != nil {
		log.Fatal(err)
	}

	err = downloadModuleCodes()
	if err != nil {
		log.Fatal(err)
	}
}
