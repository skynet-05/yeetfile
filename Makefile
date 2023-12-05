.PHONY: web cli

web:
	go build -ldflags="-s -w" -tags yeetfile-web -o yeetfile-web ./web

cli:
	go build -ldflags="-s -w" -tags yeetfile -o yeetfile ./cli

clean:
	rm -f yeetfile-web
	rm -f yeetfile
