This terraform deploys a small (n1-standard-1) on GCP and setups the Percona PAM
and file-based keyring encryption plugins.

This is meant for manual testing of the mysql-cli-plugin's support for migrating
from such source databases onto the platform.

This instance opens ssh (22/tcp) and the mysql port (3306/tcp) to the public
internet.  A random ssh keypair is generated and the private key may be found in
the terraform outputs.

**Example**

```
# export GOOGLE_APPLICATION_CREDENTIALS=/path/to/my/service_key.json
# export GOOGLE_PROJECT=my-project-name
#
# terraform apply
```

The terraform outputs will include an "instance_ip" and "ssh_private_key" for
connecting to the instance.

A dummy PAM authenticated MySQL user is created with a default password of
"password".
