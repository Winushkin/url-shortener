from locust import HttpUser, task, between

class FinalRedirectTest(HttpUser):
    # Минимальная пауза для генерации максимального RPS
    wait_time = between(0.005, 0.02)
    
    # Дефолтный код на случай, если Redis отключен и /shorten упадет с ошибкой
    short_code = "test1234"

    def on_start(self):
        """
        Пробуем создать ссылку динамически. 
        Если сервис без Redis упадет — ничего страшного, сработает дефолтный код.
        """
        payload = {"url": "https://example.com"}
        
        with self.client.post("/shorten", json=payload, name="[Setup] Try Create Link", catch_response=True) as response:
            if response.status_code == 201:
                # Если возвращается просто строка или готовый URL
                self.short_code = response.text.split("/")[-1].strip()[:-2]
            else:
                # Подавляем ошибку во время прогрева, чтобы продолжить тест редиректа
                response.success()

    @task
    def test_redirect(self):
        # Отключаем allow_redirects=False, чтобы тестировать именно СВОЙ сервис,
        # а не тот сайт, на который он перенаправляет.
        with self.client.get(f"/{self.short_code}", allow_redirects=False, catch_response=True) as response:
            
            # Успешный редирект — это статус 301, 302, 307 или 308
            if response.status_code == 302:
                response.success()
            else:
                response.failure(f"Fail! Status: {response.status_code}. Response: {response.text[:50]}")
