# Script Exporter

### Программа запускает любые скрипты и выводит метрики для Prometheus (или VictoriaMetrics)
Для правильной работы парсера - скрипт должен иметь примерно такой вывод:
```bash
node_exporter_disk_size_lsblk{disk="sda"} 62914560000
```
Где `node_exporter_disk_size_lsblk` - это название метрики, `disk` и `sda` - это лейбл и его ключ, а `62914560000`- это значение
## Установка из исходников

1. Клонировать репозиторий:
```bash
git clone https://github.com/hexqueller/Script-Exporter.git
```
2. Перейти в директорию:
```bash
cd Script-Exporter
```
3. Установить зависимости:
```bash
go get -d -v github.com/prometheus/client_golang/prometheus
go get -d -v github.com/prometheus/client_golang/prometheus/promhttp
go get -d -v github.com/robfig/cron/v3
go get -d -v gopkg.in/yaml.v2
```
4. Собрать:
```bash
go build cmd/exporter/main.go
```
4. Если надо собрать игнорируя glibc:
```bash
CGO_ENABLED=0 go build -o exporter cmd/exporter/main.go
```
5. Конфигурация
```yaml
jobs:
  - name: sh script
    cron: "* * * * *"
    script: /path/to/script1.sh
  - name: python script
    cron: "0 */2 * * *"
    script: /path/to/script2.py
```
В этом примере две джобы: sh script и python script. sh script запускается каждую минуту, а python script каждые 2 часа.

6. Запуск
```bash
./exporter -p 9105 -c path/to/config.yaml
```
Флаг `-p` указывает порт по которому будет доступны метрики
Флаг `-c` указывает путь к файлу конфигурации.