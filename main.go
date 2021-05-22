package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/kjk/notionapi"
)

var (
	flagToken  string
	flagPages  string
	flagOutput string
)

// allowedExtensions of files present in an existing output directory
var allowedExtensions = map[string]struct{}{
	".csv": {},
	".md":  {},
	".png": {},
}

func main() {
	flag.StringVar(&flagToken, "token", "", "value of the token_v2 cookie, required if the page is not public")
	flag.StringVar(&flagOutput, "output", "", "[required] directory to sync the data to; note the existing files will be deleted")
	flag.StringVar(&flagPages, "pages", "", "[required] comma-separated list of page to export the data from")
	flag.Parse()

	if flagPages == "" {
		log.Fatalf("Missing required flag -pages")
	}

	if flagOutput == "" {
		log.Fatalf("Missing required flag -output")
	}

	if err := export(flagPages, flagOutput, flagToken); err != nil {
		log.Fatal(err)
	}
}

func export(pages, output, token string) error {
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		// output dir already exists, let's do some extra verification to ensure we don't nuke something unexpected
		if err := verifyDir(output); err != nil {
			return err
		}
	}

	client := &notionapi.Client{}
	if token != "" {
		client.AuthToken = token
	}

	// create a temp directory to save and unzip the export
	staging, err := ioutil.TempDir("", "notion-exporter")
	if err != nil {
		return fmt.Errorf("can't create the directory to download the export to: %w", err)

	}
	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			log.Printf("Can't remote the staging directory %s: %v", path, err)
		}
	}(staging)

	// this blocks until ready
	for _, page := range strings.Split(pages, ",") {
		buf, err := client.ExportPages(page, notionapi.ExportTypeMarkdown, true)
		if err != nil {
			return fmt.Errorf("can't export page: %w", err)
		}

		if err := unzip(buf, staging); err != nil {
			return fmt.Errorf("can't unzip the archive to %s: %w", staging, err)
		}
	}

	// rsync from staging to target
	rsync := exec.Command("rsync", "-aq", "--delete", fmt.Sprintf("%s/", staging), fmt.Sprintf("%s/", output))
	return rsync.Run()
}

// verifyDir verifies if the given
func verifyDir(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("can't validate the target directory %s: %w", dir, err)
	}
	for _, f := range files {
		if f.IsDir() {
			if err := verifyDir(path.Join(dir, f.Name())); err != nil {
				return err
			}
			continue
		}
		if _, exists := allowedExtensions[path.Ext(f.Name())]; !exists {
			return fmt.Errorf(
				"💥 target directory contains a unknown file %s, please provide a valid target directory",
				path.Join(dir, f.Name()),
			)
		}
	}
	return nil
}

func unzip(input []byte, dir string) error {
	reader, err := zip.NewReader(bytes.NewReader(input), int64(len(input)))
	if err != nil {
		return err
	}

	for _, zipFileReader := range reader.File {
		if err := extractFileOrDir(dir, zipFileReader, err); err != nil {
			return err
		}
	}

	return nil
}

func extractFileOrDir(dir string, zipFileReader *zip.File, err error) error {
	target := path.Join(dir, zipFileReader.Name)

	if zipFileReader.FileInfo().IsDir() {
		return os.MkdirAll(target, zipFileReader.Mode())
	}

	zipFile, err := zipFileReader.Open()
	if err != nil {
		return err
	}
	defer zipFile.Close()

	if err := os.MkdirAll(path.Dir(target), 0755); err != nil {
		return err
	}

	diskFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFileReader.Mode())
	if err != nil {
		return err
	}
	defer diskFile.Close()

	if _, err := io.Copy(diskFile, zipFile); err != nil {
		return err
	}
	return nil
}
