# tgbot

## Description

Telegram bot

## Usage

```
$ tgbot
usage: tgbot config
```

## Config format

The following snippet shows a typical config file. A
complete example can be found at doc/global.cfg.

```toml
TgBin = "/path/to/telegram-cli"
TgPubKey = "/path/to/tg-server.pub"
MinOutput = "/path/to/minoutput.lua"
Chat = "ChatName"

[Echo]
Enabled = true

[Quotes]
Enabled = true
Endpoint = "https://example.com:8001/"
User = "user"
Password = "s3cr3t"

...
```

## Installation

`go get github.com/jroimartin/tgbot`

## Requirements

* [telegram-cli](https://github.com/vysheng/tg)
