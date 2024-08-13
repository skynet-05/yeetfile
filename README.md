<div align="center">
  <img style="width: 500px; padding-bottom: 20px;" src="./backend/static/img/yeetfile-logo.png?raw=true">
  <br><br>

  <p>A private and secure file vault + temporary file/text sharing service</p>

[![Latest Release](https://img.shields.io/github/v/release/benbusby/yeetfile)](https://github.com/benbusby/yeetfile/releases)
[![License: AGPLv3](https://img.shields.io/github/license/benbusby/yeetfile)](https://opensource.org/license/agpl-v3)

[![Tests (CLI)](https://github.com/benbusby/yeetfile/actions/workflows/go-tests.yml/badge.svg)](https://github.com/benbusby/yeetfile/actions/workflows/go-tests.yml)
[![Tests (JS)](https://github.com/benbusby/yeetfile/actions/workflows/ts-tests.yml/badge.svg)](https://github.com/benbusby/yeetfile/actions/workflows/ts-tests.yml)

<table>
  <tr>
    <td><a href="https://sr.ht/~benbusby/yeetfile">SourceHut</a></td>
    <td><a href="https://github.com/benbusby/yeetfile">GitHub</a></td>
  </tr>
</table>
</div>
<hr>

Contents
1. [About](#about)
1. [Features](#features)
    1. [YeetFile Vault](#yeetfile-vault)
    1. [YeetFile Send](#yeetfile-send)
    1. [Accounts](#accounts)
    1. [Other](#other)
1. [How It Works](#how-it-works)
    1. [Uploading and Downloading](#uploading-and-downloading)
    1. [Billing](#billing)
1. [Self-Hosting](#self-hosting)
1. [Development](#development)
    1. [Requirements](#requirements)
    1. [Setup](#setup)
    1. [Building](#building)
    1. [Environment Variables](#environment-variables)
1. [Support](#support)

## About

YeetFile is a file vault and file/text transferring service, with both a [web](https://yeetfile.com) and
[CLI client](https://github.com/benbusby/yeetfile/releases) officially supported. All content is encrypted
locally, and the server is incapable of decrypting any transmitted content.

## Features

### YeetFile Vault

- File storage + folder creation
- File + folder sharing w/ YeetFile users
  - Read/write permissions per user
- No upload size limit

### YeetFile Send

- Configurable upload settings
  - Expiration date/time
    - Configurable to N minutes/hours/days
      - Max 30 days
  - Number of downloads
    - Max 10 downloads
  - Optional password protection
- Free text transfers (up to 2000 characters)

### Accounts

- Email not required at signup
  - Account ID-only signup allowed
  - Signup not required for text-only transfers
- Options to pay via Stripe or BTCPay
  - BTC and XMR supported via BTCPay
  - Available for <$3/month
  - Not required when self-hosting
  - Ability to rotate payment ID to remove record of payment

### Other

- Server-specific passwords (optional for self-hosting)
- Easily self-hosted
  - Official CLI can be configured to use any server

## How It Works

### Uploading and Downloading

When uploading a file with YeetFile, an optional password is provided by the user to protect the file contents.
Whether or not a password is provided, however, a "pepper" (an additional secret value added prior to hashing)
will be generated using 3 random works from a list of ~1K words, as well as a random digit placed either before
or after any of the 3 words. This pepper + (optional) password value is used to derive a key via PBKDF2 with
SHA-512 hashing. The file is then split into individual 5MB chunks, and the key is used to encrypt each individual
chunk using AES-GCM before uploading the encrypted chunk to the server.

On the server side, a file ID is generated for the upload process and is used to associate individual file
chunks together. As file chunks are uploaded, the server forwards these chunks to a Backblaze B2 bucket. Once
the upload is complete, the server returns a link to the uploader that contains the file ID. For example:

<pre>https://yeetfile.com/file_8z74nn1spdni</pre>

This link is then formatted (by the client) with the pepper that the client generated before uploading:

<pre>https://yeetfile.com/file_8z74nn1spdni<b>#neither-unarmored-uncle9</b></pre>

The pepper is appended as part of a URL's "fragment", which is never sent to the server when making a request.

When the recipient opens the link (in web or CLI), the pepper is extracted from the link and used by iteself to
perform an initial decryption attempt of the filename. If it fails to decrypt the filename, it means that the
uploader set a password during the upload process, and the web or CLI client will prompt the recipient to enter
a password. Once the download is complete, the download

With this model, the server never sees the file's decrypted contents or filename, and it never sees the randomly
generated pepper required for decrypting file contents. In addition to not having the ability to decrypt any file,
YeetFile is also not able to determine which file was uploaded by a particular account. The user's account ID is
never tied to an uploaded file ID in the database, and the file's size is rounded down to the nearest megabyte
when adjusting the user's file transfer usage (to avoid correlating size of an upload to a user's modified
transfer limit).

Note: Uploading text works the same way, just with a single chunk limited to 1K characters.

### Billing

Each user starts with a payment ID that is separate from their account ID. When they purchase either a membership
or a file transfer upgrade, the payment ID is used in place of their account ID for the checkout process. If the
checkout is successful, a record of their purchase is added to a table in the database that corresponds to how
the user made the purchase (either via Stripe or BTCPay). At the same time, the user's payment ID is updated to
an entirely new ID, without the ability to recover what their old payment ID. The checkout confirmation will
display the original payment ID to help with any billing issues, but unless the user provides that payment ID in
correspondence with YeetFile, YeetFile has no way of knowing which payment ID belongs to which user.

Note: This functions the same when checking out via Stripe or BTCPay.

## Self-Hosting

You can quickly create your own instance of YeetFile using `docker compose`:

`docker compose up`

This will create the Postgres db and the server running on http://localhost:8090. You can modify the docker-compose.yml
to use an external data volume by running `docker volume create --name=yeetfile_data` and including the following in
your docker-compose.yml:

```
volumes:
  yeetfile_data:
    external: true
```

You should create your own `.env` file with whichever variables needed to customize your instance
(see: [Environment Variables](#environment-variables)).

## Development

### Requirements

- Go v1.20+
- PostgreSQL 15+
- Node.js 22.2+
  - `npm install typescript`
- A Backblaze B2 account (optional)

### Setup

1. Start PostgreSQL
2. Create a database in PostgreSQL named "yeetfile" using user "postgres"
    1. This can be customized to your preference, just use the `YEETFILE_DB_*` environment variables outlined
       below to configure before launching the server.
3. Build YeetFile server (see next section)
4. Run YeetFile server: `./yeetfile-server`

### Building

#### Server

`make backend`

#### CLI

`make cli`

### Environment Variables

All environment variables can be defined in a file named `.env` at the root level of the repo.

#### General Environment Variables

| Name | Purpose | Default Value | Accepted Values |
| -- | -- | -- | -- |
| `YEETFILE_HOST` | The host for running the YeetFile server | `0.0.0.0` | |
| `YEETFILE_PORT` | The port for running the YeetFile server | `8090` | |
| `YEETFILE_DEBUG` | Enable (1) or disable (0) debug mode on the server (do not use in production) | `0` | `0` or `1` |
| `YEETFILE_STORAGE` | Store files in B2 or locally on the machine running the server | `b2` | `b2` or `local` |
| `YEETFILE_DB_HOST` | The YeetFile PostgreSQL database host | `localhost` | |
| `YEETFILE_DB_PORT` | The YeetFile PostgreSQL database port | `5432` | |
| `YEETFILE_DB_USER` | The PostgreSQL user to access the YeetFile database | `postgres` | |
| `YEETFILE_DB_PASS` | The password for the PostgreSQL user | None | |
| `YEETFILE_DB_NAME` | The name of the database that YeetFile will use | `yeetfile` | |

#### Backblaze Environment Variables

These are required to be set if you want to use Backblaze B2 as your file storage backend.

| Name | Purpose |
| -- | -- |
| `B2_BUCKET_ID` | The ID of the bucket that will be used for storing uploaded content |
| `B2_BUCKET_KEY_ID` | The ID of the key used for accessing the B2 bucket |
| `B2_BUCKET_KEY` | The value of the key used for accessing the B2 bucket |

#### Misc Environment Variables

These can all be safely ignored when self-hosting, but are documented anyways since they're used
in the official YeetFile instance.

| Name | Purpose |
| -- | -- |
| `YEETFILE_SESSION_KEY` | A key to synchronize user sessions across multiple machines |
| `YEETFILE_CALLBACK_DOMAIN` | The domain to use in email correspondence |
| `YEETFILE_EMAIL_ADDR` | The email address to use for correspondence |
| `YEETFILE_EMAIL_HOST` | The host of the email address being used |
| `YEETFILE_EMAIL_PORT` | The port of the email host |
| `YEETFILE_EMAIL_PASSWORD` | The password for the email address |
| `YEETFILE_BTCPAY_API_KEY` | The API key for the BTCPay instance |
| `YEETFILE_BTCPAY_WEBHOOK_SECRET` | The webhook secret for the BTCPay instance |
| `YEETFILE_BTCPAY_STORE_ID` | The store ID within BTCPay |
| `YEETFILE_BTCPAY_SERVER_URL` | The URL of the BTCPay instance |
| `YEETFILE_BTCPAY_SUB_NOVICE_MONTHLY_LINK` | The URL for the BTCPay novice 1-month subscription |
| `YEETFILE_BTCPAY_SUB_NOVICE_YEARLY_LINK` | The URL for the BTCPay novice 1-year subscription |
| `YEETFILE_BTCPAY_SUB_REGULAR_MONTHLY_LINK` | The URL for the BTCPay regular 1-month subscription |
| `YEETFILE_BTCPAY_SUB_REGULAR_YEARLY_LINK` | The URL for the BTCPay regular 1-year subscription |
| `YEETFILE_BTCPAY_SUB_ADVANCED_MONTHLY_LINK` | The URL for the BTCPay advanced 1-month subscription |
| `YEETFILE_BTCPAY_SUB_ADVANCED_YEARLY_LINK` | The URL for the BTCPay advanced 1-year subscription |
| `YEETFILE_STRIPE_KEY` | The Stripe secret key |
| `YEETFILE_STRIPE_WEBHOOK_SECRET` | The Stripe webhook secret |
| `YEETFILE_STRIPE_PORTAL_LINK` | The link for users to manage their Stripe subscription |
| `YEETFILE_STRIPE_SUB_NOVICE_MONTHLY` | The product ID for the novice monthly subscription |
| `YEETFILE_STRIPE_SUB_NOVICE_MONTHLY_LINK` | The Stripe Checkout link for the novice monthly subscription |
| `YEETFILE_STRIPE_SUB_NOVICE_YEARLY` | The product ID for the novice yearly subscription |
| `YEETFILE_STRIPE_SUB_NOVICE_YEARLY_LINK` | The Stripe Checkout link for the novice yearly subscription |
| `YEETFILE_STRIPE_SUB_REGULAR_MONTHLY` | The product ID for the regular monthly subscription |
| `YEETFILE_STRIPE_SUB_REGULAR_MONTHLY_LINK` | The Stripe Checkout link for the regular monthly subscription |
| `YEETFILE_STRIPE_SUB_REGULAR_YEARLY` | The product ID for the regular yearly subscription |
| `YEETFILE_STRIPE_SUB_REGULAR_YEARLY_LINK` | The Stripe Checkout link for the regular yearly subscription |
| `YEETFILE_STRIPE_SUB_ADVANCED_MONTHLY` | The product ID for the advanced monthly subscription |
| `YEETFILE_STRIPE_SUB_ADVANCED_MONTHLY_LINK` | The Stripe Checkout link for the advanced monthly subscription |
| `YEETFILE_STRIPE_SUB_ADVANCED_YEARLY` | The product ID for the advanced yearly subscription |
| `YEETFILE_STRIPE_SUB_ADVANCED_YEARLY_LINK` | The Stripe Checkout link for the advanced yearly subscription |

## Support

For feature requests and bugs, you can [create an issue on
GitHub](https://github.com/benbusby/yeetfile/issues), or [submit
a ticket on SourceHut (account not required)](https://todo.sr.ht/~benbusby/yeetfile).

For issues related to the official YeetFile instance (https://yeetfile.com),
please email [support@yeetfile.com](mailto:support@yeetfile.com), or send a
message via Signal (see QR code below).

For security related issues, please [email me directly](mailto:contact@benbusby.com).
