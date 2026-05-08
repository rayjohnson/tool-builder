.PHONY: build install clean

build:
	tool-builder build --config tool.yaml -o ./bin/rtfm

install: build
	mkdir -p $(HOME)/bin
	install -m 755 ./bin/rtfm $(HOME)/bin/rtfm

clean:
	rm -f ./bin/rtfm
