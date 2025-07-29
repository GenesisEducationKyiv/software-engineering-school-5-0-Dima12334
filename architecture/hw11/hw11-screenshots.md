## Впровадження метрик

Для моніторингу стану та коректної роботи сервісу були впроваджені базові метрики за допомогою Prometheus. Моніторинг здійснюється за допомогою Grafana.

### Інтеграція з додатком

У HTTP-сервер мікросервісу ms-weather-subscription, реалізований з використанням Gin, було додано роутер для експорту метрик:
```
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

Цей endpoint повертає стандартні метрики Go та HTTP за допомогою бібліотеки `promhttp`.

--- 

### Список цілей Prometheus

![prometheus\_targets](screenshots/prometheus_targets.png)

--- 

### Загальна інформація про запуск Prometheus

![prometheus\_run\_info](screenshots/prometheus_run_info.png)

--- 

### Перегляд експорту метрик з /metrics

![prometheus\_metrics](screenshots/prometheus_metrics.png)

--- 

### Дашборд у Grafana

![grafana\_charts](screenshots/grafana_charts.png)
