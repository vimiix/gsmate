# Copyright 2024 Qian Yao
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO      := GO111MODULE=on CGO_ENABLED=0 go
_COMMIT := $(shell git describe --no-match --always --dirty)
COMMIT  := $(if $(COMMIT),$(COMMIT),$(_COMMIT))
BUILDDATE  := $(shell date '+%Y-%m-%dT%H:%M:%S')
REPO    := gsmate
LDFLAGS += -X "$(REPO)/pkg/version.Commit=$(COMMIT)"
LDFLAGS += -X "$(REPO)/pkg/version.BuildDate=$(BUILDDATE)"
FILES   := $$(find . -name "*.go")

.PHONY: help
help: ## print help info
	@printf "%-30s %s\n" "Target" "Description"
	@printf "%-30s %s\n" "------" "-----------"
	@grep -E '^[ a-zA-Z1-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## run go mod tidy
	@echo "tidy mod ..."
	@$(GO) mod tidy

.PHONY: fmt
fmt: ## format source code
	@echo "format code ..."
	@gofmt -s -l -w $(FILES) 2>&1

.PHONY: build
build: fmt tidy ## build binary
	@echo "building ..."
	@$(GO) build -ldflags "$(LDFLAGS)" -o bin/gsmate ./cmd
	@echo "build success, you can run ./bin/gsmate"

