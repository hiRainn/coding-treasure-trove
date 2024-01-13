#!/bin/bash
HOME=$(cd $(dirname $0);pwd)
PORT=3306
MYSQL_DIR=/home/service/var/mysql3306
CNF_DIR=/home/service/app/mysql3366/etc/my.cnf
DATA_DIR=/home/service/var/mysql3306/data

chown -R mysql.mysql $MYSQL_DIR

cd $HOME
if [ $(ps aux|grep -w mysqld|grep -v grep|wc -l) -eq 0 ];then
    ./bin/mysqld_safe --defaults-file=$CNF_DIR \
    --port=$PORT \
    --user=mysql --datadir=$DATA_DIR &
fi

flag=0
for ((i=1;i<=3;i++));do
    if [ $(ps aux|grep -w mysqld|grep -v grep|wc -l) -eq 1 ];then
        flag=1
        break
    else
        echo "try $i times ..."
    fi
    sleep 3
done

if [[ $flag -eq 1 ]];then
    echo "start success!"
else
    echo "start failed!"
fi