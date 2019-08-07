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
- *DOMAIN* - домен для сайта. Требуется указывать в полной форме: *scheme://host*. Например: `https://example.com`
- *SERVER_PORT* - порт для запуска сервера
 
2) Выполните на базе данных SQL скрипты из папки **sql-scripts**
3) Запустите сервер:
```
docker-compose -f local-docker-compose.yml up -d
```

Данный способ не рекомендуется к использованию, так как в такой конфигурации не защищена админка.

---

### Интеграция Traefik

Traefik позволит нам запускать несколько серверов и производить автоматический Load Balancing, защитить HTTP соединение с помощью применения Let's Encrypt TSL сертификата, а также защитить секьюрные пути с помощью HTTP Basic авторизации. Сервер имеет также логику авторизации и аутентификации по JWT токенам, но данная логика отключена.

Необходимые для работы Traefik файлы находятся в папке **traefik**. Вы можете переместить эту папку в любое удобное для вас место. Дефолтный путь - `/opt/traefik`.

**Настройка Traefik:**

Переместите папку traefik в /opt и зайдите в данную папку. Все дальнейшие пути, файлы и команды подразумеваются относительно данной директории.

1) Создайте нового пользователя с именем *admin* для последующего доступа к Traefik Dashboard:
```
htpasswd -nbB admin <password>
```

Пароль хэшируется с помощью bcrypt алгоритма (-nbB).

2) Установите вывод данной команды в файле *traefik.toml*:
```toml
[entryPoints]
  [entryPoints.dashboard]
    address = ":8081"
    [entryPoints.dashboard.auth]
      [entryPoints.dashboard.auth.basic]
        users = ["admin:hash"]
```

3) Установите имя вашего домена:
```toml
[docker]
domain = "<domain>"
```

Например: `example.com`.

4) Установите вашу почту:
```toml
[acme]
email = "<email>"
```

Например: `email@example.com`.

5) Откройте файл *docker-compose.yml* и установите ваш домен в пути для доступа к Traefik Dashboard:
```yaml
labels: 
  - "traefik.frontend.rule=Host:monitor.<domain>"
```

Например: `"traefik.frontend.rule=Host:monitor.example.com"`

7) Создайте Docker сеть для Traefik. Все контейнеры, которые заиспользуют данную сеть, будут доступны из интернета по тому пути, который 
вы укажете в frontend rule.
```
docker network create web
```

8) Запустите Traefik:
```
docker-compose up -d
```

**Запуск сервера:**

Зайдите в папку с сервером. Все дальнейшие пути, файлы и команды подразумеваются относительно данной директории.

Для запуска сервера с поддержкой Traefik используется другой docker-compose файл - *docker-compose.yml*. В нем должны быть установлены те же самые переменные окружения, как и в случае запуска сервера без Traefik.

1) Откройте *docker-compose.yml* и установите ваш домен в следующих docker label'ах:
```yaml
labels:
  - "traefik.site.frontend.rule=Host:<domain>"
  - "traefik.adminRobot.frontend.rule=Host:admin.<domain>;Path:/robots.txt"
  - "traefik.admin.frontend.rule=Host:admin.<domain>"
  - "traefik.admin.frontend.auth.basic.users=user:hash"
```

Объяснение:
* *traefik.site.frontend.rule=Host:<domain>* - основной путь для доступа к сайту
* *traefik.adminRobot.frontend.rule=Host:admin.<domain>;Path:/robots.txt* - специальное правило для поисковых роботов: мы не защищаем этот путь, чтобы робот смог получить robots.txt для админки
* *traefik.admin.frontend.rule=Host:admin.<domain>* - путь для доступа к админке
* *traefik.admin.frontend.auth.basic.users=user:hash* - Basic auth для админки. Используйте или уже раннее сгенерированные данные, или сгенерируйте новые по тому же принципу. **Важно:** знаки '$' должны быть продублированы для эскейпинга. Например, если хэш - `$2y$05$qZmUyodGOgHfc7momnSAcehoXD8ue8ZXI0wFBdLpCb9faNsgWSk2W`, то эскейпированная строка: `$$2y$$05$$qZmUyodGOgHfc7momnSAcehoXD8ue8ZXI0wFBdLpCb9faNsgWSk2W`.

2) Убедитесь, что в label *"traefik.port="* стоит тот же самый порт, что и в env переменной *SERVER_PORT*.

3) Запустите сервер:
```
docker-compose up -d
```

Теперь вам доступны следующие возможности:
* Мониторинг контейнеров из Traefik Dashboard, который находится по пути *monitor.\<domain>*
* Основной сайт доступен по адресу *\<domain>*
* Админка доступна по адресу *admin.\<domain>*