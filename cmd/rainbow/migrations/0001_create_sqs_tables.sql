-- +goose Up
CREATE TABLE IF NOT EXISTS sqs_queue_attribute (
    id		integer primary key autoincrement,
    name    text not null,
    key		text not null,
    value   text
);

CREATE TABLE IF NOT EXISTS sqs_queue_tag (
    id      integer primary key autoincrement,
    name    text not null,
    key     text not null,
    value   text
);