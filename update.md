# MQTT to SQL Converter - Update Documentation

## Tổng quan

Go MQTT Converter là ứng dụng Go nhận MQTT messages từ broker, parse JSON, và ghi vào MySQL database (SOVIGAZ). Hỗ trợ telemetry (sensor data) và attributes (device config).

### Cấu trúc DB
- **`sensor_data`**: Lịch sử sensor readings.
- **`rdas_dev`**: Trạng thái realtime device.
- **`alert`**: Cảnh báo (khi UnBox = "O").

### Logic xử lý

#### 1. Telemetry Message (`v1/DLOG4G/{deviceID}/telemetry`)
- **Parse deviceID** từ topic.
- **Parse JSON payload**.
- **Xử lý theo nội dung**:

##### Trường hợp 1: Chỉ UnBox
- **Payload ví dụ**: `[{"ts":1760294688000,"values":{"UnBox":"C"}}]` (hoặc `[{"ts":1760294688000,"values":{"UnBox":"O"}}]` nếu mở)
- **Xử lý**:
  - Update `rdas_dev`: `UPDATE rdas_dev SET status = NOW(), LatestData = '2025-10-13 12:00:00', UnBox = 'C', Type = 'MQTT' WHERE devID = 'L23007';`
  - Nếu `UnBox = "O"`: Insert `alert`: `INSERT INTO alert (EventTime, InsertTime, Source, Priority, Description, AlertType, Project, Type, Status, Note) VALUES ('2025-10-13 12:00:00', NOW(), 'L23007', 1, 'Device>Urgent:CANH BAO MO TU LUC 12:00:00 13_10_2025!!!!', 'Device', 'SOVIGAZ', 'Alert', 'V', '');`
- **Log HTML** (`LOG_DD_MM_YYYY-{deviceID}.html`):
  - **Processed Data** (UnBox = "C"): `Timestamp: 2025-10-13 12:00:00<br>DeviceID: L23007<br>Type: MQTT<br>UnBox: C<br>Alert: N/A`
  - **Processed Data** (UnBox = "O"): `Timestamp: 2025-10-13 12:00:00<br>DeviceID: L23007<br>Type: MQTT<br>UnBox: O<br>Alert: INSERT alert: SUCCESS`
  - **Database Operations**: `N/A<br>UPDATE rdas_dev: SUCCESS<br>INSERT alert: SUCCESS` (nếu UnBox = "O", else `N/A`)
- **Lệnh SQL kiểm tra**:
  ```sql
  -- Check rdas_dev update
  SELECT devID, status, LatestData, UnBox, Type FROM rdas_dev WHERE devID = 'L23007';
  -- Check alert (nếu UnBox = "O")
  SELECT * FROM alert WHERE Source = 'L23007' ORDER BY AlertID DESC LIMIT 1;
  -- Check HTML log
  ls -la /opt/lampp/htdocs/DAQ/LOG/LOG_*L23007*.html
  ```

##### Trường hợp 2: Đầy đủ sensor data
- **Payload ví dụ**: `[{"ts":1760294688000,"values":{"pressure1":4000,"level1":4000,"pressure2":4000,"level2":4000,"UnBox":"O"}}]`
- **Xử lý**:
  - Calibration: `sensor1 = Round(4000/1000 * 100) / 100 = 4.00`, tương tự sensor2-4.
  - Insert `sensor_data`: `INSERT INTO sensor_data (deviceID, status, sensor1, sensor2, sensor3, sensor4, sensor5, sensor6, sensor7, sensor8, SensorPowerStatus, GSMSignal, Current_timestamp, date_time, UnBox) VALUES ('L23007', 'active', 4.000000, 4.000000, 4.000000, 4.000000, 0, 0, 0, 0, 'ON', 0, NOW(), '2025-10-13 12:00:00', 'O');`
  - Update `rdas_dev`: `UPDATE rdas_dev SET status = NOW(), LatestData = '2025-10-13 12:00:00', Current_ss1 = 4.000000, Current_ss2 = 4.000000, Current_ss3 = 4.000000, Current_ss4 = 4.000000, UnBox = 'O', Type = 'MQTT' WHERE devID = 'L23007';`
  - Nếu `UnBox = "O"`: Insert `alert`: `INSERT INTO alert (EventTime, InsertTime, Source, Priority, Description, AlertType, Project, Type, Status, Note) VALUES ('2025-10-13 12:00:00', NOW(), 'L23007', 1, 'Device>Urgent:CANH BAO MO TU LUC 12:00:00 13_10_2025!!!!', 'Device', 'SOVIGAZ', 'Alert', 'V', '');`
