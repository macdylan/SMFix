PROJNAME = smfix
LDFLAGS = -w -s
CMD = go build -ldflags="$(LDFLAGS)"
DIST = dist/

darwin-arm64: smfix.go
	GOOS=darwin GOARCH=arm64 \
		 $(CMD) -o $(DIST)$(PROJNAME)-$@ $^

darwin-amd64: smfix.go
	GOOS=darwin GOARCH=amd64 \
		 $(CMD) -o $(DIST)$(PROJNAME)-$@ $^

linux-amd64: smfix.go
	GOOS=linux GOARCH=amd64 \
		 $(CMD) -o $(DIST)$(PROJNAME)-$@ $^

linux-arm7: smfix.go
	GOOS=linux GOARCH=arm GOARM=7 \
		 $(CMD) -o $(DIST)$(PROJNAME)-$@ $^

linux-arm6: smfix.go
	GOOS=linux GOARCH=arm GOARM=6 \
		 $(CMD) -o $(DIST)$(PROJNAME)-$@ $^

win64: smfix.go
	GOOS=windows GOARCH=amd64 \
		 $(CMD) -o $(DIST)$(PROJNAME)-$@ $^

win32: smfix.go
	GOOS=windows GOARCH=386 \
		 $(CMD) -o $(DIST)$(PROJNAME)-$@ $^

all: darwin-arm64 darwin-amd64 linux-amd64 linux-arm7 linux-arm6 win64 win32
	@true

clean:
	rm -f $(DIST)$(PROJNAME)-*

