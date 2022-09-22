![Build](https://github.com/weirdvic/floating-eye/actions/workflows/main.yml/badge.svg)
# Описание
Простой Telegram бот для игры NetHack.

Фактически большая часть команд реализована через запросы к ботам Beholder и Pinobot, находящимся в IRC Libera chat.
При запросе информации о монстрах, умеет показывать изображения монстров в виде увеличенных тайлов.

Есть команда !pom (phase of moon), отображающая текущую фазу Луны и её влияние на NetHack. Принцип работы этой команды аналогичен виджету с изображением луны на [NAO](https://alt.org/nethack/), а код для вычисления фазы Луны основан на коде, использующемся в самом NetHack.

Также бот пересылает сообщения от бота `Beholder` в IRC канале `#hardfought`, в которых упоминаются русскоязычные игроки. Список отслеживаемых игроков задаётся массивом строк `watch_players` в файле `config.json`

Действующий экземпляр бота доступен в Telegram как [__@floatingeye_bot__](https://t.me/floatingeye_bot)
## Список отслеживаемых игроков на HDF:

- wvc
- lacca
- maxlunar
- MaxLunar
- engelson
- liferooter
- nazult
- yen0
