CREATE TABLE if not exists posts
(
    ID       SERIAL PRIMARY KEY,
    TITLE    CHARACTER VARYING(120) not null,
    AUTHOR   character VARYING(36)  not null,
    DATE     TIMESTAMPTZ DEFAULT NOW(),
    METADATA text                   not null,
    SNIPPET  text                   not null,
    CONTENT  text                   not null
);