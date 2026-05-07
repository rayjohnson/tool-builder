.PHONY: build install run clean

build:
	tool-builder build --config tool.yaml -o ./bin/commit-msg

install: build
	mkdir -p $(HOME)/bin
	install -m 755 ./bin/commit-msg $(HOME)/bin/commit-msg

run:
	tool-builder run --config tool.yaml

clean:
	rm -f ./bin/commit-msg
