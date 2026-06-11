.PHONY: all clean

GO := bin/vendor/go/bin/go
SRC := src
DIST := dist

TARGETS := \
	$(DIST)/remdev-linux-amd64 \
	$(DIST)/remdev-linux-arm64 \
	$(DIST)/remdev-darwin-amd64 \
	$(DIST)/remdev-darwin-arm64 \
	$(DIST)/remdev-windows-amd64.exe \
	$(DIST)/remdev-windows-arm64.exe

all: $(TARGETS)

$(DIST)/remdev-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -C $(SRC) -o ../$@ .

$(DIST)/remdev-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -C $(SRC) -o ../$@ .

$(DIST)/remdev-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -C $(SRC) -o ../$@ .

$(DIST)/remdev-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -C $(SRC) -o ../$@ .

$(DIST)/remdev-windows-amd64.exe:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -C $(SRC) -o ../$@ .

$(DIST)/remdev-windows-arm64.exe:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GO) build -C $(SRC) -o ../$@ .

clean:
	rm -rf $(DIST)/*
