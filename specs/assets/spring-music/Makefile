OPENJDK_IMAGE ?= openjdk:17-slim

.PHONY: build
build:
	docker run -v $(CURDIR):/spring-music \
		-w /spring-music \
		-it $(OPENJDK_IMAGE) ./gradlew clean bootJar --refresh-dependencies
