# Go SQL Converter & MQTT Subscriber

## Mô tả
Chương trình này nhận dữ liệu từ MQTT, lưu lịch sử vào bảng `sensor_data` và cập nhật trạng thái realtime vào bảng `rdas_dev` trong MariaDB/MySQL.

## Cấu trúc thư mục
```
go_sql_converter/
├── config/
│   └── config.yaml         # Cấu hình MQTT & SQL
├── scripts/
│   ├── build_and_run.sh    # Build & chạy foreground
│   ├── run_background.sh   # Chạy nền, quản lý PID/log
│   ├── service_manager.sh  # Quản lý systemd service
│   └── mqtt-converter.service # File cấu hình systemd
├── src/
│   └── main.go             # Source code chính
├── go.mod, go.sum          # Go modules
└── main                    # File thực thi sau khi build
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
Sửa file `config/config.yaml` để thay đổi thông tin MQTT và database.

## Kiểm tra dữ liệu
- Kiểm tra lịch sử sensor:
  ```sql
  SELECT * FROM sensor_data WHERE deviceID = 'L98000' ORDER BY Idx DESC LIMIT 1;
  ```
- Kiểm tra trạng thái realtime:
  ```sql
  SELECT * FROM rdas_dev WHERE devID = 'L98000';
  ```

## Yêu cầu
- Go >= 1.18
- MariaDB/MySQL
- MQTT broker
- Ubuntu (khuyến nghị)

