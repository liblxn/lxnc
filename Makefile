DATA_DIR ?= $(CURDIR)/data

# for available downloads see: https://cldr.unicode.org/index/downloads
CLDR_VERSION  ?= 44.0
CLDR_DATA_URL = https://unicode.org/Public/cldr/40/cldr-common-$(CLDR_VERSION).zip


.PHONY: download-data
download-data:
	mkdir -p $(DATA_DIR)
	curl -L -o $(DATA_DIR)/schema.mprot https://raw.githubusercontent.com/liblxn/lxn/master/schema.mprot
	curl -L -o $(DATA_DIR)/cldr.zip $(CLDR_DATA_URL)
	unzip -d $(DATA_DIR)/cldr $(DATA_DIR)/cldr.zip
	rm $(DATA_DIR)/cldr.zip


.PHONY: generate
generate:
	go build -o bin/data-gen ./cmd/generate/
	rm -rf internal/locale/*
	./bin/generate -out ./internal/locale -cldr-data $(DATA_DIR)/cldr -cldr-version $(CLDR_VERSION)
	./bin/generate -out ./internal/schema -schema $(DATA_DIR)/schema.mprot
