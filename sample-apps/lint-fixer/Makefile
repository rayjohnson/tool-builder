.PHONY: build install clean

build:
	tool-builder build --config tool.yaml -o ./bin/lint-fixer

install: build
	mkdir -p $(HOME)/bin
	install -m 755 ./bin/lint-fixer $(HOME)/bin/lint-fixer

clean:
	rm -f ./bin/lint-fixer
