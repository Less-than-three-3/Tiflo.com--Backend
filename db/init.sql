DROP TABLE IF EXISTS response;
DROP TABLE IF EXISTS request;

CREATE TABLE IF NOT EXISTS "user"
(
    user_id       uuid        NOT NULL PRIMARY KEY,
    login         varchar(40) NOT NULL
        constraint login_pk
            unique,
    password_hash varchar(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS project
(
    project_id   uuid,
    image_name   text,
    user_id      uuid
        constraint user_id_fk
            references "user" (user_id),
    project_name text
);