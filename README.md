# working-on
Working On app that help the team improve the productivity

## Setup

### Digest for public channel

0. Deploy Working On app

- Clone source code
- Deploy to your server with Go and MongoDB installed
- For Heroku: Add MongoLab add-on
- Set `SLACK_BOT_TOKEN`


1. Add Bot

- Add new integration: Bots. The bot will post the digest and remind messages everyday.
- Retrieve API Token
- Back to Slack and invite the bot to the channel

2. Add Slash Command

- Add new integration: Slash Commands.
- Retrieve the Token
- Add url `<your-host>/d`

3. Configure cli

- Clone project [working-on-cli]()
- Access https://api.slack.com/web to get your token.
- Run ./setup.sh `<token>` will create bin file and config file for you.

### Digest for project channel

_Not yet supported_

## Roadmap

- [ ] Support private channel
- [ ] One click installer for backend
- [ ] Support multi Slack teams

