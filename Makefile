.PHONY: build clean

build:
	mkdir -p dist
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/loggen
	zip -j dist/log-generator.zip bootstrap
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/chaosgen
	zip -j dist/chaos-injector.zip bootstrap
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/detector
	zip -j dist/detector.zip bootstrap
	rm -f bootstrap

clean:
	rm -rf dist bootstrap
