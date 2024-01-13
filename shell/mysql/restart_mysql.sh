#!/bin/bash

HOME=$(cd $(dirname $0);pwd)
PORT=3306
PASSWORD="password"
MYSQL_DIR=/home/service/app/mysql3306
SOCK_DIR=/home/service/app/mysql3306/tmp/mysql.sock
CNF_DIR=/home/service/app/mysql3306/etc/my.cnf
DATA_DIR=/home/service/var/mysql3306/data

function start_mysql(){
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
        return 0
    else
        return 1
    fi
}

function stop_mysql(){

    if [ $(ps aux|grep -w mysqld|grep -v grep|wc -l) -eq 0 ];then
        echo "mysql is not start"
        return 0
    fi
    cd $HOME
    $MYSQL_DIR/bin/mysqladmin -uroot -S $SOCK_DIR  -p$PASSWORD shutdown

    flag=0
    for((i=1;i<=3;i++));do
    if [ $(ps aux|grep -w mysqld|grep -v grep|wc -l) -eq 0 ];then
        flag=1
        break
    fi

        sleep 3
    done

    if [[ $flag -eq 1 ]];then
        return 0
    else
        return 1
    fi
}


stop_mysql
if [[ $? -ne 0 ]];then
    echo "stop failed！"
    exit(1)
fi
start_mysql
if [[ $? -ne 0 ]];then
    echo "start failed！"
    exit(1)
fi
echo "restart mysql success!"