DATA_DIR ?= $(CURDIR)/data


.PHONY: test
test:
	go test ./...


# for available downloads see: https://cldr.unicode.org/index/downloads
.PHONY: download-data
download-data: CLDR_VERSION ?= 44
download-data: CLDR_DATA_URL = https://unicode.org/Public/cldr/$(CLDR_VERSION)/cldr-common-$(CLDR_VERSION).0.zip
download-data:
	mkdir -p $(DATA_DIR)
	rm -rf $(DATA_DIR)/*
	curl -L -o $(DATA_DIR)/schema.mprot https://raw.githubusercontent.com/liblxn/lxn/master/schema.mprot
	curl -L -o $(DATA_DIR)/cldr.zip $(CLDR_DATA_URL)
	unzip -d $(DATA_DIR)/cldr $(DATA_DIR)/cldr.zip
	echo "$(CLDR_VERSION)" > $(DATA_DIR)/cldr/version
	rm $(DATA_DIR)/cldr.zip


.PHONY: generate
generate:
	go build -o bin/generate ./cmd/generate/
	rm -rf internal/locale/* ./schema/*
	./bin/generate -out ./internal/locale -cldr-data $(DATA_DIR)/cldr -cldr-version $(shell cat $(DATA_DIR)/cldr/version)
	./bin/generate -out ./schema -schema $(DATA_DIR)/schema.mprot
