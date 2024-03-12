create table if not exists metadata
(
    id       text not null
        constraint id
            primary key,
    chunks   integer,
    filename text,
    salt     bytea,
    b2_id    text,
    length   integer
);

create table if not exists b2_uploads
(
    metadata_id text not null
        constraint metadata_id
            primary key,
    upload_url  text,
    token       text,
    upload_id   text,
    checksums   text[]
);

create table if not exists expiry
(
    id        text not null
        constraint expiry_pk
            primary key,
    downloads integer,
    date      timestamp
);

create table if not exists users
(
    id         text not null
        constraint users_pk
            primary key,
    email      text,
    pw_hash    bytea,
    meter      bigint,
    payment_id text,
    member_expiration timestamp,
    last_upgraded_month integer
);

create table if not exists stripe
(
    intent_id  text not null
        constraint stripe_pk
            primary key,
    account_id text,
    product_id text,
    quantity   integer,
    date       date
);

create table if not exists verify
(
    email   text not null
        constraint verification_pk
            primary key,
    code    text,
    date    date,
    pw_hash bytea
);