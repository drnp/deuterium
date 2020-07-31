PROJECT=Fusion.Skeletons.Deuterium
PREFIX=$(shell pwd)
VERSION=$(shell git describe --match 'v[0-9]*'  --always)
BRANCH=dev

ifndef OS
	OS=linux
endif

ifndef ARCH
	ARCH=amd64
endif

ifdef CI_COMMIT_REF_SLUG
	BRANCH=$(CI_COMMIT_REF_SLUG)
endif

ifndef DEPLOY_REPLICA
	DEPLOY_REPLICA=1
endif

ifndef GO
	GO=/usr/bin/go
endif

ifndef GOFMT
	GOFMT=/usr/bin/gofmt
endif

SOURCE_DIR=$(PREFIX)
BINARY_DIR=$(PREFIX)/bin
INSTALL_DIR=/opt/fusion

BINARY_MAIN=skel_main
BINARY_NODE=skel_node
SOURCE_MAIN_DIR=$(PREFIX)/instances/main
SOURCE_NODE_DIR=$(PREFIX)/instances/node

.PHONY: all skel_main skel_node clean install uninstall test doc dockerfile stackfile deploy docker_build
.DEFAULT: all

# Targets
all: summary fmt skel_main skel_node

summary:
	@echo -e "\033[1;37m  == \033[1;32m$(PROJECT) \033[1;33m$(VERSION) \033[1;37m==\033[0m"
	@echo -e "    Platform : \033[1;37m$(shell uname -sr)\033[0m"
	@echo -e "    Go       : \033[1;37m`$(GO) version`\033[0m"
	@echo -e "    Git      : \033[1;37m$(shell git version)\033[0m"
	@echo

fmt:
	@echo -e "\033[1;36m  Gofmt - Code syntax & format check\033[0m"
	@test -z "$$($(GOFMT) -s -l ${SOURCE_DIR} 2>&1 | tee /dev/stderr)" || \
		(echo >&2 " - Format check failed" && false)
	@echo -e "\033[1;32m    ... Passed\033[0m"
	@echo

skel_main:
	@echo -e "\033[1;36m  Compiling $(BINARY_MAIN) ...\033[0m"
	@mkdir -p $(BINARY_DIR)
	@echo -e "    \033[1;34mTarget : \033[1;35m$(BINARY_DIR)/$(BINARY_MAIN)\033[0m"
	@GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 $(GO) build -a -ldflags '-extldflags "-static"' -o $(BINARY_DIR)/$(BINARY_MAIN) $(SOURCE_MAIN_DIR)
	@echo

skel_node:
	@echo -e "\033[1;36m  Compiling $(BINARY_NODE) ...\033[0m"
	@mkdir -p $(BINARY_DIR)
	@echo -e "    \033[1;34mTarget : \033[1;35m$(BINARY_DIR)/$(BINARY_NODE)\033[0m"
	@GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 $(GO) build -a -ldflags '-extldflags "-static"' -o $(BINARY_DIR)/$(BINARY_NODE) $(SOURCE_NODE_DIR)
	@echo

clean:
	@rm -rf $(BINARY_DIR)

install:
	@mkdir -p ${INSTALL_DIR}/bin
	@install ${BINARY_DIR}/${BINARY_MAIN} ${INSTALL_DIR}/bin
	@install ${BINARY_DIR}/${BINARY_NODE} ${INSTALL_DIR}/bin
	# Echo some informations
	@echo

uninstall:
	@rm -rf ${INSTALL_DIR}
	@echo

test:
	@$(GO) clean --testcache
	@GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 $(GO) test ${SOURCE_DIR} --cover -v

doc:

docker_build:
	# Docker build
	@docker build --rm -t fusion/deuterium:$(BRANCH) ${SOURCE_DIR}
