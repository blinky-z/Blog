CREATE TABLE if not exists Users
(
  USERNAME    CHARACTER VARYING(36)  not null,
  EMAIL    CHARACTER VARYING(255) not null,
  PASSWORD CHARACTER VARYING(255) not null
);

CREATE UNIQUE INDEX if not exists loginIndex ON users (USERNAME);
CREATE UNIQUE INDEX if not exists emailIndex ON users (EMAIL);