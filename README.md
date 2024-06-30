# Script Exporter

### Программа запускает любые скрипты и выводит метрики для Prometheus (или VictoriaMetrics)
Для правильной работы парсера - скрипт должен иметь примерно такой вывод:
```bash
node_exporter_disk_size_lsblk{disk="sda"} 62914560000
```
Где `node_exporter_disk_size_lsblk` - это название метрики, `disk` и `sda` - это лейбл и его ключ, а `62914560000`- это значение \

После обновления `v2.0.0` можно создавать метрики с несколькими лейблами (P.S Без пробелов)
```bash
node_exporter_host_info{type="VM", task="Daria", description="ClusterOfK8s", creater="Boris" } 1
```
## Установка из исходников

1. Клонировать репозиторий:
```bash
git clone https://github.com/hexqueller/Script-Exporter.git
```
2. Перейти в директорию:
```bash
cd Script-Exporter
```

3. Конфигурация
```bash
./configs/default.yaml
```
```yaml
jobs:
  - name: disk script
    cron: "* * * * *"
    script: scripts/disk.sh
  - name: multi_label
    cron: "0 */2 * * *"
    script: scripts/multi_label.sh
```
В этом примере две джобы: disk script и multi_label. disk script запускается каждую минуту, а multi_label каждые 2 часа.

4. Способы запуска:
```bash
make build # Просто собрать
```
```bash
make run # Собрать и сразу запустить
```
```bash
make docker # Собрать и запустить в Docker
```

5. О параметрах запуска
```bash
./exporter -p 9105 -c path/to/config.yaml -d
```
Флаг `-p` указывает порт по которому будет доступны метрики. \
Флаг `-c` указывает путь к файлу конфигурации. \
Флаг `-d` включает дебаг.