
.PHONY: help
help: ## Help message
	@awk 'BEGIN{FS=":.*[#]#"}/:.*[#]#/{printf("%s # %s\n", $$1, $$2)}' $(MAKEFILE_LIST) | column -t -s#


.PHONY: docs
docs: html slides ## Generate docs

.PHONY: notes
notes: ## Build notes images
	plantuml -tsvg -oimg docs/notes/*

.PHONY: html
html: ## Build html site
	sphinx-build -b html docs build

.PHONY: slides
slides: ## Build slides
	sphinx-build -b revealjs docs/slides build/slides
