# Скачать исходные данные от организаций по формам статистического наблюдения оо1 и оо2 https://cabinet.miccedu.ru

## Как пользоваться
```shell
cd miccedu-downloader
make
./bin/miccedu-downloader_os_arch -r oo1 -y 2021 -l login -p password
```

## Флаги
```shell
-r: название отчёта, моджет быть только <oo1> или <oo2>
-y: год отчёта, может быть только тем годом, который доступен в личном кабинете
-l: логин от личного кабинета
-p: пароль от личного кабинета
```

## Узнать доступные для скачивания годы
```shell
./bin/miccedu-downloader_os_arch -l login -p password -ro true
```