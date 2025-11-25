# Сервис авторизации.

Базово может:

-Регистрировать новых пользователей

-Авторизовывать

-Обновлять токены доступа

-Отзывать токены обновления


Что нужно доработать:

-Не валидируется почта пользователя

-rate limit

-Восстановление пароля

Работает на порте 8080, позже можно будет менять через файл окружения

# Хендлеры

# /register(POST)

Регистрирует нового пользователя.
Если пользователь уже существует вернет ошибку 409.


Тело запроса

```json
{
    "email":"",
    "password":""
}
```

# /login(POST)

Проверяет пользователя в базе и возвращает либо токены, либо ошибку.

Тело запроса

```json
{
    "email":"",
    "password":""
}
```

Тело ответа
```json
{
    "access_token":"",
    "refresh_token":""
}
```
# /refresh(POST)

Проверяет наличие в кеше соответствующего токена.
Возвращает новую пару access + refresh токенов, либо ошибку

Тело запроса

```json
{
    "refresh_token":""
}
```

Тело ответа

```json
{
    "access_token":"",
    "refresh_token":""
}
```

# /logout(POST)

Удаляет в базе refresh токен

Тело запроса

```json
{
    "refresh_token":""
}
```

# Запуск локально
Для этого потребуется утилита migrate

Установка с Ubuntu

```bash
curl -s https://packagecloud.io/install/repositories/golang-migrate/migrate/script.deb.sh | sudo bash.
```

```bash
sudo apt-get update.
```

```bash
sudo apt-get install migrate.
```

Проверяем

```bash
migrate --version
```

В makefile предоставлены команды для успешной сборки

1. Требуется поднять базы через docker-compose

```bash
make compose
```
2. Проинициализировать таблицу users в postgres
```bash
make migrate
```

3. Собрать исполняемый файл 
```bash
make build
```

4. Запустить сервер
```bash
make run
```


