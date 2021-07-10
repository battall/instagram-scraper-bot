# Instagram Scraper Bot

- do not check "is_disabled" info with sessions. because accounts can block account, but can't block public (anonymous requests).
- consider accounts can block someone, or blocked by someone which is in chceking,

BotManager is a class for managing;
 - MongoDB
 - Instagram api
 - Login & check accounts
 - Save medias
 - etc

## Unnecessary stuff

in database
  medias - feed_type
    0 media
    1 reel
    2 profile pic
  medias - media_type
    1 jpg
    2 mp4
    8 multiple
  logs - type
    1 add
    2 delete
    3 edit
  users - last_checked
    0 is_checking
    1 info time
    2 feed time
    3 reel time
