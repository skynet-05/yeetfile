create table if not exists invoices
(
    invoice_id text not null
        constraint invoices_pk
            primary key,
    payment_id text,
    source     text,
    date       timestamp
);