
-- 
/*drop table if exists user;*/
create table if not exists admin (
  id int primary key,
  handle text not null,
  hashed_password text not null,
);

/*drop table if exists ports;*/
create table if not exists ports (
  id int primary key,
  port int not null,
  network_id int,
  foreign key(network_id) references network(id) 
);


/*drop table if exists network;*/
create table if not exists network (
  id integer primary key,
  name text not null
);

/*drop table if exists encoder;*/
create table if not exists encoder (
  id int primary key,
  ip_address text not null,
  port integer not null default(23),
  status integer not null default(0),
  name text null default ('New Encoder'),
  handle text not null,
  password text not null,
  network_id integer not null,
  foreign key(network_id) references network(id)
);

