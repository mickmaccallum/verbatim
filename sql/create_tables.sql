
/*DROP TABLE IF EXISTS user;*/
CREATE TABLE user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  is_admin BOOLEAN NOT NULL DEFAULT(0),
  handle TEXT NOT NULL,
  hashed_password TEXT NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  status INTEGER NOT NULL DEFAULT(0)
);

/*DROP TABLE IF EXISTS network;*/
CREATE TABLE network (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL
);

/*DROP TABLE IF EXISTS encoder;*/
CREATE TABLE encoder (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ip_address TEXT NOT NULL,
  port INTEGER NOT NULL,
  status INTEGER NOT NULL DEFAULT(0),

  network_id INTEGER NOT NULL,
  FOREIGN KEY(network_id) REFERENCES network(id)
);

DROP TABLE IF EXISTS connection;
CREATE TABLE connection (
  id INTEGER PRIMARY KEY AUTOINCREMENT

  /*captioner_id INTEGER NOT NULL,
  FOREIGN KEY(captioner_id) REFERENCES user(id),
  network_id INTEGER NOT NULL,
  FOREIGN KEY(network_id) REFERENCES network(id),*/
  /*encoder_id INTEGER NOT NULL,
  FOREIGN KEY(encoder_id) REFERENCES encoder(id)*/
);
