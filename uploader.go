package main

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Uploader struct {
	S3       *s3.S3
	CacheDir string
	S3Bucket string
	S3Prefix string
}

func (s *S3Uploader) UploadLoop(queue chan string) {
	for {
		select {
		case file := <-queue:
			log.Println("Uploading ", file)
			s.uploadToS3(file)
		}
	}
}

// Walk path, upload what's still there.
func (s *S3Uploader) Reconcile() {
	filepath.Walk(s.CacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Walking error", err)
			return err
		}

		path = path[len(s.CacheDir):]

		if !info.IsDir() && info.ModTime().Before(time.Now().Add(-10*time.Minute)) {
			s.uploadToS3(path)
		}

		return nil
	})
}

func (s *S3Uploader) uploadToS3(path string) {
	fullPath := filepath.Join(s.CacheDir, path)
	file, err := os.Open(fullPath)
	if err != nil {
		log.Println("Failed to open file", err)
		return
	}

	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)

	_, err = s.S3.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.S3Bucket),
		Key:         aws.String(s.S3Prefix + "/" + path),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Println("Failed to upload file to S3", err)
		return
	}

	err = os.Remove(fullPath)
	if err != nil {
		log.Println("Failed to remove file", err)
		return
	}

	log.Println("File ", path, " uploaded to S3")
}
