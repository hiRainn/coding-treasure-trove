#!/bin/bash
HOME=$(cd $(dirname $0);pwd)
PORT=3306
MYSQL_DIR=/home/service/app/mysql3306
SOCK_DIR=/home/service/app/mysql3306/tmp/mysql.sock
PASSWORD="your password"

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
  echo "stop success!"
else
  echo "stop failed!"
fi