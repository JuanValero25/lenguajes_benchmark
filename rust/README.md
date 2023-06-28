host: "localhost",
user: "root",
password: "dbpwd",
database: "testdb",

let conn = MySqlConnection::connect("mysql://root:password@localhost/db").await?;