- **Log HTML**:
  - **Processed Data**: `Timestamp: 2025-10-13 12:00:00<br>DeviceID: L23007<br>Type: MQTT<br>Current_ss1: 4.000<br>Current_ss2: 4.000<br>Current_ss3: 4.000<br>Current_ss4: 4.000<br>SensorPowerStatus: ON<br>GSMSignal: 0<br>UnBox: O<br>Alert: INSERT alert: SUCCESS`
  - **Database Operations**: `INSERT sensor_data: SUCCESS<br>UPDATE rdas_dev: SUCCESS<br>INSERT alert: SUCCESS`
- **Lệnh SQL kiểm tra**:
  ```sql
  -- Check sensor_data insert
  SELECT * FROM sensor_data WHERE deviceID = 'L23007' ORDER BY date_time DESC LIMIT 1;
  -- Check rdas_dev update
  SELECT devID, status, LatestData, Current_ss1, Current_ss2, Current_ss3, Current_ss4, UnBox, Type FROM rdas_dev WHERE devID = 'L23007';
  -- Check alert
  SELECT * FROM alert WHERE Source = 'L23007' ORDER BY AlertID DESC LIMIT 1;
  -- Check HTML log
  cat /opt/lampp/htdocs/DAQ/LOG/LOG_*L23007*.html | tail -20
  ```

##### Trường hợp 3: Parse error
- **Payload ví dụ**: `[{"ts":1760294688000,"values":{"UnBox":"C"]` (thiếu `}`)
- **Xử lý**: Log error, ghi HTML với "PARSE ERROR".
- **Log HTML**:
  - **Processed Data**: `PARSE ERROR`
  - **Database Operations**: `N/A`
- **Lệnh SQL kiểm tra**:
  ```sql
  -- No DB changes, check HTML log
  cat /opt/lampp/htdocs/DAQ/LOG/LOG_*L23007*.html | grep "PARSE ERROR"
  ```

#### 2. Attributes Message (`v1/DLOG4G/{deviceID}/attributes`)
- **Payload ví dụ**: `{"main_power":"12.5","GSM_Signal":25,"SamplingRate":1,"SendingRate":1,"UnBox":"C"}`
- **Xử lý**:
  - Update `rdas_dev`: `UPDATE rdas_dev SET MainPower = 12.5, GSMSignal = 25, sample_time = 1, SendingRate = 1, UnBox = 'C', Type = 'MQTT' WHERE devID = 'L23007';`
- **Log HTML**:
  - **Processed Data**: `Timestamp: 2025-10-13 12:00:00<br>DeviceID: L23007<br>Type: MQTT<br>MainPower: 12.500<br>GSMSignal: 25<br>sample_time: 1<br>SendingRate: 1<br>UnBox: C<br>Alert: N/A`
  - **Database Operations**: `N/A<br>UPDATE rdas_dev: SUCCESS<br>N/A`
- **Lệnh SQL kiểm tra**:
  ```sql
  -- Check rdas_dev update
  SELECT devID, MainPower, GSMSignal, sample_time, SendingRate, UnBox, Type FROM rdas_dev WHERE devID = 'L23007';
  -- Check HTML log
  cat /opt/lampp/htdocs/DAQ/LOG/LOG_*L23007*.html | tail -20
  ```

### Log HTML Features
- **File**: `/opt/lampp/htdocs/DAQ/LOG/LOG_DD_MM_YYYY-{deviceID}.html`
- **Raw Data**: Wrap mỗi 100 ký tự với `<br>`.
- **CSS**: Highlight success/error.

### Deploy
1. Build: `go build -o build/main ./cmd`
2. Install service: `./scripts/service_manager.sh install`
3. Start: `./scripts/service_manager.sh start`

### Lệnh SQL tổng quát
```sql
-- Connect to DB
mysql -u weblog -p -S /opt/lampp/var/mysql/mysql.sock SOVIGAZ

-- Check recent sensor_data
SELECT deviceID, sensor1, sensor2, sensor3, sensor4, date_time, UnBox FROM sensor_data ORDER BY date_time DESC LIMIT 10;

-- Check rdas_dev status
SELECT devID, status, LatestData, Current_ss1, Current_ss2, Current_ss3, Current_ss4, UnBox FROM rdas_dev WHERE devID LIKE 'L%';

-- Check alerts
SELECT * FROM alert ORDER BY AlertID DESC LIMIT 10;

-- Check logs
sudo journalctl -u mqtt-converter -f
```

### Notes
- Alert chỉ ghi khi UnBox = "O", bất kể có sensor hay không.
- Không ghi `sensor_data` nếu chỉ UnBox.
- Parse error ghi với deviceID từ topic.</content>
<parameter name="filePath">c:\Users\Dang\Desktop\sovigaz-dwh-ubuntu\go_sql_converter\update.md