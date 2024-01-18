PROJNAME = smfix

ifeq "$(GITHUB_REF_NAME)" ""
    VERSION := -X 'main.Version=$(shell git rev-parse --short HEAD)'
else
	VERSION := -X 'main.Version=$(GITHUB_REF_NAME)'
endif
FLAGS = -ldflags="-w -s $(VERSION)"
CMD = go build -trimpath $(FLAGS)
DIST = dist/

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

dep: # Get the dependencies
	go mod download

all: dep darwin-arm64 darwin-amd64 linux-amd64 linux-arm7 linux-arm6 win64 win32
	@true

all-zip: all
	for p in darwin-arm64 darwin-amd64 linux-amd64 linux-arm7 linux-arm6 win64 win32; do \
		zip -j $(DIST)$(PROJNAME)-$$p.zip $(DIST)$(PROJNAME)-$$p* README.md LICENSE; \
	done

clean:
	rm -f $(DIST)$(PROJNAME)-*

test:
	go test ./fix/...

pprof:
	GOOS=darwin GOARCH=arm64 \
		 $(CMD) -tags pprof -o $(DIST)$(PROJNAME)-pprof

benchmark:
	go test -bench=. -memprofile=dist/bench_mem.pprof ./fix/...

pprof-http-mem:
	go tool pprof -http=:8080 $(DIST)mem.pprof
