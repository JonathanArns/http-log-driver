ID := $(shell docker create http-log-driver true)
build:
	go build
	docker build -t http-log-driver .
	mkdir rootfs
	docker export $(ID) | tar -x -C rootfs/
	docker plugin create jonathanarns/http-log-driver .
	sudo rm -r rootfs
enable: build
	docker plugin enable jonathanarns/http-log-driver
