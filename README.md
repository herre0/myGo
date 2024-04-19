Steps before run:<br>

1- MySql Db Conf:
<code>
 CREATE DATABASE task;
<br>
CREATE TABLE task (
  id INT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(100),
  description VARCHAR(100),
  status VARCHAR(100)  
);
</code>
2- Set environment variables:<br>
$ export DBUSER="yourMySqlUsername" <br>
$ export DBPASS="yourMySqlPassword"
