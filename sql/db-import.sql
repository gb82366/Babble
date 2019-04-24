drop database if exists babble;

create database babble;

\c babble

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;
COMMENT ON EXTENSION postgis IS 'PostGIS geometry, geography, and raster spatial types and functions';

create table users(
    ID serial primary key,
    username varchar(64) UNIQUE NOT NULL,
    password varchar(128) NOT NULL,
    email varchar(254) UNIQUE NOT NULL,
    auth_token varchar(32)
);

create table threads(
    ID serial primary key,
    poster_id INT references users(ID),
    score INT DEFAULT 1,
    item_text varchar(500) NOT NULL,
    location geography(POINT),
    posted_on timestamp
);

create table comments(
    ID serial primary key,
    poster_id INT references users(ID),
    parent_thread INT references threads(ID),
    score INT DEFAULT 1,
    item_text varchar(500) NOT NULL, 
    posted_on timestamp
);

create table thread_votes(
    ID serial primary key,
    poster_id INT references users(ID),
    thread_id INT references threads(ID),
    vote SMALLINT
);

create table comment_votes(
    ID serial primary key,
    poster_id INT references users(ID),
    comment_id INT references comments(ID),
    vote SMALLINT
);

create table thread_names(
    ID serial primary key,
    thread_id INT references threads(ID),
    user_id INT references users(ID),
    icon INT,
    color varchar(7),
    username varchar(64)
);

INSERT INTO users (username, password, email) VALUES ('fuck', 'THIS_IS_A_SECRET', 'you@know.who');
