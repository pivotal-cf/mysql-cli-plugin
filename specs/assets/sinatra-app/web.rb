# Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.
#
# This program and the accompanying materials are made available under the terms of the under the Apache License,
# Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy
# of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

require 'securerandom'
require 'sinatra'
require 'mysql2'

configure do
    enable :sessions
    services = JSON.parse(ENV['VCAP_SERVICES'])
    mysql_key = services.keys.select { |svc| svc =~ /p.mysql/i }.first
    mysql = services[mysql_key].first['credentials']
    mysql_conf = {host: mysql['hostname'], username: mysql['username'], port: mysql['port'], password: mysql['password']}
    MYSQL_CLIENT = Mysql2::Client.new(mysql_conf)
end

class Dbweb < Sinatra::Base
    get '/create-db' do
        # generate random value
        dbName =  "db_" + SecureRandom.hex
        MYSQL_CLIENT.query "create database #{dbName}"
        MYSQL_CLIENT.query "create table #{dbName}.t1 (data char(40), PRIMARY KEY(data))"
        MYSQL_CLIENT.query "insert into #{dbName}.t1 values(SHA1(RAND()))"
        dbValue = MYSQL_CLIENT.query "select * from #{dbName}.t1 limit 1"

        content_type :json
        {"db": dbName, "table": "t1", "value": dbValue.first["data"] }.to_json
    end

    get '/show-db' do
        dbHash = {}
        (MYSQL_CLIENT.query "SHOW DATABASES LIKE 'db\_%'").each(as: :array) do |db|
            dbname = db.first
            result = MYSQL_CLIENT.query "SELECT data FROM #{dbname}.t1 LIMIT 1"
            dbHash[dbname] = result.first["data"]
        end

        content_type :json
        dbHash.to_json
    end
end
