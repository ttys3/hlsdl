build_linux:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/hlsdl ./cmd/hlsdl
	@md5sum ./bin/hlsdl

build_osx:
	env GOOS=darwin GOARCH=amd64 go build -o ./bin/hlsdl_osx ./cmd/hlsdl
	@md5sum ./bin/hlsdl_osx

build_windows:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/hlsdl_windows.exe ./cmd/hlsdl
	@md5sum ./bin/hlsdl_windows.exe

build: build_osx build_linux build_windows

clean:
	rm -f ./bin/hlsdl*