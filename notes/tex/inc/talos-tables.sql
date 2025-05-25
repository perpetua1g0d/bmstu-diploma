CREATE SCHEMA IF NOT EXISTS infra2infra;

CREATE TABLE infra2infra."Permissions" (
    ClientName TEXT NOT NULL,
    ServerName TEXT NOT NULL,
    roles TEXT[] NOT NULL
);
