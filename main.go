// An example FTP server build on top of go-raval. graval handles the details
// of the FTP protocol, we just provide a basic in-memory persistence driver.
//
// If you're looking to create a custom graval driver, this example is a
// reasonable starting point. I suggest copying this file and changing the
// function bodies as required.
//
// USAGE:
//
//    go get github.com/yob/graval
//    go install github.com/yob/graval/graval-mem
//    ./bin/graval-mem
//
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/yob/graval"
)

const (
	fileOne = "This is the first file available for download.\n\nBy Jàmes"
	fileTwo = "This is file number two.\n\n2012-12-04"
)

// A minimal driver for graval that stores everything in memory. The authentication
// details are fixed and the user is unable to upload, delete or rename any files.
//
// This really just exists as a minimal demonstration of the interface graval
// drivers are required to implement.
type S3Driver struct {
	CacheDir    string
	UploadQueue chan string
}

func (driver *S3Driver) Authenticate(user string, pass string) bool {
	return user == "test" && pass == "1234"
}
func (driver *S3Driver) Bytes(path string) (bytes int) {
	return -1
}
func (driver *S3Driver) ModifiedTime(path string) (time.Time, error) {
	return time.Now(), nil
}
func (driver *S3Driver) ChangeDir(path string) bool {
	return true
}
func (driver *S3Driver) DirContents(path string) (files []os.FileInfo) {
	return []os.FileInfo{}
}

func (driver *S3Driver) DeleteDir(path string) bool {
	return false
}
func (driver *S3Driver) DeleteFile(path string) bool {
	return false
}
func (driver *S3Driver) Rename(fromPath string, toPath string) bool {
	return false
}
func (driver *S3Driver) MakeDir(path string) bool {
	return false
}
func (driver *S3Driver) GetFile(path string) (data string, err error) {
	return "", errors.New("GET not supported")
}
func (driver *S3Driver) PutFile(destPath string, data io.Reader) bool {

	// Write the file to the temp dir
	fileName := filepath.Base(destPath)
	filePath := filepath.Dir(destPath)
	outDir := filepath.Join(driver.CacheDir, filePath)
	outFile := filepath.Join(outDir, fileName)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Printf("Error making out dir %s\b", err)
		return false
	}

	f, err := os.Create(outFile)
	if err != nil {
		log.Printf("Error creating out file %s\n", err)
		return false
	}

	defer func() { f.Close() }()

	log.Printf("Writing to %s\n", outFile)

	if _, err := io.Copy(f, data); err != nil {
		log.Printf("Error copying to file %s\n", err)
		return false
	}

	// Message the s3 uploader to move it to S3.
	select {
	case driver.UploadQueue <- destPath:
	default:
		log.Printf("Queue full, skipping upload of %s\n", destPath)
	}

	return true
}

// graval requires a factory that will create a new driver instance for each
// client connection. Generally the factory will be fairly minimal. This is
// a good place to read any required config for your driver.
type S3DriverFactory struct {
	CacheDir    string
	UploadQueue chan string
}

func (factory *S3DriverFactory) NewDriver() (graval.FTPDriver, error) {
	return &S3Driver{
		CacheDir:    factory.CacheDir,
		UploadQueue: factory.UploadQueue,
	}, nil
}

// it's alive!
func main() {
	cacheDir := os.Getenv("FTP2S3_CACHE_DIR")
	if cacheDir == "" {
		fmt.Println("Set FTP2S3_CACHE_DIR")
		os.Exit(1)
	}

	bucket := os.Getenv("FTP2S3_BUCKET")
	if bucket == "" {
		fmt.Println("Set FTP2S3_BUCKET")
		os.Exit(1)
	}

	prefix := os.Getenv("FTP2S3_PREFIX")
	if prefix == "" {
		fmt.Println("Set FTP2S3_PREFIX to the prefix you want to store in on S3")
		os.Exit(1)
	}

	fileQueue := make(chan string, 10000)

	sess := session.New(&aws.Config{Region: aws.String("us-east-1")})
	svc := s3.New(sess)

	uploader := &S3Uploader{
		S3:       svc,
		CacheDir: cacheDir,
		S3Bucket: bucket,
		S3Prefix: prefix,
	}

	go uploader.UploadLoop(fileQueue)

	factory := &S3DriverFactory{
		CacheDir:    cacheDir,
		UploadQueue: fileQueue,
	}

	ftpServer := graval.NewFTPServer(&graval.FTPServerOpts{
		Factory: factory,
		//Hostname: "0.0.0.0",
		Port: 2121,
	})
	err := ftpServer.ListenAndServe()
	if err != nil {
		log.Print(err)
		log.Fatal("Error starting server!")
	}
}
