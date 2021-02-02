# --- Global -------------------------------------------------------------------
O = out
COVERAGE = 91
SEMVER ?= $(shell git describe --tags --dirty)
COMMIT_SHA ?= $(shell git rev-parse --short HEAD)

all: build test check-coverage lint  ## build, test, check coverage and lint
	@if [ -e .git/rebase-merge ]; then git --no-pager log -1 --pretty='%h %s'; fi
	@echo '$(COLOUR_GREEN)Success$(COLOUR_NORMAL)'

clean::  ## Remove generated files
	-rm -rf $(O)

.PHONY: all clean

# --- Build --------------------------------------------------------------------
GO_LDFLAGS = \
	-X main.Semver=$(SEMVER) \
	-X main.CommitSha=$(COMMIT_SHA)

build: | $(O)  ## Build binaries of directories in ./cmd to out/
	go build -o $(O) -ldflags='$(GO_LDFLAGS)' ./cmd/...

install:  ## Build and install binaries in $GOBIN or $GOPATH/bin
	go install -ldflags='$(GO_LDFLAGS)' ./cmd/...

run: build  ## Run foxtrot server
	$(O)/foxtrot

.PHONY: build install run

# --- Test ---------------------------------------------------------------------
COVERFILE = $(O)/coverage.txt

test: ## Run tests and generate a coverage file
	go test -coverprofile=$(COVERFILE) ./...

build-test: build  ## Run integration tests against a locally started foxtrot server
	$(O)/foxtrot & \
		pid=$$!; \
		go test ./pkg/foxtrot --api-base-url http://localhost:8080; \
		kill $$pid

check-coverage: test  ## Check that test coverage meets the required level
	@go tool cover -func=$(COVERFILE) | $(CHECK_COVERAGE) || $(FAIL_COVERAGE)

cover: test  ## Show test coverage in your browser
	go tool cover -html=$(COVERFILE)

CHECK_COVERAGE = awk -F '[ \t%]+' '/^total:/ {print; if ($$3 < $(COVERAGE)) exit 1}'
FAIL_COVERAGE = { echo '$(COLOUR_RED)FAIL - Coverage below $(COVERAGE)%$(COLOUR_NORMAL)'; exit 1; }

.PHONY: build-test check-coverage cover test

# --- Lint ---------------------------------------------------------------------
GOLINT_VERSION = 1.33.2
GOLINT_INSTALLED_VERSION = $(or $(word 4,$(shell golangci-lint --version 2>/dev/null)),0.0.0)
GOLINT_USE_INSTALLED = $(filter $(GOLINT_INSTALLED_VERSION),v$(GOLINT_VERSION) $(GOLINT_VERSION))
GOLINT = $(if $(GOLINT_USE_INSTALLED),golangci-lint,golangci-lint-v$(GOLINT_VERSION))

GOBIN ?= $(firstword $(subst :, ,$(GOPATH)))/bin

lint: $(if $(GOLINT_USE_INSTALLED),,$(GOBIN)/$(GOLINT))  ## Lint go source code
	$(GOLINT) run

$(GOBIN)/$(GOLINT):
	cd /tmp; \
	GOBIN=/tmp GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLINT_VERSION); \
	mv /tmp/golangci-lint $@

.PHONY: lint

# --- Docker -------------------------------------------------------------------
DOCKER_TAG ?= $(or $(DEV),$(error DOCKER_TAG not set))
DOCKER_TAGS = $(DOCKER_TAG) $(if $(filter true,$(DOCKER_PUSH_LATEST)),latest)
DOCKER_BUILD_ARGS = \
	--build-arg=SEMVER=$(SEMVER) \
	--build-arg=COMMIT_SHA=$(COMMIT_SHA)

docker-build:
	docker build $(DOCKER_BUILD_ARGS) --tag foxtrot:latest .

docker-build-release:
	docker buildx build $(DOCKER_BUILD_ARGS) \
		--push \
		$(foreach tag,$(DOCKER_TAGS),--tag foxygoat/foxtrot:$(tag) ) \
		--platform linux/amd64,linux/arm/v7 .

docker-run: docker-build
	docker run --rm -it -p8080:8080 foxtrot:latest

docker-test: docker-build
	docker run --rm --detach -p8083:8080 --name foxtrot-test foxtrot:latest
	go test ./pkg/foxtrot --api-base-url http://localhost:8083; \
		rc=$$?; \
		docker kill foxtrot-test; \
		exit $$rc

