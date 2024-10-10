OS := $(shell uname -s)
ARCH := $(shell uname -m)
NAME := "miccedu_downloader"

GO_BIN := go
OUTPUT_DIR := ./bin

ifeq ($(OS),Windows_NT)
    TARGET := $(OUTPUT_DIR)/$(NAME)_$(OS)_$(ARCH).exe
else
    TARGET := $(OUTPUT_DIR)/$(NAME)_$(OS)_$(ARCH)
endif

.PHONY: all build clean

all: build

build:
	@echo "Detected OS: $(OS), Arch: $(ARCH)"
	@echo "Building for target: $(TARGET)"
	@mkdir -p $(OUTPUT_DIR)
	$(GO_BIN) build -v -o $(TARGET)

clean:
	@echo "Cleaning up..."
	rm -rf $(OUTPUT_DIR)

build-cross:
	@echo "Building cross-platform for GOOS=$(GOOS) and GOARCH=$(GOARCH)"
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_BIN) build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/myapp-$(GOOS)-$(GOARCH)
