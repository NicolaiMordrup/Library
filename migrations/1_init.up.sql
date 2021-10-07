CREATE TABLE library(
    isbn TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    createTime timestamp NOT NULL,
    updateTime timestamp NOT NULL
);

CREATE TABLE author(
    isbn TEXT PRIMARY KEY,
    firstName TEXT NOT NULL,
    lastName TEXT NOT NULL
);

