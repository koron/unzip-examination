package main

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"log"
	"sync"
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
		start = time.Now()
		wg    sync.WaitGroup
		n     int
		last  int
		mu    sync.Mutex
	)
	log.Printf("%d files are in %s", total, name)
	for i, f := range r.File {
		wg.Add(1)
		go func(i int, f *zip.File) {
			defer func() {
				mu.Lock()
				n++
				curr := (n * 10 / total) * 10
				if curr > last {
					last = curr
					log.Printf("progress %d%%", curr)
				}
				mu.Unlock()
				wg.Done()
			}()
			if f.Mode().IsDir() {
				return
			}
			fr, err := f.Open()
			if err != nil {
				log.Printf("WARN: zip.File.Open() failed: %s", err)
				return
			}
			fw := ioutil.Discard
			_, err = io.Copy(fw, fr)
			fr.Close()
			if err != nil {
				log.Printf("WARN: io.Copy() failed: %s", err)
				return
			}
		}(i, f)
	}
	wg.Wait()
	log.Printf("done in %s", time.Since(start))
	return nil
}
