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