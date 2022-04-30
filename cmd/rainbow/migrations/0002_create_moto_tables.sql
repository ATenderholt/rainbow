-- +goose Up
CREATE TABLE IF NOT EXISTS moto_requests (
    id              integer primary key autoincrement,
    service         text not null,
    method          text not null,
    path            text not null,
    authorization   text not null,
    content_type    text not null,
    target          text not null,
    payload         text not null
);