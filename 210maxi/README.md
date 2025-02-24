<p align="center"> 
    <img alt="210maxi" src="./static/img/210maxi-logo-transparent.png" width="150" height="150" />
</p>

<h1 align="center">
A nostr relay that only accepts 210 character events.
</h1>

<br/>

The 210Maxi relay is a relay that forces events content to be at most 210 chars. Which is suitable for NIP-B1 feeds.

## Screenshot

<img alt="210maxi" src="./static/img/screenshot.png"/>

## Features

- **Limited Kinds**: Only accepts kinds related and accepted in NIP-B1.
- **NIP-50: Search**: Allow to search on 210 char events.
- **Limited chars**: it checks each kind 25 and 1111 to have at most 210 char (not bytes).

## Installation

### Setup

Here's an adapted **Setup** section considering that you'll push the base image to **Docker Registry**, allowing users to pull and run it easily.

---

## **Installation**

### **Setup**

#### **Option 1: Use Prebuilt Docker Image (Recommended)**

The easiest way to run 210Maxi is by using the prebuilt image:

1. **Pull the latest image**

   ```sh
   docker pull dezhtech/210maxi
   ```

2. **Run 210Maxi with environment variables**
   ```sh
   docker run -d --name zapoli \
   -p 3334:3334 \
   -e RELAY_NAME="21maxi" \
   -e RELAY_PUBKEY="your_pubkey" \
   -e RELAY_DESCRIPTION="Only accepts 210 char events" \
   -e RELAY_URL="wss://210maxi.com" \
   -e RELAY_ICON="https://your-icon-url.png" \
   -e RELAY_BANNER="https://your-banner-url.png" \
   -e RELAY_CONTACT="https://dezh.tech" \
   -e WORKING_DIR="210maxi_wd/" \
   -e RELAY_PORT=":2100" \
   dezhtech/210maxi
   ```

---

#### **Option 2: Using Docker Compose**

For a more structured deployment, use **Docker Compose**:

1. **use `compose.yml`**

use the exist compose file in the 210Maxi directory


2. **Run with Compose**
   ```sh
   docker-compose up -d
   ```

## Configuration

Modify the `env` variables in `.env` file, docker compose file or docker command to customize settings:

### Relay Metadata

- `RELAY_NAME` – The name of the relay (default: `210Maxi`).
- `RELAY_PUBKEY` – The owner's hex key (convert `npub` to hex [here](https://nostrcheck.me/converter/)).
- `RELAY_DESCRIPTION` – A short description of the relay.
- `RELAY_URL` – WebSocket URL for the relay (e.g., `wss://abc.com`).
- `RELAY_ICON` – URL to the relay's icon.
- `RELAY_BANNER` – URL to the relay's banner image.
- `RELAY_CONTACT` – Contact URL (e.g., `https://dezh.tech`).

### Storage & Working Directory

- `WORKING_DIR` – Configuration working directory (default: `210maxi_wd`).

### Networking & Ports

- `RELAY_PORT` – Port on which the relay listens (default: `:2100`).

## Contributing

Pull requests are welcome! Feel free to open an issue if you have feature requests or find bugs.

## License

This software is published under [MIT License](../LICENSE).
