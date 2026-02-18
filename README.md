# mysql-cli-plugin

The mysql-tools plugin can be used to establish replication between
two instances of [pxc-release](https://github.com/cloudfoundry/pxc-release) configured to allow multi-site replication. This plugin
functionality is for use alongside VMWare Tanzu for MySQL on Tanzu Platform. More information is available on [Using Tanzu for
MySQL for multi-site replication](https://techdocs.broadcom.com/us/en/vmware-tanzu/platform/tanzu-mysql-tanzu-platform/10-1/mysql-tp/use-multi-site.html). 

The mysql-tools can also be used to migrate a legacy p-mysql (v1) service instance to a
[p.mysql (v2)](https://network.tanzu.vmware.com/products/pivotal-mysql/) service instance.

## Installation

A binary version of the plugin can be installed from the CF-Community plugin repository:

```
$ cf install-plugin -r CF-Community MysqlTools
```

## Usage - Configure multi-site using the mysql-tools plug-in

This procedure is available on the VMware documentation page [Using Tanzu for
MySQL for multi-site replication](https://techdocs.broadcom.com/us/en/vmware-tanzu/platform/tanzu-mysql-tanzu-platform/10-1/mysql-tp/use-multi-site.html#create-multi-dc-with-mysql-tools).

## Usage - Service Instance Migration

Once the plugin is installed, migrate a v1 service instance to a new v2 service instance with the following command:

```
$ cf mysql-tools migrate V1-INSTANCE V2-PLAN
```

Where `V1-INSTANCE` is the name of the v1 service instance that you wish to migrate, and `V2-PLAN`  is the name of the
v2 service plan to use for the new v2 service instance.

This will create a new v2 service instance and copy the data from the v1 service instance into it.

At the end of this operation, the v2 service instance will have the same name as the original v1 service
instance (`V1-INSTANCE`), and the v1 instance will have `-old` appended to its name.

More detailed instructions are available in the
[VMware Tanzu for MySQL on Tanzu Platform for Cloud Foundry Documentation](https://techdocs.broadcom.com/us/en/vmware-tanzu/platform/tanzu-mysql-tanzu-platform/3-3/mysql-tp/migrate-data.html).

Some Notice:

* Stop all apps bound to the service instance before migrating.
* The Database instance should not be receiving any write traffic while the migration is happening.
* The apps bounded to the original v1 service instance need to be manually bound to the new v2 service instance at the
  end of the migration.
* There will be token timeout messages when migrating lots of data, which can be ignored.

## Building

### Prerequisites

* [Go](https://golang.org/): 1.21+
* [CF CLI](https://github.com/cloudfoundry/cli): 6.53.0+
* [Docker](https://www.docker.com/)

1. `cd` into the root directory of the project.
2. Run the following script to build the mysql binary and other assets, generate golang fixtures, and compile the binary

   ```
   $ ./scripts/build-plugin
   ```

## Testing

## Running unit tests

### Prerequisites

* [Go](https://golang.org/): 1.21+
* [Docker](https://www.docker.com/)

Some of the tests use Docker to integrate with a MySQL database and require the docker cli and a local docker daemon for
spinning up containers.

```
$ ./scripts/run-unit-and-docker-tests
```

## Running System Tests

### Prerequisites

* [Go](https://golang.org/): 1.21+
* [CF CLI](https://github.com/cloudfoundry/cli): 6.53.0+
* [Docker](https://www.docker.com/)

System tests interact with a real environment running Cloudfoundry, for instance
a [bosh-bootloader](https://github.com/cloudfoundry/bosh-bootloader) environment
with [cf-deployment](https://github.com/cloudfoundry/cf-deployment).

System tests assume the plugin provided by this repo has already been installed. To run these tests you should:

1. Install the cli plugin

   ```
    $ ./scripts/build-plugin
    $ cf install-plugin -f ./mysql-cli-plugin
    ```

2. Run the test script

   ```
   $ ./scripts/run-specs
   ```

   Specific tests can be run by passing [ginkgo](https://github.com/onsi/ginkgo) filtering options to the underlying
   script.

   **Example**
   ```
   $ ./scripts/run-specs --label-filter="smoke_test"
   ```
