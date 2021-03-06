CREATE TABLE if not exists posts
(
    ID         SERIAL PRIMARY KEY,
    TITLE      CHARACTER VARYING(200) not null,
    DATE       TIMESTAMPTZ DEFAULT NOW(),
    METADATA   text                   not null,
    SNIPPET    text                   not null,
    CONTENT    text                   not null,
    CONTENT_MD text                   not null
);