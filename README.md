# working-on
Working On is a productivity tool that integrated with Slack, the most popular team collaboration platform nowadays. At Dwarves Foundation, teamwork is about [*Synchronisation*](http://tieubao.me/writing/2014/12/05/it-is-hard-to-become-a-team-member/) between the team members, which means you need to know and align with the team goal, and all of your activities need to be known by other members. That will raise the awareness to the next level and keep the team synced.

We had successfully apply it to our team and we want to share it to you. Too simple to setup via Heroku.

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/dwarvesf/working-on)

## How to use

- When you start to do something, go to Slack and use slash command `/working <what are you going to do>` to let your teammates know about it. (The geek can use `cli`)
- On the next day morning, the bot will make the digest and post it to the digest channel, so that everyone else can have a full view, even the manager or leader. It's also make others motivated by seeing what you've achieved.
- All the team members should follow the rule for the team sake.

*What does it look like*

![Sample](/static/sample.png)

## Setup

### Digest for public channel

* Deploy Working On app

    - Clone source code
    - Deploy to your server with Go and MongoDB installed
    - For Heroku: Add MongoLab add-on.
    - For Heroku: Add NewRelic add-on.

    - Set env `MONGOLAB_URI` which is database url.
    - Set env `DB_NAME` which is the name of the database, the last part of *MONGOLAB_URI*
    - Add digest time as env `DIGEST_TIME` in UTC +0. Mine is "02:30", which means we will have a digest message on 9:30 AM GMT+7
    - Add digest channel as env `DIGEST_CHANNEL`. Mine is "#general"
    - Add working channel as env `WORKING_CHANNEL`. Mine is "#working"

    ![Heroku Env](/static/heroku-env.png)

* Add Bot

    ![Add bot](/static/bot.png)

    - Add new integration: Bots. The bot will post the digest and remind messages everyday.
    - Retrieve API Token. Set env `BOT_TOKEN`
    - Back to Slack and invite the bot to the channel.

* Add Slash Command

    ![Add slash command](/static/slash.png)

    - Add new integration: Slash Commands.
    - Retrieve the Token. Set env `SLASH_TOKEN`
    - Add url `<your-host>/on`. For Heroku, it is `http://xyz.herokuapp.com/on`

* Setup NewRelic (to keep your Heroku server awake)

    - Add NewRelic add-on for Heroku or you can register one for yourself
    - Set env `NEW_RELIC_LICENSE_KEY` and `NEW_RELIC_LOG`
    - Setup ping for your Heroku server

    ![NewRelic](/static/newrelic.png)

### Configure cli (for geek)

- Clone project [working-on-cli](https://github.com/dwarvesf/working-on-cli)
- Access https://api.slack.com/web to get own your token.
- Run ./setup.sh --token `<token>` --domain `<domain>` will create bin file and config file for you.

### Digest for project channel

_Not yet supported_

## Roadmap

- [ ] Support private channel
- [x] One click installer for backend
- [ ] Support multi Slack teams
- [ ] Restructure

## Contribution

Team collaboration and synchronization are always the main factors of the good team. We want to build a simple tool that can integrate with Slack, help the team increase productivity and also fit for remote team. Feel free to contribute any idea or pull request to make it better.

