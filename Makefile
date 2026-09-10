PROJNAME = smfix

ifeq "$(GITHUB_REF_NAME)" ""
VERSION := -X 'main.Version=$(shell git rev-parse --short HEAD)'
else
VERSION := -X 'main.Version=$(GITHUB_REF_NAME)'
endif
FLAGS = -ldflags="-w -s $(VERSION)"
CMD = go build -trimpath $(FLAGS)
DIST = dist/

.PHONY: darwin-arm64 darwin-amd64 linux-amd64 linux-arm7 linux-arm6 win64 win32 \
	all all-zip clean test benchmark pprof pprof-http-mem

darwin-arm64:
	GOOS=darwin GOARCH=arm64 \
	$(CMD) -o $(DIST)$(PROJNAME)-$@

darwin-amd64:
	GOOS=darwin GOARCH=amd64 \
	$(CMD) -o $(DIST)$(PROJNAME)-$@

linux-amd64:
	GOOS=linux GOARCH=amd64 \
	$(CMD) -o $(DIST)$(PROJNAME)-$@

linux-arm7:
	GOOS=linux GOARCH=arm GOARM=7 \
	$(CMD) -o $(DIST)$(PROJNAME)-$@

linux-arm6:
	GOOS=linux GOARCH=arm GOARM=6 \
	$(CMD) -o $(DIST)$(PROJNAME)-$@

win64:
	GOOS=windows GOARCH=amd64 \
	$(CMD) -o $(DIST)$(PROJNAME)-$@.exe

win32:
	GOOS=windows GOARCH=386 \
	$(CMD) -o $(DIST)$(PROJNAME)-$@.exe

all: darwin-arm64 darwin-amd64 linux-amd64 linux-arm7 linux-arm6 win64 win32
	@true

all-zip: all
	for p in darwin-arm64 darwin-amd64 linux-amd64 linux-arm7 linux-arm6 win64 win32; do \
		zip -j $(DIST)$(PROJNAME)-$$p.zip $(DIST)$(PROJNAME)-$$p* README.md LICENSE; \
	done

clean:
	rm -f $(DIST)$(PROJNAME)-*

# both modules: fix/ is a separate module, ./... from root never includes it
test:
	go test ./...
	cd fix && go test ./...

pprof:
	GOOS=darwin GOARCH=arm64 \
	$(CMD) -tags pprof -o $(DIST)$(PROJNAME)-pprof

benchmark:
	go test -bench=. -benchmem -run=NONE ./...
	cd fix && go test -bench=. -benchmem -run=NONE ./...

# profiles are written to the current dir by the pprof build (see pprof_enabled.go)
pprof-http-mem:
	go tool pprof -http=:8080 mem.pprof
