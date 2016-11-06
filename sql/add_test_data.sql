-- Assumes that it is being run on an empty database

Insert into network (name) values ('fox');
Insert into network (name) values ('C-SPANNER');

insert into encoder(ip_address, port, name, network_id) values (
  '127.0.0.1',
  '8080',
  'Fox main',
  1 -- Should be fox
);

insert into encoder(ip_address, port, name, network_id) values (
  '127.0.0.1',
  '8081',
  'Fox backup',
  1 -- Should be fox
);

insert into encoder(ip_address, port, name, network_id) values (
  '128.0.0.1',
  '9000',
  'CSPAN-main',
  2 -- Should be C-SPANNER
);
