DROP TABLE IF EXISTS audio_part;
DROP TABLE IF EXISTS project;
DROP TABLE IF EXISTS "user";

CREATE TABLE IF NOT EXISTS "user"
(
    user_id       uuid NOT NULL PRIMARY KEY default gen_random_uuid(),
    login         text NOT NULL
        constraint login_pk
            unique,
    password_hash text NOT NULL
);

CREATE TABLE IF NOT EXISTS project
(
    project_id uuid PRIMARY KEY default gen_random_uuid(),
    path       TEXT,
    user_id    uuid
        constraint user_id_fk
            references "user" (user_id),
    name       TEXT
);

CREATE TABLE IF NOT EXISTS audio_part
(
    part_id    uuid NOT NULL PRIMARY KEY default gen_random_uuid(),
    project_id uuid
        constraint project_id_fk
            references project (project_id),
    start      int,
    duration   int,
    text       TEXT,
    path       TEXT
);

CREATE OR REPLACE FUNCTION increment_project_name()
    RETURNS TRIGGER AS
$$
DECLARE
    next_project_number INTEGER;
BEGIN
    SELECT COUNT(*) + 1
    INTO next_project_number
    FROM project
    WHERE user_id = NEW.user_id;

    NEW.name := 'awesomeProject' || next_project_number;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER increment_project_name_trigger
    BEFORE INSERT
    ON project
    FOR EACH ROW
EXECUTE FUNCTION increment_project_name();
