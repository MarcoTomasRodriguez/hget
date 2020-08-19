clean:
	@echo "====> Remove installed binary"
	rm -f bin/hget

build:
	@echo "====> Build hget in ./bin "
	go build -o bin/hget

install:
	@echo "====> Installing hget"
	go install