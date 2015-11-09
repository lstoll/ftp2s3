# ftp2s3

FTP Server that uploads to data. Not general purpose, designed to be the bare minimum to get data from D-Link cameras offsite.

## Using

There's a docker container `docker pull lstoll/ftp2s3`

Set up the following env:

* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`
* `FTP2S3_BUCKET`
* `FTP2S3_PREFIX` prefix inside the bucket to store uploads
* `FTP2S3_USERNAME` ftp username
* `FTP2S3_PASSWORD` ftp password
* `FTP2S3_PORT` (optional) port to listen on, default is 2121


Also set up your camera like this:

![Camera settings screenshot](https://cdn.lstoll.net/screen/D-Link_Corporation.__WIRELESS_INTERNET_CAMERA__SETUP__FTP_2015-11-08_19-13-05.png)