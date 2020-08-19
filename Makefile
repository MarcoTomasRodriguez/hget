deps:
	@echo "====> Install dependencies..."
	go get -d github.com/fatih/color
	go get -d github.com/mattn/go-colorable
	go get -d github.com/mattn/go-isatty
	go get -d github.com/fatih/color
	go get -d gopkg.in/cheggaaa/pb.v1
	go get -d github.com/mattn/go-isatty

clean:
	@echo "====> Remove installed binary"
	rm -f bin/hget

install: deps
	@echo "====> Build hget in ./bin "
	go build -o bin/hget
