.PHONY: build install clean

build:
	tool-builder build --config tool.yaml -o ./bin/commit-msg

install: build
	mkdir -p $(HOME)/bin
	install -m 755 ./bin/commit-msg $(HOME)/bin/commit-msg

clean:
	rm -f ./bin/commit-msg
