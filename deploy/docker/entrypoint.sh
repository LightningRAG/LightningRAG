#!/bin/bash
if [ ! -d "/var/lib/mysql/lrag" ]; then
    mysqld --initialize-insecure --user=mysql --datadir=/var/lib/mysql
    mysqld --daemonize --user=mysql
    sleep 5s
    mysql -uroot -e "create database lrag default charset 'utf8' collate 'utf8_bin'; grant all on lrag.* to 'root'@'127.0.0.1' identified by '123456'; flush privileges;"
else
    mysqld --daemonize --user=mysql
fi
redis-server &
if [ "$1" = "actions" ]; then
    cd /opt/lrag/server && go run main.go &
    cd /opt/lrag/web/ && yarn serve &
else 
    /usr/sbin/nginx &
    cd /usr/share/nginx/html/ && ./server &
fi
echo "lrag ALL start!!!"
tail -f /dev/null