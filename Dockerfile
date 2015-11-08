FROM golang:1.5.1

ADD . /go/src/github.com/lstoll/ftp2s3

RUN cd /go/src/github.com/lstoll/ftp2s3 && go get .
