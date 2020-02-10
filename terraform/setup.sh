#!/bin/bash

set -eux

setup_mysql() {
    addgroup --system mysql
    adduser --ingroup mysql \
        --system \
        --disabled-login \
        --no-create-home \
        --home /var/lib/mysql \
        --shell /bin/false \
        --gecos "MySQL Server" mysql >/dev/null

    mkdir -p /etc/mysql/conf.d

    cat > /etc/mysql/conf.d/auth_pam.cnf <<EOF
[mysqld]
plugin-load-add = auth_pam.so
plugin-load-add = auth_pam_compat.so
EOF

    cat > /etc/mysql/conf.d/encrypt_data_at_rest.cnf <<EOF
[mysqld]
early-plugin-load=keyring_file.so
keyring_file_data=/var/lib/mysql/keyring
innodb_encrypt_tables = on
EOF
}

setup_percona_utils() {
    DEBIAN_FRONTEND=noninteractive \
        apt-get update -y

    wget "https://repo.percona.com/apt/percona-release_latest.$(lsb_release -sc)_all.deb"
    apt install -y -f "./percona-release_latest.$(lsb_release -sc)_all.deb"

    percona-release setup ps57

    DEBIAN_FRONTEND=noninteractive \
        apt-get update -y

    DEBIAN_FRONTEND=noninteractive \
        apt-get install -y \
            sysbench \
            percona-server-server-5.7
}

setup_pam() {

    # Test setup using pam_unix to authenticate mysql users against the /etc/shadow database
    # The mysql user must have access to /etc/shadow or pam will be unable to authenticate
    cat > /etc/pam.d/mysqld <<EOF
auth       required     pam_warn.so
auth       required     pam_unix.so audit
account    required     pam_unix.so audit
EOF

    chgrp mysql /etc/shadow
    chmod g+r /etc/shadow

    adduser \
        --disabled-login \
        --no-create-home \
        --home /home/pam_user \
        --shell /bin/false \
        --gecos "MySQL PAM User for testing" pam_user >/dev/null

    chpasswd <<< "pam_user:password"

    mysql <<EOF
CREATE USER 'pam_user'@'%' IDENTIFIED WITH auth_pam;
GRANT SELECT ON service_instance_db.* TO 'pam_user';
EOF
}

initialize_mysql_encryption_at_rest() {
# create at least one table, aand this should be auto-encrypted per
# the "innodb-encrypt-tables" option
    mysql <<EOF
CREATE SCHEMA service_instance_db;
CREATE TABLE service_instance_db.t1 (id int primary key);
EOF
}

main() {
    setup_mysql
    setup_percona_utils
    initialize_mysql_encryption_at_rest
    setup_pam
}

main "$@"
