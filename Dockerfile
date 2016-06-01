FROM golang:onbuild
MAINTAINER docker@mailserver.1n3t.de

RUN go get github.com/muhproductions/muh-api

EXPOSE 8080
CMD COMPRESSION=snappy muh-api
