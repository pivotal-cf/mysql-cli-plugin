Spring Music for Dedicated MySQL Team
============

This is a fork of the [spring-music](https://github.com/cloudfoundry-samples/spring-music) sample CF application used as a test app in the [dedicated-mysql-adapter-release](https://github.com/pivotal-cf/dedicated-mysql-adapter-release) bosh release. In the smoke test the app is pushed to CF and bound to a dedicated-mysql instance. The `/appinfo` endpoint is queried to verify the app has beensuccessfully bound to a mysql instance. The test also writes and reads data using the `/albums` endpoint.

The build/libs/spring-music.jar file is committed in this repo so that the app can be pushed without needing java in a test environment. **Rebuild the app using `./gradlew clean assemble` and commit the .jar file whenever changes to the app have been made.**

To push to CF:
```
$ cf push spring-music -m 512M -k 256M -p build/libs/spring-music.jar
```
---

Spring Music
============

This is a sample application for using database services on [Cloud Foundry](http://cloudfoundry.org) with the [Spring Framework](http://spring.io) and [Spring Boot](http://projects.spring.io/spring-boot/).

This application has been built to store the same domain objects in one of a variety of different persistence technologies - relational, document, and key-value stores. This is not meant to represent a realistic use case for these technologies, since you would typically choose the one most applicable to the type of data you need to store, but it is useful for testing and experimenting with different types of services on Cloud Foundry. 

The application use Spring Java configuration and [bean profiles](http://docs.spring.io/spring-boot/docs/current/reference/html/boot-features-profiles.html) to configure the application and the connection objects needed to use the persistence stores. It also uses the [Spring Cloud Connectors](http://cloud.spring.io/spring-cloud-connectors/) library to inspect the environment when running on Cloud Foundry. See the [Cloud Foundry documentation](http://docs.cloudfoundry.org/buildpacks/java/spring-service-bindings.html) for details on configuring a Spring application for Cloud Foundry.

## Running the application locally

One Spring bean profile should be activated to choose the database provider that the application should use. The profile is selected by setting the system property `spring.profiles.active` when starting the app.

The application can be started locally using the following command:

~~~
$ ./gradlew clean assemble
$ java -jar -Dspring.profiles.active=<profile> build/libs/spring-music.jar
~~~

where `<profile>` is one of the following values:

* `in-memory` (no external database required)
* `mysql`

If no profile is provided, `in-memory` will be used. If any other profile is provided, the appropriate database server must be started separately. The application will use the host name `localhost` and the default port to connect to the database. The connection parameters can be configured by setting the appropriate [Spring Boot properties](http://docs.spring.io/spring-boot/docs/current/reference/html/common-application-properties.html). 

If more than one of these profiles is provided, the application will throw an exception and fail to start.

## Running the application on Cloud Foundry

When running on Cloud Foundry, the application will detect the type of database service bound to the application (if any). If a service of one of the supported types (MySQL) is bound to the app, the appropriate Spring profile will be configured to use the database service. The connection strings and credentials needed to use the service will be extracted from the Cloud Foundry environment.

If no bound services are found containing any of these values in the name, then the `in-memory` profile will be used.

If more than one service containing any of these values is bound to the application, the application will throw an exception and fail to start.

After installing the 'cf' [command-line interface for Cloud Foundry](http://docs.cloudfoundry.org/cf-cli/), targeting a Cloud Foundry instance, and logging in, the application can be built and pushed using these commands:

~~~
$ ./gradlew clean assemble
$ cf push
~~~

The application will be pushed using settings in the provided `manifest.yml` file. The output from the command will show the URL that has been assigned to the application.

### Creating and binding services

Using the provided manifest, the application will be created without an external database (in the `in-memory` profile). You can create and bind database services to the application using the information below.

#### System-managed services

Depending on the Cloud Foundry service provider, persistence services might be offered and managed by the platform. These steps can be used to create and bind a service that is managed by the platform:

~~~
# view the services available
$ cf marketplace
# create a service instance
$ cf create-service <service> <service plan> <service name>
# bind the service instance to the application
$ cf bind-service <app name> <service name>
# restart the application so the new service is detected
$ cf restart
~~~

#### User-provided services

Cloud Foundry also allows service connection information and credentials to be provided by a user. In order for the application to detect and connect to a user-provided service, a single `uri` field should be given in the credentials using the form `<dbtype>://<username>:<password>@<hostname>:<port>/<databasename>`.

These steps use examples for username, password, host name, and database name that should be replaced with real values.

~~~
# create a user-provided MySQL database service instance
$ cf create-user-provided-service mysql-db -p '{"uri":"mysql://root:secret@dbserver.example.com:3306/mydatabase"}'
# bind a service instance to the application
$ cf bind-service <app name> <service name>
# restart the application so the new service is detected
$ cf restart
~~~

#### Changing bound services

To test the application with different services, you can simply stop the app, unbind a service, bind a different database service, and start the app:

~~~
$ cf unbind-service <app name> <service name>
$ cf bind-service <app name> <service name>
$ cf restart
~~~

#### Database drivers

Database driver for MySQL is included in the project.
