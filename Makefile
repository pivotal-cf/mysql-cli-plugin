integration: install-plugin
	ginkgo -skipPackage specs -r .

build-plugin:
	./scripts/build-plugin

install-plugin: build-plugin
	cf install-plugin -f mysql-cli-plugin

acceptance:
	ginkgo specs
