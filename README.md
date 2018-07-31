# mysql-cli-plugin

The mysql-tools plugin can be used to migrate a p-mysql (v1) service instance to a [p.mysql (v2)](https://network.pivotal.io/products/pivotal-mysql/) service instance.

## Installation

A binary version of the plugin can be installed from the CF-Community plugin repository:
```
$ cf install-plugin -r CF-Community MysqlTools
```

## Usage

Once the plugin is installed, migrate a v1 service instance to a new v2 service instance with the following command:

```
$ cf mysql-tools migrate V1-INSTANCE V2-PLAN
```

Where `V1-INSTANCE` is the the name of the v1 service instance that you wish to migrate, and `V2-PLAN`  is the name of the v2 service plan to use for the new v2 service instance.

This will create a new v2 service instance and copy the data from the v1 service instance into it.

At the end of this operation, the v2 service instance will have the same name as the original v1 service instance (`V1-INSTANCE`),
and the v1 instance will have `-old` appended to its name.

More detailed instructions are available in the [MySQL for PCF docs.](http://docs.pivotal.io/p-mysql/2-3/migrate-to-v2.html)

Some Notice:
* Stop all apps bound to the service instance before migrating.
* The Database instance should not be receiving any write traffic while the migration is happening.
* The apps bounded to the original v1 service instance need to be manually bound to the new v2 service instance at the end of the migration.
* There will be token timeout messages when migrating lots of data, which can be ignored.

## Building

### Prerequisites

* [Golang](https://golang.org/): 1.9+
* [CF CLI](https://github.com/cloudfoundry/cli): 6.31.0+
* [Docker](https://www.docker.com/)

1. `cd` into the root directory of the project.
1. Execute the `./scripts/build-assets` script in the corresponding docker image to compile utilites (mysqldump, mysql) to be bundled with the plugin.
   
   ```
   $ docker run -v $PWD:$PWD -w $PWD -t cloudfoundry/cflinuxfs2 ./scripts/build-assets
   ```
1. Run go generate to create golang compatible static assets out of the utilities
   ```
   $ go generate ./...
   ```
1. Compile the final binary
   ```
   $ go build .
   ```

### Cleaning Up

```
$ packr clean
```
