package main

import (
	"archive/zip"
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

var (
	parallelism int
	outdir      string
	writeToFile bool
	useTruncate bool
)

func main() {
	defaultOutdir := filepath.Join("outdir", strconv.FormatInt(time.Now().Unix(), 10))
	flag.IntVar(&parallelism, "p", runtime.NumCPU(), `degree of parallelism`)
	flag.StringVar(&outdir, "o", defaultOutdir, `name of output dir`)
	flag.BoolVar(&writeToFile, "f", false, `write to real file`)
	flag.BoolVar(&useTruncate, "t", false, `truncate before write contents`)
	flag.Parse()
	err := run("test.zip", outdir)
	if err != nil {
		log.Fatal(err)
	}
}

type progress struct {
	total int
	last  int
	curr  int

	l sync.Mutex
}

func newProgress(total int) *progress {
	return &progress{
		total: total,
	}
}

func (p *progress) done() {
	p.l.Lock()
	p.curr++
	q := (p.curr * 10 / p.total) * 10
	if q > p.last {
		p.last = q
		log.Printf("progress %d%%", q)
	}
	p.l.Unlock()
}

func prepareFile(outdir string, zf *zip.File) (*os.File, error) {
	name := filepath.Join(outdir, zf.Name)
	dir := filepath.Dir(name)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, err
	}
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	if useTruncate {
		err := f.Truncate(int64(zf.UncompressedSize64))
		if err != nil {
			f.Close()
			os.Remove(name)
			return nil, err
		}
	}
	return f, nil
}

func run(name string, outdir string) error {
	r, err := zip.OpenReader(name)
	if err != nil {
		return err
	}
	defer r.Close()

	var (
		total = len(r.File)
		start = time.Now()
		wg    sync.WaitGroup
		sem   *semaphore.Weighted
		prog  = newProgress(total)
	)
	if parallelism > 0 {
		sem = semaphore.NewWeighted(int64(parallelism))
	}

	log.Printf("%d files are in %s", total, name)
	for i, zf := range r.File {
		wg.Add(1)
		go func(i int, zf *zip.File) {
			defer wg.Done()
			defer prog.done()
			if sem != nil {
				err := sem.Acquire(context.Background(), 1)
				if err != nil {
					log.Printf("INFO: semaphore.Weighted.Acquire() failed: %s", err)
					return
				}
				defer sem.Release(1)
			}
			if zf.Mode().IsDir() {
				return
			}

			fr, err := zf.Open()
			if err != nil {
				log.Printf("WARN: zip.File.Open() failed: %s", err)
				return
			}
			defer fr.Close()

			var fw io.Writer = ioutil.Discard
			if writeToFile {
				f, err := prepareFile(outdir, zf)
				if err != nil {
					log.Printf("WARN: prepareFile() failed: %s", err)
					return
				}
				defer f.Close()
				fw = f
			}

			_, err = io.Copy(fw, fr)
			if err != nil {
				log.Printf("WARN: io.Copy() failed: %s", err)
				return
			}
		}(i, zf)
	}
	wg.Wait()
	log.Printf("done in %s", time.Since(start))
	log.Printf("output to dir %s", outdir)
	return nil
}
