all: clean build

build:
	go build -buildmode=plugin -o access-log-plugin.so .

clean:
	go clean

login:
	echo ${GH_PAT} | docker login ghcr.io -u USERNAME --password-stdin

b:
	docker buildx build -t gateway .

r:
	docker-compose down --remove-orphans && \
    docker-compose rm -f -v && \
	docker-compose up && \
	docker-compose down --remove-orphans && \
	docker-compose rm -f -v

br: b r

run-ci:
	docker buildx build -t gateway . && docker run -d -p "8080:8080" gateway && sleep 5