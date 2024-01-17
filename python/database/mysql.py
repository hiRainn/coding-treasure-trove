# -*- coding: utf-8 -*-
import pymysql

# connect to mysql
connection = pymysql.connect(
    host='localhost',  
    user='username', 
    password='password',  
    database='database_name' 
)

# create cursor obj
cursor = connection.cursor()

# query
cursor.execute("SELECT * FROM table_name")
results = cursor.fetchall()
for row in results:
    print(row)

# insert
sql = "INSERT INTO table_name (column1, column2) VALUES (%s, %s)"
values = ("value1", "value2")
cursor.execute(sql, values)
connection.commit()

# update
sql = "UPDATE table_name SET column1 = %s WHERE column2 = %s"
values = ("new_value", "condition_value")
cursor.execute(sql, values)
connection.commit()

# delete
sql = "DELETE FROM table_name WHERE column = %s"
value = "value"
cursor.execute(sql, value)
connection.commit()

# close cursor and mysql
cursor.close()
connection.close()
