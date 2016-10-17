
/*drop table if exists user;*/
create table if not exists admin (
  id integer primary key,
  handle text not null,
  hashed_password text not null
);

/*drop table if exists network;*/
create table if not exists network (
  id integer primary key,
  listening_port integer unique not null,
  name text not null
);

/*drop table if exists encoder;*/
create table if not exists encoder (
  id integer primary key,
  ip_address text not null,
  port integer not null default(23),
  name text null default ('New Encoder'),
  handle text not null,
  password text not null,
  network_id integer not null,
  foreign key(network_id) references network(id)
);
