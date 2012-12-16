all: test

prepare:
	go get -u launchpad.net/gocheck
	go get -u github.com/AlekSi/test_nut1

# format, vet, build
fvb:
	gofmt -e -s -w .
	go tool vet .
	go build -v -o gonut ./nut

test: fvb
	cd ../test_nut1 && rm -f *.nut
	cd ../test_nut1 && ../nut/gonut generate -v
	cd ../test_nut1 && ../nut/gonut check -v
	cd ../test_nut1 && ../nut/gonut pack -v
	cd ../test_nut1 && ../nut/gonut check -v test_nut1-0.0.1.nut
	cd ../test_nut1 && ../nut/gonut unpack -v test_nut1-0.0.1.nut

	go test -v ./...

	cd ../test_nut1 && ../nut/gonut install -v test_nut1-0.0.1.nut

	cd ../test_nut2 && rm -f *.nut
	cd ../test_nut2 && ../nut/gonut generate -v
	cd ../test_nut2 && ../nut/gonut check -v
	cd ../test_nut2 && ../nut/gonut pack -v
	cd ../test_nut2 && ../nut/gonut check -v test_nut2-0.0.2.nut
	cd ../test_nut2 && ../nut/gonut unpack -v test_nut2-0.0.2.nut
	cd ../test_nut2 && ../nut/gonut install -v test_nut2-0.0.2.nut

test_server: test
	cd ../test_nut1 && GONUTS_SERVER=localhost:8080 ../nut/gonut publish -v test_nut1-0.0.1.nut
	cd ../test_nut2 && GONUTS_SERVER=localhost:8080 ../nut/gonut publish -v test_nut2-0.0.2.nut
	cd ../test_nut1 && GONUTS_SERVER=localhost:8080 ../nut/gonut get -v test_nut2/0.0.2
