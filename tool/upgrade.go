// +build ignore

package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	site := "https://api.github.com/repos/CanonicalLtd/sqlite/releases/latest"
	fmt.Printf("scraping %v\n", site)
	resp, err := http.Get(site)
	if err != nil {
		log.Fatal(err)
	}
	latest := &struct {
		Assets []struct {
			URL string `json:"browser_download_url"`
		} `json:"assets"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(latest)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var url string
	for _, asset := range latest.Assets {
		if !strings.Contains(asset.URL, "sqlite-src-") {
			continue
		}
		url = asset.URL
		break
	}

	if url == "" {
		return
	}
	fmt.Printf("downloading %v\n", url)
	resp, err = http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		log.Fatal(err)
	}

	fmt.Printf("extracting %v\n", path.Base(url))
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		resp.Body.Close()
		log.Fatal(err)
	}
	resp.Body.Close()
	tr := tar.NewReader(r)
	if err != nil {
		log.Fatal(err)
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var f *os.File
		switch path.Base(header.Name) {
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
			log.Fatal(err)
		}

		_, err = io.WriteString(f, "#ifndef USE_LIBSQLITE3\n")
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(tr)
		for scanner.Scan() {
			text := scanner.Text()
			if text == `#include "sqlite3.h"` {
				text = `#include "sqlite3-binding.h"`
			}
			_, err = fmt.Fprintln(f, text)
			if err != nil {
				break
			}
		}
		err = scanner.Err()
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
		_, err = io.WriteString(f, "#else // USE_LIBSQLITE3\n // If users really want to link against the system sqlite3 we\n// need to make this file a noop.\n #endif")
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
		f.Close()
		fmt.Printf("extracted %v\n", filepath.Base(f.Name()))
	}
}
