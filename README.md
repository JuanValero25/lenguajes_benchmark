DATABASE_HOST=localhost
DATABASE_USER=dbuser
DATABASE_PASSWORD=dbpwd
DATABASE_NAME=testdb


docker run --name some-mysql -p 3307:3306 -e MYSQL_ROOT_PASSWORD=dbpwd -d mysql:8.0.33

create table users
(
first   varchar(255) null,
last    varchar(255) null,
city    varchar(255) null,
country varchar(255) null,
age     int          null,
email   varchar(255) not null,
constraint users_pk
primary key (email)
);