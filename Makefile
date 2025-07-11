# Makefile for TEN VAD Go project

.PHONY: build test clean

build:
	bash build.sh

test:
	bash run_test.sh

clean:
	rm -rf build/
	rm -f ten_vad.test 