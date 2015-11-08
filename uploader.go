package main

import (
	"log"
	"os"
	"path/filepath"

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

func (s *S3Uploader) uploadToS3(path string) {
	fullPath := filepath.Join(s.CacheDir, path)
	file, err := os.Open(fullPath)
	if err != nil {
		log.Fatal("Failed to open file", err)
		return
	}

	_, err = s.S3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.S3Bucket),
		Key:    aws.String(s.S3Prefix + "/" + path),
		Body:   file,
	})
	if err != nil {
		log.Fatal("Failed to upload file to S3", err)
		return
	}

	err = os.Remove(fullPath)
	if err != nil {
		log.Fatal("Failed to remove file", err)
		return
	}

	log.Println("File ", path, " uploaded to S3")
}
