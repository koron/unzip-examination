package main

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	err := run("test.zip")
	if err != nil {
		log.Fatal(err)
	}
}

func run(name string) error {
	r, err := zip.OpenReader(name)
	if err != nil {
		return err
	}
	defer r.Close()

	var (
		total = len(r.File)
		last  int
		start = time.Now()
	)
	log.Printf("%d files are in %s", total, name)
	for i, f := range r.File {
		if f.Mode().IsDir() {
			continue
		}
		fr, err := f.Open()
		if err != nil {
			return err
		}
		fw := ioutil.Discard
		_, err = io.Copy(fw, fr)
		fr.Close()
		if err != nil {
			return err
		}

		curr := ((i + 1) * 10 / total) * 10
		if curr > last {
			last = curr
			log.Printf("progress %d%%", curr)
		}
	}
	log.Printf("done in %s", time.Since(start))
	return nil
}
