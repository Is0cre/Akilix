.PHONY: all clean validate terminal web os icons print
all: web os icons terminal print
web os icons print:
	python3 tools/build-assets.py
terminal:
	python3 tools/build-terminal.py
validate:
	python3 tools/validate-assets.py
clean:
	find web desktop os print -type f \( -name '*.png' -o -name '*.ico' \) -delete
