DATA ?= $(CURDIR)/data


.PHONY: test
test:
	go test ./...


# for available downloads see: https://cldr.unicode.org/index/downloads
.PHONY: download-data
download-data: CLDR_VERSION ?= 44
download-data: CLDR_DATA_URL = https://unicode.org/Public/cldr/$(CLDR_VERSION)/cldr-common-$(CLDR_VERSION).0.zip
download-data:
	mkdir -p $(DATA)
	rm -rf $(DATA)/*
	curl -L -o $(DATA)/schema.mprot https://raw.githubusercontent.com/liblxn/lxn/main/schema.mprot
	curl -L -o $(DATA)/cldr.zip $(CLDR_DATA_URL)
	unzip -d $(DATA)/cldr $(DATA)/cldr.zip
	echo "$(CLDR_VERSION)" > $(DATA)/cldr/version
	rm $(DATA)/cldr.zip


.PHONY: generate
generate:
	go build -o bin/generate ./cmd/generate/
	rm -rf locale/*
	./bin/generate -out ./locale -cldr-data $(DATA)/cldr -cldr-version $(shell cat $(DATA)/cldr/version)
	./bin/generate -out ./lxn -schema $(DATA)/schema.mprot
