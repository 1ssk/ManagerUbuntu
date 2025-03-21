# Диспетчер задач (ManagerUbuntu)

**Диспетчер задач (ManagerUbuntu)** — это веб-приложение, написанное на Go, которое в реальном времени отображает состояние системы Ubuntu. Программа показывает информацию о загрузке CPU, использовании оперативной памяти и жёсткого диска, сетевую активность, а также список запущенных процессов, отсортированных по использованию CPU. Кроме того, приложение позволяет завершать задачи через удобный веб-интерфейс.

## Особенности

- **Реальное время:** Данные обновляются каждые 3 секунды без перезагрузки страницы.
- **Системная информация:** Отображение загрузки CPU, состояния оперативной памяти и жёсткого диска (в ГБ) и сетевой активности (в МБ).
- **Список процессов:** Все запущенные процессы сортируются по убыванию использования CPU.
- **Управление процессами:** Возможность завершения задачи нажатием кнопки «Завершить задачу».
- **Футуристический дизайн:** Интерфейс в стиле SpaceX с использованием шрифта Orbitron и тёмной цветовой гаммой.

## Установка

1. **Клонируйте репозиторий:**

   ```bash
   git clone https://github.com/1ssk/ManagerUbuntu.git
   cd ManagerUbuntu
   ```

2. **Установите зависимости:**

   Выполните команду:
   ```bash
   go get github.com/shirou/gopsutil/v3
   ```

3. **Запустите приложение:**

   ```bash
   go run main.go
   ```

4. **Откройте браузер:**

   Перейдите по адресу: [http://localhost:7000](http://localhost:7000)

## Использование

- **Обновление данных:** Интерфейс автоматически обновляется каждые 3 секунды, показывая актуальную информацию о системе.
- **Завершение процесса:** Для завершения задачи нажмите кнопку «Завершить задачу» рядом с нужным процессом.

## Настройка

При необходимости вы можете изменить период обновления или доработать дизайн, отредактировав HTML-шаблон в файле `main.go`. Также можно расширить функционал, добавив, например, поддержку дополнительных метрик или интеграцию с другими утилитами.

## Лицензия

Этот проект распространяется под лицензией MIT. Подробности смотрите в файле [LICENSE](License).
