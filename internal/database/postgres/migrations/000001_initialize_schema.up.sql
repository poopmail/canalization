begin;

create table if not exists invites (
    code text not null,
    primary key (code)
);

create table if not exists accounts (
    username text not null,
    password text not null,
    admin bool not null default false,
    primary key (username)
);

create table if not exists mailboxes (
    address text not null,
    account text not null,
    primary key (address)
);

create table if not exists messages (
    id bigint not null,
    mailbox text not null,
    "from" text not null,
    subject text not null,
    content text not null,
    primary key (id)
);

commit;