create table if not exists metadata
(
    id       text not null
        constraint id
            primary key,
    chunks   integer,
    filename text,
    b2_id    text,
    length   bigint
);

create table if not exists b2_uploads
(
    metadata_id text not null
        constraint metadata_id
            primary key,
    upload_url  text,
    token       text,
    upload_id   text,
    checksums   text[],
    local       boolean,
    name        text
);

create table if not exists expiry
(
    id        text not null
        constraint expiry_pk
            primary key,
    downloads smallint,
    date      timestamp
);

create table if not exists users
(
    id                  text not null
        constraint users_pk
            primary key,
    email               text,
    pw_hash             bytea,
    payment_id          text,
    member_expiration   timestamp,
    last_upgraded_month smallint default 0,
    protected_key       bytea,
    public_key          bytea,
    storage_available   bigint   default 0,
    storage_used        bigint   default 0,
    send_available      bigint   default 0,
    send_used           bigint   default 0,
    sub_duration        text     default ''::text,
    sub_type            text     default ''::text,
    sub_method          text     default ''::text
);

create table if not exists stripe
(
    customer_id text not null
        constraint stripe_pk
            primary key,
    payment_id  text
        constraint stripe_uk
            unique
);

create table if not exists verify
(
    identity        text not null
        constraint verification_pk
            primary key,
    code            text,
    date            timestamp,
    pw_hash         bytea,
    protected_key   bytea,
    public_key      bytea,
    root_folder_key bytea
);

create table if not exists vault
(
    id            text not null
        constraint vault_pk
            primary key,
    owner_id      text not null,
    name          text not null,
    b2_id         text    default ''::text,
    length        bigint,
    modified      timestamp,
    folder_id     text,
    chunks        integer,
    shared_by     text    default ''::text,
    protected_key bytea,
    link_tag      text    default ''::text,
    can_modify    boolean default true,
    ref_id        text
);

create table if not exists folders
(
    id            text  not null,
    name          text  not null,
    owner_id      text  not null,
    protected_key bytea not null,
    shared_by     text    default ''::text,
    parent_id     text    default ''::text,
    modified      timestamp,
    link_tag      text    default ''::text,
    can_modify    boolean default true,
    ref_id        text    default ''::text
);

create table if not exists sharing
(
    id           text not null
        constraint sharing_pk
            primary key,
    owner_id     text,
    recipient_id text,
    item_id      text,
    can_modify   boolean,
    is_folder    boolean
);

