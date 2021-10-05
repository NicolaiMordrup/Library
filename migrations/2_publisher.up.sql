ALTER TABLE library RENAME TO _library_old;

CREATE TABLE library(
    isbn TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    createTime timestamp NOT NULL,
    updateTime timestamp NOT NULL,
    publisher TEXT
);

INSERT INTO library (isbn,title, createTime,updateTime)
	SELECT isbn, title, createTime,updateTime
	FROM _library_old;

DROP TABLE _library_old;