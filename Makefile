.PHONY: backend web cli

web:
	@echo -----------------------------------------
	@echo "   Generating TypeScript files from go..."
	@echo -----------------------------------------
	go run utils/generate_typescript.go ./web/ts
	@echo -----------------------------------------
	@echo "   Compiling TypeScript to JavaScript..."
	@echo -----------------------------------------
	tsc --removeComments

backend: web
	@echo -----------------------------------------
	@echo "   Building backend..."
	@echo -----------------------------------------
	go build -ldflags="-s -w" -tags yeetfile-server -o yeetfile-server ./backend
	@echo -----------------------------------------
	@echo "   Build complete: ./yeetfile-server"
	@echo -----------------------------------------

cli:
	go build -ldflags="-s -w" -tags yeetfile -o yeetfile ./cli

deploy:
	GOOS=linux GOARCH=amd64 go build \
		 -tags yeetfile-server \
		 -o yeetfile-server-deploy \
		 ./backend
	mv yeetfile-server-deploy ansible/roles/yeetfile/files/

deploy_dev: deploy
	test -s ./ansible/roles/yeetfile/files/dev.env || { echo "dev.env not in ansible/roles/yeetfile/files -- exiting..."; exit 1; }
	ansible-playbook -i ansible/inventory/dev.yml ansible/deploy.yml

deploy_prod: deploy
	test -s ./ansible/roles/yeetfile/files/prod.env || { echo "prod.env not in ansible/roles/yeetfile/files -- exiting..."; exit 1; }
	ansible-playbook -i ansible/inventory/prod.yml deploy.yml

clean:
	rm -f yeetfile-web
	rm -f yeetfile
