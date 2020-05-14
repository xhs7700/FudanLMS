use fudanlms;

create table if not exists users(
    id char(11) primary key,
    password char(64) not null,
    authority tinyint
);

delete from users;

create table if not exists books(
    isbn char(13) primary key,
    title varchar(64),
    author varchar(64)
);

delete from books;

create table if not exists borrec(
    id char(11),
    isbn char(13),
    bortime datetime,
    deadline datetime,
    extendtime tinyint,
    foreign key(id)references users(id),
    foreign key(isbn)references books(isbn)
);

delete from borrec;

create table if not exists retrec(
    id char(11),
    isbn char(13),
    bortime datetime,
    rettime datetime,
    primary key(id,isbn,bortime),
    foreign key(id)references users(id),
    foreign key(isbn)references books(isbn)
);

delete from retrec;

create table if not exists rmrec(
    isbn char(13),
    title varchar(64),
    author varchar(64),
    removetime datetime,
    reason varchar(128),
    primary key(isbn,removetime)
);

delete from rmrec;
