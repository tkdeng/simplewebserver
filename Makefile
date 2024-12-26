build all:
	make dependencies
	go mod tidy

deps dependencies:
ifeq (,$(wildcard $(/usr/bin/dnf)))
	sudo dnf install pcre-devel
	sudo dnf install go
else ifeq (,$(wildcard $(/usr/bin/apt)))
	sudo apt install libpcre3-dev
else ifeq (,$(wildcard $(/usr/bin/yum)))
	sudo yum install pcre-dev
endif
