<p align="center"> 
    <img alt="ddsr" src="./.images/ddsr.png" width="150" height="150" />
</p>

<h1 align="center">
Domain-Specific Relays (DSR)
</h1>

<br/>

This repository contains a set of [Nostr](https://nostr.com) relays designed to be used for specific purposes.

This project is based on [Khatru](https://github.com/fiatjaf/khatru), [Event Store](https://github.com/fiatjaf/eventstore), [Blob Store](github.com/kehiy/blobstore) and [Go Nostr](github.com/nbd-wtf/go-nostr).

Here is a list of DSRs developed by the Dezh:

1. [Zapoli](./zapoli): A relay designed to used by [NIP-82](https://github.com/nostr-protocol/nips/pull/1336) clients.
    - **NIP-50**: You can search between multiple software applications.
    - **Blossom**: You can store software images, icons, and binaries.
    - **NIP-86**: You can limit write access to software publishers.

>[!NOTE]
> We suggest using [Zapstore relay](https://github.com/zapstore/relay) which is designed for the same purpose.

2. [210maxi](./210maxi): A relay that only accepts 210-character events, tuned for [NIP-B1](https://github.com/nostr-protocol/nips/pull/1710) feeds.
    - **Limits**: Only accept kind 25, 1111, 7, 5, 9734, 9735, 0, 3.
    - **NIP-50**: search your favorite tiny notes.
    - **NIP-86**: manageable with the support of reporting. (WIP)

3. [Pages](./pages): A relay that only keeps profiles and follow lists. You can simply resolve any pubkey from it. 
    - **NIP-50**: You can search profiles.
    - **Control**: You can directly send your profile update/deletion to it.
    - **Discovery**: It scrapes new profiles.
    - **Admin and Moderators**: - Admins can call management APIs (NIP-86) and moderators can send reporting events (kind 1984) to remove a profile from relay.

4. [Bunklay](./bunklay): A relay that only accepts bunker-related events.
    - **Limits**: Only accept kind 24133.
    - **Database**: Keeps events for a while in the database for reliability.
    - **Filter checking**: Only accept valid filters with at least one author or #p and only keywords matching the Bunker protocol.

5. [NWCLay](./nwclay/): A relay that only accepts NWC-related events.
    - **Limits**: Only accept kinds from the NWC spec.
    - **Database**: Keeps events for a while in the  database for reliability.
    - **Filter checking**: Only accept valid filters with at least authors or #p and only kinds matching NWC.

6. [Chapar](./chapar/): A relay that only accepts chat app messages.
    - **Limited Kinds**: Only accepts kinds related and accepted in NIP-59.
    - **Limited queries**: You need to authenticate before reading any events, and you can only read events related to you.

> [!NOTE]
> You can open your target relay and find full documentation there.

## Contribution

All kinds of contributions are welcome!

## Donation

Donations and financial support for the development process are possible using Bitcoin and Lightning:

**on-chain**:

```
bc1qfw30k9ztahppatweycnll05rzmrn6u07slehmc
```

**lightning**: 

```
donate@dezh.tech
```

## License

This software is published under [MIT License](./LICENSE)
