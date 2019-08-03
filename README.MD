# Блог

## Сервер

## Быстрый старт

### Запуск сервера
1) Установите следующие переменные окружения в *docker-compose-local.yml*

- *DB_USER* - имя пользователя базы данных
- *DB_PASSWORD* - пароль пользователя базы данных
- *DB_NAME* - имя базы данных
- *DB_HOST* - хост базы данных
- *DB_PORT* - порт базы данных
- *JWT_SECRET_KEY* - секретный ключ для подписания JWT токенов
- *SERVER_PORT* - порт для запуска сервера
 
2) Выполните на базе данных SQL скрипты из папки **sql-scripts**
3) Запустите сервер:
```
docker-compose -f docker-compose-local.yml up -d
```
4) Зарегистрируйте пользователя на странице **/register**
5) Дайте данному пользователю права админа, записав его имя в список админов в переменной окружения *ADMINS*
6) Остановите и запустите сервер снова для применения обновлений списка админов

---

### Интеграция Traefik
После того, как сервер настроен, мы можем интегрировать Traefik. Traefik позволит нам запускать несколько серверов и производить автоматический Load Balancing, а также защитить HTTP соединение с помощью применения Let's Encrypt TSL сертификата.

Необходимые для работы Traefik файлы находятся в папке **traefik**. Вы можете переместить эту папку в любое удобное для вас место, например - `/opt/traefik`.

1) Создайте нового пользователя с именем *admin* для последующего доступа к Traefik Dashboard:
```
htpasswd -nbB admin <password>
```

2) Установите вывод данной команды в файле *traefik.toml*:
```toml
[entryPoints]
  [entryPoints.dashboard]
    address = ":8081"
    [entryPoints.dashboard.auth]
      [entryPoints.dashboard.auth.basic]
        users = ["<insert here>"]
```

3) Установите имя вашего домена:
```toml
[docker]
domain = "<domain>"
```

4) Установите вашу почту:
```toml
[acme]
email = "<email>"
```

5) Откройте файл *docker-compose.yml* и установите ваш домен в пути для доступа к Traefik Dashboard:
```yaml
labels: 
  - "traefik.frontend.rule=Host:monitor.<domain>"
```

7) Создайте Docker сеть для Traefik. Все контейнеры, которые заиспользуют данную сеть, будут доступны из интернета.
```
docker network create web
```

8) Запустите Traefik контейнер.

Из папки с Traefik:
```
docker-compose up -d
```

9) Запустите сервер. 

Для запуска сервера с поддержкой Traefik используется другой docker-compose файл - *docker-compose.yml*. В нем должны быть установлены те же самые переменные окружения, как и в случае запуска сервера без Traefik.

Однако, мы также должны установить адрес, по которому будет доступен сервер. Откройте *docker-compose.yml* и установите следующий Docker label:
```yaml
labels:
  - "traefik.frontend.rule=Host:<domain>"
```

Теперь запустите сервер:
```
docker-compose up -d
```

Теперь вы можете мониторить ваши контейнеры из Traefik Dashboard, зайти в который вы сможете через HTTP Basic аутентификацию с помощью раннее созданного пользователя *admin* по пути `monitor.<domain>`. Сервер же доступен по адресу `<domain>`.
Например, пусть название домена - *example.com*, тогда Traefik Dashboard находится по адресу `monitor.example.com`, а сервер - `example.com`.