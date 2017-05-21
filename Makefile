SRC_FILES=src/main.go src/heatmap.go
TEST_FILES=src/heatmap_test.go
VENDOR_DIR=vendor
BIN_NAME=heatmap

all: test build

test: dep $(TEST_FILES) $(SRC_FILES)
	go test $(TEST_FILES) $(SRC_FILES)

build: dep bin/$(BIN_NAME)
bin/$(BIN_NAME): $(SRC_FILES)
	go build -o bin/$(BIN_NAME) $(SRC_FILES)

dep: $(VENDOR_DIR)
$(VENDOR_DIR):
	dep ensure

clean:
	$(RM) -r bin
	$(RM) -r $(VENDOR_DIR)

.PHONY: build test clean
