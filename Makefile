.PHONY: web

web:
	go build -tags yeetfile-web -o yeetfile-web ./web

clean:
	rm -f yeetfile-web
	rm -f yeetfile-cli
