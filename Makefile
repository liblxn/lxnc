# for available downloads see: https://cldr.unicode.org/index/downloads
CLDR_VERSION  ?= 40.0
CLDR_DATA_URL ?= https://unicode.org/Public/cldr/40/cldr-common-$(CLDR_VERSION).zip
CLDR_DATA_DIR ?= data/cldr


.PHONY: download-cldr
download-cldr:
	mkdir -p $(CLDR_DATA_DIR)
	curl -L -o $(CLDR_DATA_DIR)/cldr.zip $(CLDR_DATA_URL)
	unzip -d $(CLDR_DATA_DIR) $(CLDR_DATA_DIR)/cldr.zip
	rm $(CLDR_DATA_DIR)/cldr.zip

.PHONY: generate-locale
generate-locale:
	go build -o bin/data-gen ./cmd/data-gen/
	rm -rf internal/locale/*
	./bin/data-gen -cldr-data $(CLDR_DATA_DIR) -cldr-version $(CLDR_VERSION) -out ./internal/locale -pkg locale
