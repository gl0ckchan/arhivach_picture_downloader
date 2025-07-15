# Arhivach Picture Downloader
Загружает фото и видео с сайта arhivach.

# Сборка
```console
$ git clone https://github.com/gl0ckchan/arhivach_picture_downloader
$ cd arhivach_picture_downloader
$ go build
$ ./arhivach_picture_downloader -link <LINK_TO_THREAD> -goroutines 5
```

# Использование
```console
-goroutines int
    	количество запускаемых горутин (по умолчанию 3)
  -link string
    	ссылка на тред
```

# TODO
- [ ] Добавить поддержку прокси/TOR
