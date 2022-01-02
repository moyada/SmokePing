#!/bin/sh


case "$1" in
	linux)
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
		;;
	mac)
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build main.go
    ;;
	win)
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go
		;;
	windows)
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go
		;;
	*)
    go build main.go
    ;;
esac


