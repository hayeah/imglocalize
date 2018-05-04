package main

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"sync"
)

type capturedImages map[string]string

// addImage record an src to a file name. If collision, append a hash to make the name unique.
func (m capturedImages) addImage(imgSrc string) (string, error) {
	u, err := url.Parse(string(imgSrc))
	if err != nil {
		return "", err
	}

	filePath := path.Base(u.Path)

	_, found := m[filePath]
	if found {
		// append a hash to file name if collision filezilla6-[hash].png
		ext := path.Ext(filePath)
		h := fnv.New32a()
		h.Write([]byte(imgSrc))
		hashhex := hex.EncodeToString(h.Sum(nil))

		filePath = filePath[:len(filePath)-len(ext)] + "-" + hashhex + ext
	}

	m[imgSrc] = filePath

	return filePath, nil
}

func (m capturedImages) scanAndRewrite(re *regexp.Regexp, text string) string {
	result := reMarkdownImage.ReplaceAllStringFunc(text, func(match string) string {
		ii := reMarkdownImage.FindStringSubmatchIndex(match)

		// first group
		g1 := match[ii[2]:ii[3]]
		imgSrc := match[ii[4]:ii[5]]

		filePath, err := m.addImage(imgSrc)
		if err != nil {
			fmt.Println("cannot rename image:", imgSrc)
			return match
		}

		rewriteSrc := g1 + filePath

		return rewriteSrc
	})

	return result
}

func (m capturedImages) downloadAll(outputDir string) {
	images := m
	concurrency := 5
	workers := make(chan struct{}, concurrency)

	// download images
	var wg sync.WaitGroup
	wg.Add(len(images))
	for src, targetName := range images {
		src := src
		targetPath := path.Join(outputDir, targetName)
		_, err := os.Stat(targetPath)
		if !os.IsNotExist(err) {
			wg.Done()
			continue
		}

		workers <- struct{}{}
		go func() {
			fmt.Printf("downloading %s =>\n\t%s\n", src, targetPath)
			err := downloadFile(src, targetPath)
			if err != nil {
				log.Println("download failed:", src)
			}
			<-workers
			wg.Done()
		}()
	}

	wg.Wait()
}

func downloadFile(from, to string) (err error) {
	res, err := http.Get(from)
	if err != nil {
		return
	}
	defer res.Body.Close()

	f, err := os.Create(to)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)

	return
}
