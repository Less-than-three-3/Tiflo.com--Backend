DROP TABLE IF EXISTS "user";
DROP TABLE IF EXISTS project;
DROP TABLE IF EXISTS audio_part;

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
    project_id uuid,
    path       TEXT,
    user_id    uuid
        constraint user_id_fk
            references "user" (user_id),
    name       TEXT
);

CREATE TABLE IF NOT EXISTS audio_part
(
    part_id    uuid NOT NULL PRIMARY KEY,
    project_id uuid
        constraint project_id_fk
            references project (project_id),
    start      int,
    duration   int,
    text       TEXT,
    path       TEXT
);
