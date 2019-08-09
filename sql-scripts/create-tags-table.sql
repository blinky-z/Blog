CREATE TABLE if not exists tags
(
    TAG_ID SERIAL PRIMARY KEY,
    TAG    varchar(36) not null
);

Create UNIQUE index tagsTagIndex on tags (TAG);

CREATE TABLE if not exists post_tags
(
    POST_ID INTEGER not null,
    TAG_ID  INTEGER not null,
    PRIMARY KEY (POST_ID, TAG_ID)
);

Create index postTagsTagIDIndex on post_tags (TAG_ID);