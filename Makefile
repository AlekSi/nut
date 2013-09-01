all: short

prepare:
	go env
	go get -u launchpad.net/gocheck
	-go get -u github.com/kisielk/errcheck
	go get -u github.com/AlekSi/test_nut1
	-go get -u github.com/AlekSi/test_nut2
	-go get -u github.com/AlekSi/test_nut3

# format, vet, build
fvb:
	gofmt -e -s -w .
	go tool vet .
	go install github.com/AlekSi/nut
	go build -o gonut.exe -ldflags "-X main.NutVersion `git describe --tags --always`" github.com/AlekSi/nut/nut
	-errcheck github.com/AlekSi/nut
	-errcheck github.com/AlekSi/nut/nut
	-errcheck github.com/AlekSi/nut/integration_test

test: fvb
	cd ../test_nut1 && ../nut/gonut.exe pack
	go test -v github.com/AlekSi/nut -gocheck.v
	go test -v github.com/AlekSi/nut/nut -gocheck.v

short: test
	go test -v -short github.com/AlekSi/nut/integration_test -gocheck.v

full: test
	GONUTS_IO_SERVER=http://localhost:8080 go test -v github.com/AlekSi/nut/integration_test -gocheck.v
