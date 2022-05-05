-- +goose Up
CREATE TABLE IF NOT EXISTS sqs_queues (
    id integer primary key autoincrement,
    name text not null
);

CREATE TABLE IF NOT EXISTS sqs_queue_attributes (
    id		    integer primary key autoincrement,
    queue_id    integer,
    name	    text not null,
    value       text,
    foreign key (queue_id) references sqs_queues(id)
);

CREATE TABLE IF NOT EXISTS sqs_queue_tags (
    id          integer primary key autoincrement,
    queue_id    integer,
    name        text not null,
    value       text,
    foreign key (queue_id) references sqs_queues(id)
);