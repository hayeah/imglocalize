// Scan document for images. If not local, download it and rewrite the file
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
)

var reImageTag = regexp.MustCompile(`(<img +src=")([^"]+)`)
var reMarkdownImage = regexp.MustCompile(`(!\[.*\]\()([^\)]+)`)

func main() {
	inputFile := os.Args[1]
	err := replaceAndDownloadImages(inputFile)
	if err != nil {
		log.Fatalln(err)
	}
}

func replaceAndDownloadImages(filePath string) (err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	fmt.Sprintln("find remote images")

	src := string(data)

	outputDir := path.Dir(filePath)

	// url => filename
	images := make(capturedImages)

	var result string

	result = images.scanAndRewrite(reMarkdownImage, src)
	result = images.scanAndRewrite(reImageTag, src)

	images.downloadAll(outputDir)

	// overwrite the original file
	err = ioutil.WriteFile(filePath, []byte(result), 0x644)
	if err != nil {
		return
	}

	return
}
