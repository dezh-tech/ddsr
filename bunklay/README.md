<p align="center"> 
    <img alt="bunklay" src="./static/img/210maxi-logo-transparent.png" width="150" height="150" />
</p>

<h1 align="center">
The Bunklay only accepts bunker event!
</h1>

<br/>

## Features

- **Limits**: Only accept kind 24133.

## Installation

### Setup

Here's an adapted **Setup** section considering that you'll push the base image to **Docker Registry**, allowing users to pull and run it easily.

---

## **Installation**

### **Setup**

#### **Option 1: Use Prebuilt Docker Image (Recommended)**

The easiest way to run Bunklay is by using the prebuilt image:

1. **Pull the latest image**

   ```sh
   docker pull dezhtech/bunklay
   ```

2. **Run 210Maxi with environment variables**
   ```sh
   docker run -d --name bunklay \
   -p 3334:3334 \
   -e RELAY_NAME="bunklay" \
   -e RELAY_PUBKEY="your_pubkey" \
   -e RELAY_DESCRIPTION="Only accepts bunker events" \
   -e RELAY_URL="wss://bunklay.com" \
   -e RELAY_ICON="https://your-icon-url.png" \
   -e RELAY_BANNER="https://your-banner-url.png" \
   -e RELAY_CONTACT="https://dezh.tech" \
   -e RELAY_PORT=":3334" \
   dezhtech/bunklay
   ```

---

#### **Option 2: Using Docker Compose**

For a more structured deployment, use **Docker Compose**:

1. **use `compose.yml`**

use the exist compose file in the Bunklay directory


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

### Networking & Ports

- `RELAY_PORT` – Port on which the relay listens (default: `:2100`).

## Contributing

Pull requests are welcome! Feel free to open an issue if you have feature requests or find bugs.

## License

This software is published under [MIT License](../LICENSE).
