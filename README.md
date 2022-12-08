# Youtube Stalker Bot

## Что делает?
Ищет случайные видео на [youtube.com](https://www.youtube.com/)

## Как?
Видео на YouTube имеет идентификатор из 11 символов base64, который передётся после знака вопроса в параметре `v`

Например: `https://www.youtube.com/watch?v=8ybmEKq-9BE`

Тут мы имеем параметр `v` равный `8ybmEKq-9BE`

Бот генерирует совершенно случайным образом только 5 символов из 11. 

Например `yw9um` (мы можем убрать _большие_ (прописные, заглавные - кому как нравится) буквы, так как при поиске размер букв не учитывается.

А затем обращается к Youtube API `https://www.googleapis.com/youtube/v3/search` (10 000 запросов в день бесплатно) c запросом `inurl:yw9um` - т.е. мы ищем такие видео, у которых URL будет содержать `yw9um`.

## Попробовать
Экземпляр бота доступен по адресу https://t.me/youtubestalkerbot

В ветке [F@ST2704N](https://github.com/thedmdim/youtube-stalker-bot/tree/F%40ST2704N) представлен код, который запущен на моём [старом роутере](https://openwrt.org/toh/hwdata/sagem/sagem_fast2704n_v1) Sagecom.

Чтобы запустить самому надо сделать:
1. `git clone https://github.com/thedmdim/youtube-stalker-bot`
2. `cd youtube-stalker-bot`
3. `go build .`
4. Включить Youtube API и получить токен [Google Cloud](https://console.cloud.google.com/apis/api/youtube.googleapis.com)
5. Создать бота и получить токен в Телеграме https://t.me/BotFather
6. Установить переменные окружения
   - Для **Windows**
     1. `set GCLOUD_TOKEN=<Google Cloud Token>`
     2. `set TGBOT_TOKEN=<Telegram Bot Token>`
   - Для **Linux**
     1. `export GCLOUD_TOKEN=<Google Cloud Token>`
     2. `export TGBOT_TOKEN=<Telegram Bot Token>`
7. Запустить: `./youtube-stalker-bot` если вы на Linux или `youtube-stalker-bot.exe` если вы на Windows

