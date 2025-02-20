<p align="center"> 
    <img alt="zapoli" src="./static/img/logo-transp.png" width="150" height="150" />
</p>

<h1 align="center">
A Specialized Nostr Relay For AppStores
</h1>

<br/>

Zapoli is a purpose-built Nostr relay designed for projects like [ZapStore](https://zapstore.dev/). It provides BlobStore and management (NIP-86) and implements robust access control mechanisms, allowing pubkeys to be explicitly allowed or banned.  

## Screenshot

<img alt="zapoli" src="./static/img/image.png"/>

## Features  

- **BlobStore Support**: Enables efficient storage and retrieval of blobs.  
- **Management(NIP-86)**: Allow or ban pubkeys to manage relay participation.  
- **Optimized for ZapStore**: Tailored to support projects leveraging Nostr for payments, content distribution, and more.  
- **Efficient Event Handling**: Ensures seamless Nostr event propagation while enforcing custom rules.  

## Installation  

### Setup

Here's an adapted **Setup** section considering that you'll push the base image to **GitHub Container Registry (GHCR)**, allowing users to pull and run it easily.  

---

## **Installation**  

### **Setup**  

#### **Option 1: Use Prebuilt Docker Image (Recommended)**  

The easiest way to run Zapoli is by using the prebuilt image:  

1. **Pull the latest image**  
   ```sh
   docker pull ghcr.io/dezh-tech/zapoli:latest
   ```

2. **Run Zapoli with environment variables**  
   ```sh
   docker run -d --name zapoli \
     -p 3334:3334 \
     -e RELAY_NAME="zapoli" \
     -e RELAY_PUBKEY="your_pubkey" \
     -e RELAY_DESCRIPTION="A Nostr relay" \
     -e RELAY_URL="wss://abc.com" \
     -e RELAY_ICON="https://your-icon-url.png" \
     -e RELAY_BANNER="https://your-banner-url.png" \
     -e RELAY_CONTACT="https://dezh.tech" \
     -e WORKING_DIR="zapoli_wd/" \
     -e RELAY_PORT=":3334" \
     -e BLOSSOM_PORT=":3334" \
     ghcr.io/dezh-tech/zapoli:latest
   ```

---

#### **Option 2: Using Docker Compose**  

For a more structured deployment, use **Docker Compose**:  

1. **Create `docker-compose.yml`**  

   ```yaml
   version: '3.8'

   services:
     zapoli:
       image: ghcr.io/dezh-tech/zapoli:latest
       container_name: zapoli
       restart: unless-stopped
       ports:
         - "3334:3334"
       environment:
         RELAY_NAME: "zapoli"
         RELAY_PUBKEY: "your_pubkey"
         RELAY_DESCRIPTION: "A Nostr relay"
         RELAY_URL: "wss://abc.com"
         RELAY_ICON: "https://your-icon-url.png"
         RELAY_BANNER: "https://your-banner-url.png"
         RELAY_CONTACT: "https://dezh.tech"
         WORKING_DIR: "zapoli_wd/"
         RELAY_PORT: ":3334"
         BLOSSOM_PORT: ":3334"
   ```

2. **Run with Compose**  
   ```sh
   docker-compose up -d
   ```

## Configuration  

Modify the `env` variables in `.env` file, docker compose file or docker command to customize settings:  

### Relay Metadata  

- `RELAY_NAME` – The name of the relay (default: `zapoli`).  
- `RELAY_PUBKEY` – The owner's hex key (convert `npub` to hex [here](https://nostrcheck.me/converter/)).  
- `RELAY_DESCRIPTION` – A short description of the relay.  
- `RELAY_URL` – WebSocket URL for the relay (e.g., `wss://abc.com`).  
- `RELAY_ICON` – URL to the relay's icon.  
- `RELAY_BANNER` – URL to the relay's banner image.  
- `RELAY_CONTACT` – Contact URL (e.g., `https://dezh.tech`).  

### Storage & Working Directory  

- `WORKING_DIR` – Configuration working directory (default: `zapoli_wd`).  

### Networking & Ports  

- `RELAY_PORT` – Port on which the relay listens (default: `:3334`).  
- `BLOSSOM_PORT` – Port for Blossom (default: `:3334`).  

### Pubkey Access Control  

- `ALLOWED_PUBKEYS` – Comma-separated list of allowed public keys.  
- `BANNED_PUBKEYS` – Comma-separated list of banned public keys.  

## API & Usage  

Zapoli follows standard Nostr relay behavior while integrating additional management features to be optimize and efficient for app stores on Nostr.


## Contributing  

Pull requests are welcome! Feel free to open an issue if you have feature requests or find bugs.  

## License  

This software is published under [MIT License](../LICENSE).
