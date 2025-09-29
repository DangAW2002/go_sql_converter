# Go MQTT to SQL Converter

## Mô tả
Chương trình Go này subscribe MQTT topic `v1/DLOG4G/#`, xử lý dữ liệu telemetry (sensor readings + UnBox status) và attributes từ thiết bị IoT, sau đó lưu lịch sử vào bảng `sensor_data` và cập nhật trạng thái realtime vào bảng `rdas_dev` trong MariaDB/MySQL.

### Tính năng chính
- **Telemetry Processing**: Xử lý dữ liệu sensor (pressure1, level1, pressure2, level2) và UnBox status. Nếu chỉ gửi UnBox, chỉ cập nhật UnBox. Payload là JSON hợp lệ (ví dụ: `{"ts":1759115543000,"values":{"pressure1":0,"level1":0,"pressure2":0,"level2":0}}` hoặc `{"ts":1759090704000,"values":{"UnBox":"C"}}`).
- **Attributes Processing**: Cập nhật các thuộc tính như MainPower, GSMSignal, sample_time, SendingRate, UnBox.
- **Database Operations**: Insert lịch sử sensor, update trạng thái realtime.
- **Logging**: Ghi log HTML chi tiết vào `/opt/lampp/htdocs/DAQ/LOG/`.
- **Reliability**: Retry connection database/MQTT nếu thất bại.

## Cấu trúc thư mục
```
go_sql_converter/
├── cmd/
│   └── main.go             # Entry point chính
├── config/
│   └── config.yaml         # Cấu hình MQTT & SQL
├── internal/
│   ├── database.go         # Xử lý database & MQTT messages
│   ├── logger.go           # Ghi log HTML
│   └── mqtt.go             # Setup MQTT connection & subscription
├── scripts/
│   ├── build_and_run.sh    # Build & chạy foreground
│   ├── service_manager.sh  # Quản lý systemd service
│   └── mqtt-converter.service # File cấu hình systemd
├── old/
│   └── main.go             # Phiên bản cũ
├── build/                  # Thư mục build output
├── go.mod, go.sum          # Go modules
└── README.md
```

## Hướng dẫn sử dụng

### 1. Build & chạy foreground
```bash
./scripts/build_and_run.sh
```

### 2. Chạy nền (background)
```bash
chmod +x scripts/run_background.sh
./scripts/run_background.sh start
```
- Xem log: `./scripts/run_background.sh logs`
- Dừng: `./scripts/run_background.sh stop`

### 3. Tự động chạy khi khởi động (systemd)
```bash
chmod +x scripts/service_manager.sh
./scripts/service_manager.sh install
```
- Khởi động: `./scripts/service_manager.sh start`
- Dừng: `./scripts/service_manager.sh stop`
- Trạng thái: `./scripts/service_manager.sh status`
- Xem log: `./scripts/service_manager.sh logs`
- Gỡ service: `./scripts/service_manager.sh uninstall`

## Cấu hình
Sửa file `config/config.yaml` để thay đổi thông tin MQTT và database:
```yaml
mqtt:
  host: "localhost"
  port: 1883
  user: "weblog"
  pass: "Stm32f103rd"
  protocol: "tcp"
  topic: "v1/DLOG4G/#"

sql:
  host: "localhost"
  port: 3306
  user: "weblog"
  pass: "Weblog08052020"
  dbname: "SOVIGAZ"
```

## Kiểm tra dữ liệu
- Kiểm tra lịch sử sensor:
  ```sql
  SELECT * FROM sensor_data WHERE deviceID = 'L98000' ORDER BY Idx DESC LIMIT 10;
  ```
- Kiểm tra trạng thái realtime:
  ```sql
  SELECT devID, status, LatestData, Current_ss1, Current_ss2, Current_ss3, Current_ss4, UnBox FROM rdas_dev WHERE devID = 'L98000';
  ```
- Xem log HTML: `/opt/lampp/htdocs/DAQ/LOG/LOG_DD_MM_YYYY-DEVICEID.html`

## Yêu cầu hệ thống
- Go >= 1.18
- MariaDB/MySQL
- MQTT broker (e.g., Mosquitto)
- Ubuntu/Debian (khuyến nghị cho systemd)

## Dependencies
- `github.com/eclipse/paho.mqtt.golang` (MQTT client)
- `github.com/go-sql-driver/mysql` (MySQL driver)
- `gopkg.in/yaml.v3` (YAML config parser)

