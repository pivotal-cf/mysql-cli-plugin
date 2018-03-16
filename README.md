# mysql-cli-plugin

A Cloud Foundry CLI plugin intended for interaction with [p.mysql](https://network.pivotal.io/products/pivotal-mysql/) service instances.

## Installation

### Prerequisites

* [Golang](https://golang.org/): 1.9+
* [CF CLI](https://github.com/cloudfoundry/cli): 6.31.0+
* [Docker](https://www.docker.com/)

```
$ go get github.com/pivotal-cf/mysql-cli-plugin
```

## Building
1. `cd` into the root directory of the project.
1. Execute the `./scripts/build-assets` script in the corresponding docker image to compile utilites (mysqldump, mysql) to be bundled with the plugin.
   
   ```
   $ docker run -v $PWD:$PWD -w $PWD -t cloudfoundry/cflinuxfs2 ./scripts/build-assets
   ```
1. Run go generate to create golang compatible static assets out of the utilities
   ```
   $ go generate .
   ```
1. Compile the final binary
   ```
   $ go build .
   ```

### Cleaning Up

```
$ packr clean
```