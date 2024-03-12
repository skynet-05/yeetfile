<div align="center">
  <img style="width: 500px; padding-bottom: 20px;" src="https://github.com/meddlehead/yeetfile/blob/main/web/static/img/yeetfile-logo.png?raw=true">
  <br><br>

  <p>A private and secure temporary file + text sharing service</p>

[![Latest Release](https://img.shields.io/github/v/release/meddlehead/yeetfile)](https://github.com/meddlehead/yeetfile/releases)
[![License: AGPLv3](https://img.shields.io/github/license/meddlehead/yeetfile)](https://opensource.org/license/agpl-v3)

<table>
  <tr>
    <td><a href="https://sr.ht/~benbusby/yeetfile">SourceHut</a></td>
    <td><a href="https://github.com/meddlehead/yeetfile">GitHub</a></td>
  </tr>
</table>
</div>
<hr>

Contents
1. [About](#about)
    1. [Features](#features)
2. [How It Works](#how-it-works)
    1. [Uploading and Downloading](#uploading-and-downloading)
    2. [Billing](#billing)
3. [Development / Self-Hosting](#development--self-hosting)
    1. [Requirements](#requirements)
    2. [Setup](#setup)
    3. [Building](#building)
    4. [Environment Variables](#environment-variables)

## About

YeetFile is a text and file transferring service, with both a [web](https://yeetfile.com) and
[CLI client](https://github.com/meddlehead/yeetfile/releases) officially supported.

### Features

- Configurable upload settings
    - Expiration date/time
        - Configurable to N minutes/hours/days
        - Max 30 days
    - Number of downloads
        - Max 10 downloads
    - Password protection (optional)
- Free text transfers (up to 1K characters)
- Low cost (<$0.20/GB) file transfers
    - No upload size limit
    - Subscription and non-subscription options available
    - Options to pay via Stripe or BTCPay
        - BTC and XMR accepted
- Email not required at signup
    - Account ID-only signup allowed
    - Email only used for billing notifications
    - Signup not required for text-only transfers
- Easy self-hosting, if desired

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

## Development / Self-Hosting

Whether you're wanting to help with development or want to self-host your own instance of YeetFile, the process
is fairly simple and designed to be relatively low effort.

### Requirements

- Go v1.20+
- PostgreSQL 15+
- A Backblaze B2 account (optional)

### Setup

1. Start PostgreSQL
2. Create a database in PostgreSQL named "yeetfile" using user "postgres"
   1. This can be customized to your preference, just use the `YEETFILE_DB_*` environment variables outlined
      below to configure before launching the server.
3. Build YeetFile server (see next section)
4. Run YeetFile server: `./yeetfile-web`

### Building

#### Server

`make web`

#### CLI

`make cli`

### Environment Variables

All environment variables can be defined in a file named `.env` at the root level of the repo.

| Name | Purpose | Default Value | Accepted Values |
| -- | -- | -- | -- |
| `YEETFILE_HOST` | The host for running the YeetFile server | `localhost` | |
| `YEETFILE_PORT` | The port for running the YeetFile server | `8090` | |
| `YEETFILE_SESSION_KEY` | A key to synchronize user sessions across multiple machines | None (randomly generated) | |
| `YEETFILE_DEBUG` | Enable (1) or disable (0) debug messages from the server | `0` | `0` or `1` |
| `YEETFILE_STORAGE` | Store files in B2 or locally on the machine running the server | `b2` | `b2` or `local` |
| `YEETFILE_DB_HOST` | The YeetFile PostgreSQL database host | `localhost` | |
| `YEETFILE_DB_PORT` | The YeetFile PostgreSQL database port | `5432` | |
| `YEETFILE_DB_USER` | The PostgreSQL user to access the YeetFile database | `postgres` | |
| `YEETFILE_DB_PASS` | The password for the PostgreSQL user | None | |
| `YEETFILE_DB_NAME` | The name of the database that YeetFile will use | `yeetfile` | |

If you're wanting to use Backblaze B2 for storing uploaded content, you'll need to set the following variables as well:

| Name | Purpose |
| -- | -- |
| `B2_BUCKET_ID` | The ID of the bucket that will be used for storing uploaded content |
| `B2_BUCKET_KEY_ID` | The ID of the key used for accessing the B2 bucket |
| `B2_BUCKET_KEY` | The value of the key used for accessing the B2 bucket |
