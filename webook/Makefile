.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=amd64 go build -o webook .
	@docker rmi -f codersx/webook:v0.0.1 
	@docker build -t codersx/webook:v0.0.1 .