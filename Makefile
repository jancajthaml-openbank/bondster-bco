
META := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null | sed 's:.*/::')
VERSION := $(shell git fetch --tags --force 2> /dev/null; tags=($$(git tag --sort=-v:refname)) && ([ $${\#tags[@]} -eq 0 ] && echo v0.0.0 || echo $${tags[0]}))

export COMPOSE_DOCKER_CLI_BUILD = 1
export DOCKER_BUILDKIT = 1
export COMPOSE_PROJECT_NAME = bondster-bco

.ONESHELL:
.PHONY: arm64
.PHONY: amd64
.PHONY: armhf

.PHONY: all
all: bootstrap sync test package bbtest

.PHONY: package
package:
	@$(MAKE) package-amd64
	@$(MAKE) bundle-docker

.PHONY: package-%
package-%: %
	@$(MAKE) bundle-binaries-$^
	@$(MAKE) bundle-debian-$^

.PHONY: bundle-binaries-%
bundle-binaries-%: %
	@docker-compose \
		run \
		--rm package \
		--arch linux/$^ \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-rest \
		--output /project/packaging/bin
	@docker-compose \
		run \
		--rm package \
		--arch linux/$^ \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-import \
		--output /project/packaging/bin

.PHONY: bundle-debian-%
bundle-debian-%: %
	@docker-compose \
		run \
		--rm debian-package \
		--version $(VERSION) \
		--arch $^ \
		--pkg bondster-bco \
		--source /project/packaging

.PHONY: bundle-docker
bundle-docker:
	@docker build \
		-t openbank/bondster-bco:$(VERSION)-$(META) \
		-f packaging/docker/Dockerfile \
		.

.PHONY: bootstrap
bootstrap:
	@docker-compose build --force-rm go

.PHONY: lint
lint:
	@docker-compose \
		run \
		--rm lint \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-rest \
	|| :
	@docker-compose \
		run \
		--rm lint \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-import \
	|| :

.PHONY: sec
sec:
	@docker-compose \
		run \
		--rm sec \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-rest \
	|| :
	@docker-compose \
		run \
		--rm sec \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-import \
	|| :

.PHONY: sync
sync:
	@docker-compose \
		run \
		--rm sync \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-rest
	@docker-compose \
		run \
		--rm sync \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-import

.PHONY: scan
scan:
	docker scan \
	  openbank/bondster-bco:$(VERSION)-$(META) \
	  --file ./packaging/docker/Dockerfile \
	  --exclude-base

.PHONY: test
test:
	@docker-compose \
		run \
		--rm test \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-rest \
		--output /project/reports/unit-tests
	@docker-compose \
		run \
		--rm test \
		--source /go/src/github.com/jancajthaml-openbank/bondster-bco-import \
		--output /project/reports/unit-tests

.PHONY: release
release:
	@docker-compose \
		run \
		--rm release \
		--version $(VERSION) \
		--token ${GITHUB_RELEASE_TOKEN}

.PHONY: bbtest
bbtest:
	@META=$(META) VERSION=$(VERSION) docker-compose up -d bbtest
	@docker exec -t $$(docker-compose ps -q bbtest) python3 /opt/app/bbtest/main.py
	@docker-compose down -v
