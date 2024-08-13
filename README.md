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
1. [How It Works / Security](#how-it-works--security)
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

## How It Works / Security

See: [https://docs.yeetfile.com/security](https://docs.yeetfile.com/security/)

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
