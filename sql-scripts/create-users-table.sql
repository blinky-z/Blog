CREATE TABLE if not exists users
(
    USERNAME CHARACTER VARYING(36)  not null,
    EMAIL    CHARACTER VARYING(255) not null,
    PASSWORD CHARACTER VARYING(255) not null
);

CREATE UNIQUE INDEX if not exists usernameIndex ON users (USERNAME);
CREATE UNIQUE INDEX if not exists emailIndex ON users (EMAIL);