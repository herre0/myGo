Steps before run:

1- MySql Db Conf:
 CREATE DATABASE task;

CREATE TABLE task (
  id INT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(100),
  description VARCHAR(100),
  status VARCHAR(100)  
);

2- Set environment variables:
$ export DBUSER="yourMySqlUsername"
$ export DBPASS="yourMySqlPassword"
