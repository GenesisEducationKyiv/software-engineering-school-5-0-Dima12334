# System Design: Weather Subscription API

## 1. Вимоги системи

### Функціональні вимоги
- Користувачі мають змогу створити підписку на оновлення погоди для конкретного міста, вказавши свій email
- Попередньо потрібно підтрвердити підписку через посилання, яке надійде на вказаний email
- Система надсилає регулярні повідомлення на поштову скриньку (щодня, щогодини)
- API для керування підписками (створення і видалення підписки)
- Можливість отримання поточної інформації про погоду в обраному місті

### Нефункціональні вимоги
- **Доступність**: 95.5% uptime
- **Масштабованість**: до 80K користувачів, 800К повідомлень/день
- **Затримка**: < 300ms для API запитів
- **Надійність**: гарантована доставка поштових повідомлень
- **Безпека**: підтвердження створення підписки і валідація даних

### Обмеження
- **Budget**: мінімальна кількість інфраструктурних компонентів
- **External API rate limits**: 1М запитів/місяць
- **Compliance**: система відповідає вимогам Загального регламенту про захист даних (GDPR)

---

## 2. Оцінка навантаження

### Користувачі та трафік
- **Активні користувачі**: 50K
- **Підписки користувача**: 1–2
- **API запити**: 800 RPS (пік)
- **Повідомлення**: 90K/день

### Дані
- **Підписка**: ~350 bytes
- **Логи**: ~1KB (на один запит)
- **Загальний обсяг**: ~33GB/рік

### Bandwidth
- **Incoming**: 1 Mbps
- **Outgoing**: 3 Mbps (sending email)
- **External API**: 50 Mbps

---

## 3. High-Level архітектура
![img.png](high_level_architecture.png)

---

## 4. Детальний дизайн компонентів

### 4.1. Load Balancer (Nginx)
**Відповідальність**: 
- Розподіл навантаження між кількома екземплярами API сервісу
- Rate limiting для захисту від DDoS атак
- Health-check на `/ping`

### 4.2. API Gateway (Nginx)
**Відповідальність**:
- Маршрутизація запитів до API сервісу
- Rate limiting (1000 запитів/год, 5 запитів/сек на користувача)
- Обробка HTTPS

### 4.3. API Service (Go + Gin)
**Відповідальність**:
- Обробка HTTP-запитів (Gin framework)
- Робота з підписками (зберігання, підтвердження, видалення)
- Взаємодія з WeatherAPI та PostgreSQL

**Endpoints**:
- `GET    /ping                     `
- `GET    /api/weather/?city={city} `
- `POST   /api/subscribe            `
- `GET    /api/confirm/:token       `
- `GET    /api/unsubscribe/:token   `

### 4.4. Weather API (WeatherAPI integration)
**Відповідальність**:
- Запити до зовнішнього API (WeatherAPI)
- Перетворення даних у внутрішній формат

**Client settings**:
```go
const weatherAPIClientTimeout    = 10 * time.Second

type WeatherAPIClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewWeatherAPIClient(apiKey string) *WeatherAPIClient {
	return &WeatherAPIClient{
		APIKey:     apiKey,
		BaseURL:    "https://api.weatherapi.com/v1",
		HTTPClient: &http.Client{Timeout: weatherAPIClientTimeout},
	}
}
```

### 4.5. Scheduler (Cron jobs)
**Відповідальність**:
- Періодичні задачі по розкладу

**Задачі**:
```go
type CronRunner struct {
	services *service.Services
	cron     *cron.Cron
}

func (c *CronRunner) Start() {
	c.registerTasks()
	c.cron.Start()
}

func (c *CronRunner) registerTasks() {
	// Top of each hour (7:00, 8:00, 9:00, etc.)
	c.addTask("0 * * * *", c.hourlyWeatherEmailTask, "hourly weather email sending")
	// Daily at 7AM
	c.addTask("0 7 * * *", c.dailyWeatherEmailTask, "daily weather email sending")
}
```

### 4.6. PostgreSQL
**Відповідальність**:
- Зберігання даних про підписки на розсилку погоди

**Схема бази даних**:
![img.png](../adr/db_schema.png)

**Оптимізація**:
- B-Tree індекси на полях `id`, `email`, `token`
- Використання `uuid_generate_v4()` для генерації унікальних ідентифікаторів
- PRIMARY KEY constraint на полі `id` для швидкого доступу до підписки
- UNIQUE constraint на полях `email` і `token`
- Використання `TIMESTAMPTZ` для зберігання дати з часовим поясом