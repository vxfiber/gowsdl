GHACCOUNT := hooklift
NAME := gowsdl
VERSION := v0.5.0

include common.mk

run:
	go run ./cmd/gowsdl -d tmp -p spri -c VersionType,AnsprechpartnerType ./spri/wsdl/spri.wsdl

deps:
	go get github.com/c4milo/github-release
	go get github.com/mitchellh/gox
	go get github.com/hooklift/assert