.PHONY: docker-build docker-build-release docker-run docker-test

# --- Deployment -------------------------------------------------------------------
LOCAL_MAIN = deployment/main.jsonnet
LOCAL_OVERLAY = deployment/$*/overlay.jsonnet

REMOTE_BASE = https://github.com/foxygoat/foxtrot/raw/$(COMMIT_SHA)
REMOTE_MAIN = $(REMOTE_BASE)/$(LOCAL_MAIN)
REMOTE_OVERLAY = $(REMOTE_BASE)/$(LOCAL_OVERLAY)

SOURCE = LOCAL
MAIN = $($(SOURCE)_MAIN)
OVERLAY = $($(SOURCE)_OVERLAY)

TLA_ARGS = \
	--tla-str docker_tag=$(DOCKER_TAG) \
	--tla-str commit_sha=$(COMMIT_SHA) \
	$(if $(DEV), --tla-str dev=$(DEV)) \
	--tla-code-file overlay=$(OVERLAY)

KUBECFG_DEPLOY = kubecfg update $(TLA_ARGS) $(MAIN)
KUBECFG_SHOW = kubecfg show $(TLA_ARGS) $(MAIN)
KUBECFG_DIFF = kubecfg diff --diff-strategy subset $(TLA_ARGS) $(MAIN)
KUBECFG_UNDEPLOY = kubecfg delete $(TLA_ARGS) $(MAIN)

deploy-%: | deployment/% deployment/%/secret.json deployment/%/overlay.jsonnet  ## Generate and deploy k8s manifests
	$(KUBECFG_DEPLOY)

show-deploy-%: | deployment/% deployment/%/secret.json deployment/%/overlay.jsonnet  ## Show k8s manifests that would be deployed
	$(KUBECFG_SHOW)

diff-deploy-%: ## Show diff of k8s manifests between files and deployed
	$(KUBECFG_DIFF)

undeploy-%:  ## Delete deployment
	$(KUBECFG_UNDEPLOY)

deployment/%:
	mkdir $@

deployment/%/secret.json:
	kubectl create secret generic foxtrot -n foxtrot --from-literal=authsecret=$$(openssl rand -hex 32) --dry-run=client -o yaml | kubeseal -w $@

deployment/%/overlay.jsonnet:
	@printf '{ manifest+: [$$.sealedSecret], \n config+: { hostname: null // add your hostname \n }, \n sealedSecret:: import "secret.json"}' | jsonnetfmt -o $@ -
	@printf '\ndefault overlay generated: %s \n' $@
	@printf 'review and run `make deploy-$*` again.\n\n'
	@exit 1

show-secret:  ## Show currently deployed foxtrot auth secret
	kubectl get secret -n foxtrot foxtrot -o go-template='{{.data.authsecret | base64decode}}{{"\n"}}'

.PRECIOUS: deployment/% deployment/%/secret.json deployment/%/overlay.jsonnet
.PHONY: show-secret

# --- JCDC --------------------------------------------------------------------
CURL_FLAGS = --silent --show-error --retry 3 --dump-header -
JCDC_PAYLOAD = { "command": "$(JCDC_COMMAND)", "apiKey": "$(JCDC_API_KEY)" }
JCDC_RUN = curl $(CURL_FLAGS) --data '$(JCDC_PAYLOAD)' '$(JCDC_URL)'

jcdc-deploy-%: SOURCE = REMOTE
jcdc-deploy-%: JCDC_COMMAND = $(KUBECFG_DEPLOY)
jcdc-deploy-%:
	$(JCDC_RUN)

jcdc-undeploy-%: SOURCE = REMOTE
jcdc-undeploy-%: JCDC_COMMAND = $(KUBECFG_UNDEPLOY)
jcdc-undeploy-%:
	$(JCDC_RUN)

# --- Utilities ----------------------------------------------------------------
COLOUR_NORMAL = $(shell tput sgr0 2>/dev/null)
COLOUR_RED    = $(shell tput setaf 1 2>/dev/null)
COLOUR_GREEN  = $(shell tput setaf 2 2>/dev/null)
COLOUR_WHITE  = $(shell tput setaf 7 2>/dev/null)

help:
	@awk -F ':.*## ' 'NF == 2 && $$1 ~ /^[A-Za-z0-9%_-]+$$/ { printf "$(COLOUR_WHITE)%-30s$(COLOUR_NORMAL)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

$(O):
	@mkdir -p $@

.PHONY: help
