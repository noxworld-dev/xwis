# Nox XWIS tools

Tools for working with XWIS lobby servers. Tailored specifically for Nox.

## Installation

```bash
go install ./cmd/xwis
```

## Listing games

```bash
$ xwis list

Total rooms: 5
        1:NoXWorld.ru   0/31
        Daybreak        2/32
        Korean Ladder   5/32
        NoxCommunity EU 0/29
        Sephira Serve   0/13
```

## Registering a game

```bash
$ cp xwis-game-example.json xwis-game.json
$ xwis register

Hosting game: "My Server" on "mymap" (arena)
```