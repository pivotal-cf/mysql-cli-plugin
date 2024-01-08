.PHONY: build
build:
	docker run -v $(CURDIR):/spring-music \
		-w /spring-music \
		-it openjdk:17-slim ./gradlew clean bootJar --refresh-dependencies
