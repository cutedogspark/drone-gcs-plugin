package main

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type (
	Repo struct {
		Owner string
		Name  string
	}

	Build struct {
		Tag     string
		Event   string
		Number  int
		Commit  string
		Ref     string
		Branch  string
		Author  string
		Message string
		Status  string
		Link    string
		Started int64
		Created int64
	}

	Job struct {
		Started int64
	}

	Plugin struct {
		Repo
		Build
		Job

		Credentials string
		Bucket      string

		Access []string

		Source           string
		Target           string
		TargetDateFolder bool

		TriggerModule string
		TriggerEven   string
		TriggerBranch string
		TriggerSHA    string
		TriggerFolder string

		TagModule string
		TagFolder string

		StripPrefix string
		Exclude     []string
		Compress    []string
	}
)

// normalize target path
func (p *Plugin) normalizeTargetPath() {
	if strings.HasPrefix(p.Target, "/") {
		p.Target = p.Target[1:]
	}

	if strings.HasSuffix(p.Target, "/") {
		p.Target = p.Target[:len(p.Target)-1]
	}
	return
}

func (p *Plugin) detectionTarget() error {

	if p.TriggerEven != "" {
		log.Info("--- Sub Project --- ")
		if p.TriggerEven == "pull_request" {
			log.Info("Trigger Mode")
			p.Target = p.TriggerFolder
		} else if p.TriggerEven == "push" && p.TriggerBranch != "master" {
			log.Info("Trigger Mode")
			p.Target = p.TriggerFolder
		} else if p.TriggerEven == "tag" {
			log.Info("Tag Mode")
			p.Target = p.TagFolder
		} else {
			log.Info("Normal Mode")
			p.Target = p.Target
		}
	}else{
		log.Info("--- Main Project  --- ")
		if p.Event == "pull_request" {
			log.Info("Trigger Mode")
			p.Target = p.TriggerFolder
		} else if p.Event == "push" && p.Branch != "master" {
			log.Info("Trigger Mode")
			p.Target = p.TriggerFolder
		} else if p.Event == "tag" {
			log.Info("Tag Mode")
			p.Target = p.TagFolder
		} else {
			log.Info("Normal Mode")
			p.Target = p.Target
		}
	}


	return nil
}

func (p *Plugin) Exec() error {

	p.detectionTarget()

	p.normalizeTargetPath()

	log.Debug("Target folder => ", p.Target)

	if p.TargetDateFolder {
		p.Target = fmt.Sprintf("%s/%s/%s", p.Target, time.Now().Format("2006"), time.Now().Format("01-02"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := google.JWTConfigFromJSON([]byte(p.Credentials), storage.ScopeFullControl)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("google parse gcs key fail")
		return err
	}

	gcc, err := storage.NewClient(ctx, option.WithTokenSource(cfg.TokenSource(ctx)))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("google client fail")
		return err
	}
	defer gcc.Close()

	bkt := gcc.Bucket(p.Bucket)

	matches, err := LoopSrcPath(p.Source)
	if err != nil {
		log.Fatalf("Source Path : %v", err)
	}

	for _, match := range matches {
		target := strings.TrimPrefix(filepath.Join(p.Target, strings.TrimPrefix(match, p.StripPrefix)), "/")
		if err := p.uploadFile(ctx, bkt, match, target); err != nil {
			log.WithFields(log.Fields{
				"name":   match,
				"bucket": p.Bucket,
				"target": target,
				"error":  err,
			}).Error("Could not upload file")
			return err
		}
	}

	return nil
}

// uploadFile performs the actual uploading process.
func (p *Plugin) uploadFile(ctx context.Context, bkt *storage.BucketHandle, match, target string) error {

	// gcp has pretty crappy default content-type headers so this plugin
	// attempts to provide a proper content-type.
	content := contentType(match)

	// log file for debug purposes.
	log.WithFields(log.Fields{
		"name":         match,
		"bucket":       p.Bucket,
		"target":       target,
		"content-type": content,
	}).Debug("Uploading file")

	f, err := os.Open(match)
	if err != nil {
		return err
	}
	defer f.Close()

	obj := bkt.Object(target)

	var w io.WriteCloser = obj.NewWriter(ctx)

	if _, err := io.Copy(w, f); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	attrs := storage.ObjectAttrsToUpdate{
		ContentType: content,
	}

	_, err = obj.Update(ctx, attrs)
	if err != nil {
		return err
	}

	return nil
}

func contentType(path string) string {
	ext := filepath.Ext(path)
	typ := mime.TypeByExtension(ext)
	if typ == "" {
		typ = "application/octet-stream"
	}
	return typ
}

func stringInSlice(a string) bool {
	list := []string{".gitkeep"}
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func LoopSrcPath(src string) ([]string, error) {
	var items []string
	err := filepath.Walk(src, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		if !stringInSlice(fi.Name()) {
			items = append(items, p)
		}
		return nil
	})
	return items, err
}
