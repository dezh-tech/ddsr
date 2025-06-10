<p align="center"> 
    <img alt="nwclay" src="./static/img/logo-transp.png" width="150" height="150" />
</p>

<h1 align="center">
The NWCLay only accepts NWC events!
</h1>

<br/>

## Screenshot

<img alt="nwclay" src="./static/img/ss.png"/>

## Features

- **Limits**: Only accept kinds from NWC spec.
- **Database**: Keeps event for a while on database for reliability.
- **Filter checking**: Only accept valid filters with at least authors or #p and only kinds matching NWC.

## Installation

### Setup

Here's an adapted **Setup** section, considering that you'll push the base image to **Docker Registry**, allowing users to pull and run it easily.

---

## **Installation**

### **Setup**

#### **Option 1: Use Prebuilt Docker Image (Recommended)**

The easiest way to run NWCLay is by using the prebuilt image:

1. **Pull the latest image**

   ```sh
   docker pull dezhtech/nwclay
   ```

2. **Run 210Maxi with environment variables**
   ```sh
   docker run -d --name nwclay \
   -p 3334:3334 \
   -e RELAY_NAME="nwclay" \
   -e RELAY_PUBKEY="your_pubkey" \
   -e RELAY_DESCRIPTION="Only accepts NWC events" \
   -e RELAY_URL="wss://nwclay.com" \
   -e RELAY_ICON="https://your-icon-url.png" \
   -e RELAY_BANNER="https://your-banner-url.png" \
   -e RELAY_CONTACT="https://dezh.tech" \
   -e RELAY_PORT=":3334" \
   -e WORKING_DIR="nwclay_wd/" \
   -e KEEP_IN_MINUTES=10 \
   -e ACCEPT_WINDOW_IN_MINUTES=1 \
   dezhtech/nwclay
   ```

---

#### **Option 2: Using Docker Compose**

For a more structured deployment, use **Docker Compose**:

1. **use `compose.yml`**

use the existing compose file in the NWCLay directory


2. **Run with Compose**
   ```sh
   docker-compose up -d
   ```

## Configuration

Modify the `env` variables in the `.env` file, docker compose file, or docker command to customize settings:

### Relay Metadata

- `RELAY_NAME` – The name of the relay (default: `NWCLay`).
- `RELAY_PUBKEY` – The owner's hex key (convert `npub` to hex [here](https://nostrcheck.me/converter/)).
- `RELAY_DESCRIPTION` – A short description of the relay.
- `RELAY_URL` – WebSocket URL for the relay (e.g., `wss://abc.com`).
- `RELAY_ICON` – URL to the relay's icon.
- `RELAY_BANNER` – URL to the relay's banner image.
- `RELAY_CONTACT` – Contact URL (e.g., `https://dezh.tech`).

### Database And Events Config

- `KEEP_IN_MINUTES` – Remove events that are KEEP_IN_MINUTES old. (default: `10 minutes`)
- `ACCEPT_WINDOW_IN_MINUTES` – Only accept events from KEEP_IN_MINUTES past or future. (default: `1 minute`)

### Storage & Working Directory

- `WORKING_DIR` – Configuration working directory (default: `nwclay_wd`).

### Networking & Ports

- `RELAY_PORT` – Port on which the relay listens (default: `:2100`).

## Contributing

Pull requests are welcome! Feel free to open an issue if you have feature requests or find bugs.

## License

This software is published under [MIT License](../LICENSE).
