build:
	rm dist.zip
	GOARCH=amd64 GOOS=linux go build -o bin/application
	zip -r dist.zip bin
	rm -rf bin