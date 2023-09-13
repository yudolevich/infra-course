create database app;
create user app;
grant all on database app to app;
\connect app;
create table users(
  id serial primary key,
  name varchar(50),
  email varchar(100)
);
grant all on all tables in schema public to app;
