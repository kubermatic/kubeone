IMAGE_REPOSITORY = <IMAGE-REPO>
APPLICATION_NAME = <APP-NAME>
BUILD_VERSION = <VERSION>

.PHONY: dependencies
dependencies: 
	yarn

.PHONY: build
build: dependencies 
	yarn theia build

.PHONY: run
run: build
	yarn start --hostname 0.0.0.0 --port 8081

.PHONY: docker-build
docker-build: 
	docker build -t ${IMAGE_REPOSITORY}/${APPLICATION_NAME}:${BUILD_VERSION} .

.PHONY: docker-run
docker-run: 
	docker run --security-opt seccomp=unconfined -it --rm --init --name ide -p 3000:3000 -v "<WORKSPACE-DIR>:/home/project:cached" ${IMAGE_REPOSITORY}/${APPLICATION_NAME}:${BUILD_VERSION}

.PHONY: docker-push
docker-push: docker-build
	docker push ${IMAGE_REPOSITORY}/${APPLICATION_NAME}:${BUILD_VERSION}
	