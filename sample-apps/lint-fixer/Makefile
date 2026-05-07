.PHONY: build install run clean

build:
	tool-builder build --config tool.yaml -o ./bin/lint-fixer

install: build
	mkdir -p $(HOME)/bin
	install -m 755 ./bin/lint-fixer $(HOME)/bin/lint-fixer

run:
	tool-builder run --config tool.yaml

clean:
	rm -f ./bin/lint-fixer
