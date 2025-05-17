CREATE DATABASE central_repo;

\c central_repo

CREATE TABLE IF NOT EXISTS users (
    id int PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS groups (
    group_name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users_groups (
    user_id int references users(id),
    assigned_group TEXT references groups(group_name),

    PRIMARY KEY (user_id, assigned_group)
);

INSERT INTO users VALUES (1, 'Admin', 'Prova');
INSERT INTO groups VALUES ('ADMIN'), ('DOCTORS'), ('PATIENTS');

