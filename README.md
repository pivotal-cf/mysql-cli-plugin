# mysql-cli-plugin

The mysql-tools plugin can be used to migrate a p-mysql (v1) service instance to a [p.mysql (v2)](https://network.pivotal.io/products/pivotal-mysql/) service instance.

## Installation

A binary version of the plugin can be installed from the CF-Community plugin repository:
```
$ cf install-plugin -r CF-Community MysqlTools
```

## Usage

Once the plugin is installed, migrate a v1 service instance to a new v2 service instance with the following command:

### Option 1: Specify V2 Destination Plan
```
$ cf mysql-tools migrate V1-INSTANCE -p V2-PLAN
```

Where `V1-INSTANCE` is the the name of the v1 service instance that you wish to migrate, and `V2-PLAN` is the name of the v2 service plan to use for the new v2 service instance.

This will create a new v2 service instance and copy the data from the v1 service instance into it.

At the end of this operation, the v2 service instance will have the same name as the original v1 service instance (`V1-INSTANCE`),
and the v1 instance will have `-old` appended to its name.

### Option 2: Specify V2 Destination Service Instance
```
$ cf mysql-tools migrate V1-INSTANCE -s V2-DESTINATION-SI
```

Where `V1-INSTANCE` is the the name of the v1 service instance that you wish to migrate,
and `V2-DESTINATION-DB` is an existing v2 service instance.

This will populate the v2 service instance with the v1 service instance's data. 

The plugin will not overwrite any existing data in a targeted v2 schema. 
Any overlapping v2 schemas must be empty.
If the same schema exists in both v1 and v2 and the v2 schema is non-empty, the plugin halts and exits with status 1.

This operation makes no changes to the v1 and v2 service instance names.

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
1. Run the following script to build the mysql binary and other assets, generate golang fixtures, and compile the binary
   
   ```
   $ ./scripts/build-plugin
   ```

Once the build-plugin script has been run and the assets have been built, you can alternatively use the following commands to just build the go code
1. Generate fixtures
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
