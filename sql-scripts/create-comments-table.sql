CREATE TABLE if not exists comments
(
    ID        SERIAL PRIMARY KEY,
    POST_ID   INTEGER                 not null,
    PARENT_ID INTEGER,
    AUTHOR    CHARACTER VARYING(36)   not null,
    DATE      TIMESTAMPTZ DEFAULT NOW(),
    CONTENT   CHARACTER VARYING(2048) not null,
    DELETED   BOOLEAN     DEFAULT FALSE
);

CREATE INDEX if not exists postIdIndex ON comments (POST_ID);