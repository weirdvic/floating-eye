![Build](https://github.com/weirdvic/floating-eye/actions/workflows/main.yml/badge.svg)
# Описание
Простой Telegram бот для игры NetHack.

Фактически большая часть команд реализована через запросы к ботам Beholder и Pinobot, находящимся в IRC [Libera chat](https://libera.chat/).

При запросе информации о монстрах, бот умеет показывать изображения монстров в виде увеличенных тайлов.

Команда `!pom` (phase of moon), отображает текущую фазу луны и её влияние на NetHack.

Также бот пересылает сообщения от бота `Beholder` в IRC канале `#hardfought`, в которых упоминаются русскоязычные игроки. Список отслеживаемых игроков задаётся массивом строк `watch_players` в файле `config.json`

Действующий экземпляр бота доступен в Telegram как [__@floatingeye_bot__](https://t.me/floatingeye_bot)
