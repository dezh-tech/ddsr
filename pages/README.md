<p align="center"> 
    <img alt="pages" src="./static/img/logo-transp.png" width="150" height="150" />
</p>

<h1 align="center">
A Nostr relay only for profiles!
</h1>

<br/>

## Screenshot

<img alt="pages" src="./static/img/ss.png"/>

## Features

- NIP-50: you can search profiles.
- You can directly send your profile update/deletion to it.
- It scrape new profiles.
- Admins can call management APIs and moderators can send reporting events (kind 1984) to remove a profile from relay.
- NIP-86: Ban specific profiles/Check reported ones.


## Installation

### Setup

Here's an adapted **Setup** section considering that you'll push the base image to **Docker Registry**, allowing users to pull and run it easily.

---

## **Installation**

### **Setup**

#### **Option 1: Use Prebuilt Docker Image (Recommended)**

The easiest way to run Pages is by using the prebuilt image:

1. **Pull the latest image**

   ```sh
   docker pull dezhtech/pages
   ```

2. **Run Pages with environment variables**
   ```sh
   docker run -d --name pages \
   -p 3334:3334 \
   -e RELAY_NAME="pages" \
   -e RELAY_PUBKEY="your_pubkey" \
   -e RELAY_DESCRIPTION="Specialized Nostr Relay For AppStores" \
   -e RELAY_URL="wss://jellyfish.land" \
   -e RELAY_ICON="https://your-icon-url.png" \
   -e RELAY_BANNER="https://your-banner-url.png" \
   -e RELAY_CONTACT="https://dezh.tech" \
   -e WORKING_DIR="pages_wd/" \
   -e RELAY_PORT=":3334" \
   -e ADMIN_PUBKEYS="" \
   -e MODERATOR_PUBKEYS="" \
   -e DISC_RELAYS="nos.lol,purplepag.es,relay.nostr.lol,jellyfish.land,relay.primal.net,nostr.mom,nostr.wine,nostr.land" \
   dezhtech/pages
   ```

---

#### **Option 2: Using Docker Compose**

For a more structured deployment, use **Docker Compose**:

1. **use `compose.yml`**
use the exist compose file in the pages directory


2. **Run with Compose**
   ```sh
   docker-compose up -d
   ```

## Configuration

Modify the `env` variables in `.env` file, docker compose file or docker command to customize settings:

### Relay Metadata

- `RELAY_NAME` – The name of the relay (default: `pages`).
- `RELAY_PUBKEY` – The owner's hex key (convert `npub` to hex [here](https://nostrcheck.me/converter/)).
- `RELAY_DESCRIPTION` – A short description of the relay.
- `RELAY_URL` – WebSocket URL for the relay (e.g., `wss://abc.com`).
- `RELAY_ICON` – URL to the relay's icon.
- `RELAY_BANNER` – URL to the relay's banner image.
- `RELAY_CONTACT` – Contact URL (e.g., `https://dezh.tech`).

### Storage & Working Directory

- `WORKING_DIR` – Configuration working directory (default: `pages_wd`).

### Networking & Ports

- `RELAY_PORT` – Port on which the relay listens (default: `:3334`).

### Admin Access Control

- `ADMIN_PUBKEYS` – Comma-separated list of allowed public keys that can call management APIs.
- `MODERATORS_PUBKEYS` – Comma-separated list of allowed public keys that can remove events from db by sending kind 1984 events.

## Contributing

Pull requests are welcome! Feel free to open an issue if you have feature requests or find bugs.

## License

This software is published under [MIT License](../LICENSE).
