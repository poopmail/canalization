begin;

create table if not exists invites (
    "code" text not null,
    "created" bigint not null default date_part('epoch'::text, now()),
    primary key ("code")
);

create table if not exists accounts (
    "id" bigint not null,
    "username" text not null unique,
    "password" text not null,
    "admin" bool not null default false,
    "created" bigint not null default date_part('epoch'::text, now()),
    primary key ("id")
);

create table if not exists refresh_tokens (
    "id" bigint not null,
    "account" bigint not null,
    "token" text not null,
    "description" text not null,
    "created" bigint not null default date_part('epoch'::text, now()),
    primary key ("id", "account")
);

create table if not exists mailboxes (
    "address" text not null,
    "account" bigint not null,
    "created" bigint not null default date_part('epoch'::text, now()),
    primary key ("address")
);

create table if not exists messages (
    "id" bigint not null,
    "mailbox" text not null,
    "from" text not null,
    "subject" text not null,
    "content_plain" text not null,
    "content_html" text not null,
    "created" bigint not null default date_part('epoch'::text, now()),
    primary key ("id")
);

commit;