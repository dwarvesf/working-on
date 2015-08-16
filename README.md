# working-on
Working On app that help the team improve the productivity

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/dwarvesf/working-on)

## Setup

### Digest for public channel

* Deploy Working On app

    - Clone source code
    - Deploy to your server with Go and MongoDB installed
    - For Heroku: Add MongoLab add-on
    - Set `BOT_TOKEN`

* Add Bot

    - Add new integration: Bots. The bot will post the digest and remind messages everyday.
    - Retrieve API Token
    - Back to Slack and invite the bot to the channel

* Add Slash Command

    - Add new integration: Slash Commands.
    - Retrieve the Token
    - Add url `<your-host>/on`

* Configure cli (for dev)

    - Clone project [working-on-cli]()
    - Access https://api.slack.com/web to get your token.
    - Run ./setup.sh --token `<token>` --domain `<domain>` will create bin file and config file for you.

### Digest for project channel

_Not yet supported_

## Roadmap

- [ ] Support private channel
- [ ] One click installer for backend
- [ ] Support multi Slack teams

