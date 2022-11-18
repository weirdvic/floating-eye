![Build](https://github.com/weirdvic/floating-eye/actions/workflows/main.yml/badge.svg)
# Описание
Вспомогательный Telegram бот для русскоязычного чата игры NetHack.

Фактически большая часть команд реализована через запросы к ботам Beholder, Pinobot и Croesus, находящимся в IRC Libera chat.
При запросе информации о монстрах, умеет показывать изображения монстров и ссылку на NetHack Wiki.

Команда !pom (phase of moon), отображает текущую фазу Луны и её влияние на NetHack. Принцип работы этой команды аналогичен виджету с изображением луны на [NAO](https://alt.org/nethack/), а код для вычисления фазы Луны основан на коде, использующемся в самом NetHack.

Также бот пересылает в чат в Telegram сообщения от бота `Beholder` в IRC канале `#hardfought` с упоминаниями определённых игроков. Список отслеживаемых игроков задаётся массивом строк `watch_players` в файле `config.json`

Действующий экземпляр бота доступен в Telegram как [__@floatingeye_bot__](https://t.me/floatingeye_bot)
