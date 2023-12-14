//go:build !cgo && upgrade && ignore
// +build !cgo,upgrade,ignore

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const buildFlags = "-DSQLITE_ENABLE_UPDATE_DELETE_LIMIT=1"

func main() {
	err := func() error {
		fmt.Println("Go-SQLite3 Upgrade Tool")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		if filepath.Base(wd) != "upgrade" {
			return fmt.Errorf("Current directory is %q but should run in upgrade directory", wd)
		}

		// Download Source
		source, hash, err := download("sqlite-src-")
		if err != nil {
			return fmt.Errorf("failed to download: sqlite-src; %v", err)
		}
		fmt.Printf("Download successful and verified hash %x\n", hash)

		// Extract Source
		baseDir, err := extractZip(source)
		if baseDir != "" && !filepath.IsAbs(baseDir) {
			defer func() {
				fmt.Println("Cleaning up source: deleting", baseDir)
				os.RemoveAll(baseDir)
			}()
		}
		if err != nil {
			return fmt.Errorf("failed to extract source: %v", err)
		}
		fmt.Println("Extracted sqlite source to", baseDir)

		// Build amalgamation files (OS-specific)
		fmt.Printf("Starting to generate amalgamation with build flags: %s\n", buildFlags)
		if err := buildAmalgamation(baseDir, buildFlags); err != nil {
			return fmt.Errorf("failed to build amalgamation: %v", err)
		}
		fmt.Println("SQLite3 amalgamation built")

		// Patch bindings
		patchSource(baseDir, "sqlite3.c", "../sqlite3-binding.c", "ext/userauth/userauth.c")
		patchSource(baseDir, "sqlite3.h", "../sqlite3-binding.h", "ext/userauth/sqlite3userauth.h")
		patchSource(baseDir, "sqlite3ext.h", "../sqlite3ext.h")

		fmt.Println("Done patching amalgamation")
		return nil
	}()
	if err != nil {
		log.Fatal("Returned with error:", err)
	}
}

func download(prefix string) (content, hash []byte, err error) {
	year := time.Now().Year()

	site := "https://www.sqlite.org/download.html"
	//fmt.Printf("scraping %v\n", site)
	doc, err := goquery.NewDocument(site)
	if err != nil {
		return nil, nil, err
	}

	url, hashString := "", ""
	doc.Find("tr").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		found := false
		s.Find("a").Each(func(_ int, s *goquery.Selection) {
			if strings.HasPrefix(s.Text(), prefix) {
				found = true
				url = fmt.Sprintf("https://www.sqlite.org/%d/", year) + s.Text()
			}
		})
		if found {
			s.Find("td").Each(func(_ int, s *goquery.Selection) {
				text := s.Text()
				split := strings.Split(text, "(sha1: ")
				if len(split) < 2 {
					return
				}
				text = split[1]
				hashString = strings.Split(text, ")")[0]
			})
		}
		return !found
	})

	targetHash, err := hex.DecodeString(hashString)
	if err != nil || len(targetHash) != sha1.Size {
		return nil, nil, fmt.Errorf("unable to find valid sha1 hash on sqlite.org: %q", hashString)
	}

	if url == "" {
		return nil, nil, fmt.Errorf("unable to find prefix '%s' on sqlite.org", prefix)
	}

	fmt.Printf("Downloading %v\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}

	// Ready Body Content
	shasum := sha1.New()
	content, err = ioutil.ReadAll(io.TeeReader(resp.Body, shasum))
	defer resp.Body.Close()
	if err != nil {
		return nil, nil, err
	}

	computedHash := shasum.Sum(nil)
	if !bytes.Equal(targetHash, computedHash) {
		return nil, nil, fmt.Errorf("invalid hash of file downloaded from %q: got %x instead of %x", url, computedHash, targetHash)
	}

	return content, computedHash, nil
}

func extractZip(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	if len(zr.File) == 0 {
		return "", errors.New("no files in zip archive")
	}
	if !zr.File[0].Mode().IsDir() {
		return "", errors.New("expecting base directory at the top of zip archive")
	}
	baseDir := zr.File[0].Name

	for _, zf := range zr.File {
		if !strings.HasPrefix(zf.Name, baseDir) {
			return baseDir, fmt.Errorf("file %q in zip archive not in base directory %q", zf.Name, baseDir)
		}

		if zf.Mode().IsDir() {
			if err := os.Mkdir(zf.Name, zf.Mode()); err != nil {
				return baseDir, err
			}
			continue
		}
		f, err := os.OpenFile(zf.Name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, zf.Mode())
		if err != nil {
			return baseDir, err
		}
		if zf.UncompressedSize == 0 {
			continue
		}

		zr, err := zf.Open()
		if err != nil {
			return baseDir, err
		}

		_, err = io.Copy(f, zr)
		if err != nil {
			return baseDir, err
		}

		if err := zr.Close(); err != nil {
			return baseDir, err
		}
		if err := f.Close(); err != nil {
			return baseDir, err
		}
	}

	return baseDir, nil
}

func patchSource(baseDir, src, dst string, extensions ...string) error {
	srcFile, err := os.Open(filepath.Join(baseDir, src))
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.WriteString(dstFile, "#ifndef USE_LIBSQLITE3\n")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(srcFile)
	for scanner.Scan() {
		text := scanner.Text()
		if text == `#include "sqlite3.h"` {
			text = `#include "sqlite3-binding.h"`
		}
		_, err = fmt.Fprintln(dstFile, text)
		if err != nil {
			break
		}
	}
	err = scanner.Err()
	if err != nil {
		return err
	}
	_, err = io.WriteString(dstFile, "#else // USE_LIBSQLITE3\n// If users really want to link against the system sqlite3 we\n// need to make this file a noop.\n#endif\n")
	if err != nil {
		return err
	}

	for _, ext := range extensions {
		ext = filepath.FromSlash(ext)
		fmt.Printf("Merging: %s into %s\n", ext, dst)

		extFile, err := os.Open(filepath.Join(baseDir, ext))
		if err != nil {
			return err
		}
		_, err = io.Copy(dstFile, extFile)
		extFile.Close()
		if err != nil {
			return err
		}
	}

	if err := dstFile.Close(); err != nil {
		return err
	}

	fmt.Printf("Patched: %s -> %s\n", src, dst)

	return nil
}
