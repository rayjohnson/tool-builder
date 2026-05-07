.PHONY: build install clean

build:
	tool-builder build --config tool.yaml -o ./bin/test-builder

install: build
	mkdir -p $(HOME)/bin
	install -m 755 ./bin/test-builder $(HOME)/bin/test-builder

clean:
	rm -f ./bin/test-builder
