CREATE SCHEMA IF NOT EXISTS service2infra;

CREATE TABLE service2infra."Permissions" (
    ClientName TEXT NOT NULL,
    ServerName TEXT NOT NULL,
    roles TEXT[] NOT NULL
);
