#!/bin/sh

output=smokeping

case "$1" in
	linux)
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $output *.go
		;;
	mac)
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $output *.go
    ;;
	win)
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $output.exe *.go
		;;
	windows)
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $output.exe *.go
		;;
	*)
    go build -o $output *.go
    ;;
esac


