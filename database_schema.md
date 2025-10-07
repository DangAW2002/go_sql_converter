# Database Schema Description

This document describes the schema of two main tables in the SOVIGAZ database: `rdas_dev` and `sensor_data`. These tables are used for storing device information and sensor telemetry data.

## Table: rdas_dev

The `rdas_dev` table stores real-time device status and configuration information. It has 133 columns.

| Field                  | Type          | Null | Key | Default                                      | Extra |
|------------------------|---------------|------|-----|----------------------------------------------|-------|
| userID                 | int(4)        | NO   |     | NULL                                         |       |
| devID                  | varchar(16)   | NO   | MUL | NULL                                         |       |
| SerialNumber           | varchar(4)    | NO   |     | 0000                                         |       |
| Name                   | varchar(128)  | NO   |     | NULL                                         |       |
| Name1                  | varchar(32)   | NO   |     | NULL                                         |       |
| single_dual            | varchar(2)    | NO   |     | 1                                            |       |
| Tank                   | int(4)        | YES  |     | NULL                                         |       |
| SubTank                | int(4)        | YES  |     | NULL                                         |       |
| SubTank1               | int(4)        | YES  |     | NULL                                         |       |
| Location               | varchar(100)  | NO   |     | 468A, Nguyễn Văn Công, Gò Vấp, TP HCM        |       |
| Description            | varchar(16)   | NO   |     | Unknown                                      |       |
| Type                   | varchar(11)   | NO   |     | G103WM                                       |       |
| Project                | varchar(50)   | NO   | MUL | HCLOUD                                       |       |
| status                 | datetime      | YES  |     | NULL                                         |       |
| LatestData             | datetime      | YES  |     | NULL                                         |       |
| Current_ss1            | varchar(12)   | YES  |     | NULL                                         |       |
| Current_ss2            | varchar(12)   | YES  |     | NULL                                         |       |
| Current_ss3            | varchar(12)   | YES  |     | NULL                                         |       |
| Current_ss4            | varchar(12)   | YES  |     | NULL                                         |       |
| Current_ss5            | varchar(12)   | YES  |     | NULL                                         |       |
| Current_ss6            | varchar(12)   | YES  |     | NULL                                         |       |
| Current_ss7            | varchar(12)   | YES  |     | NULL                                         |       |
| Current_ss8            | varchar(12)   | YES  |     | NULL                                         |       |
| Totalizer1             | double        | NO   |     | 0                                            |       |
| Totalizer2             | double        | NO   |     | 0                                            |       |
| alarm_ss1_H            | varchar(8)    | NO   |     | 16000                                        |       |
| alarm_ss2_H            | varchar(8)    | NO   |     | 16000                                        |       |
| alarm_ss3_H            | varchar(8)    | NO   |     | 16000                                        |       |
| alarm_ss4_H            | varchar(8)    | NO   |     | 16000                                        |       |
| alarm_ss5_H            | varchar(8)    | NO   |     | 16000                                        |       |
| alarm_ss6_H            | varchar(8)    | NO   |     | 16000                                        |       |
| alarm_ss1_L            | varchar(8)    | NO   |     | 4000                                         |       |
| alarm_ss2_L            | varchar(8)    | NO   |     | 4000                                         |       |
| alarm_ss3_L            | varchar(8)    | NO   |     | 4000                                         |       |
| alarm_ss4_L            | varchar(8)    | NO   |     | 4000                                         |       |
| alarm_ss5_L            | varchar(8)    | NO   |     | 4000                                         |       |
| alarm_ss6_L            | varchar(8)    | NO   |     | 4000                                         |       |
| Lat                    | varchar(20)   | NO   |     | 106.68929859995                              |       |
| Lon                    | varchar(20)   | NO   |     | 10.798705000731                              |       |
| Connection             | varchar(16)   | NO   |     | DOWN                                         |       |
| Warning                | varchar(256)  | NO   |     | Non                                          |       |
| RstCnt                 | int(9)        | NO   |     | 0                                            |       |
| SendFailed             | int(9)        | NO   |     | 0                                            |       |
| TotalFrm               | int(9)        | NO   |     | 0                                            |       |
| GSMRst                 | int(9)        | NO   |     | 0                                            |       |
| FWRev                  | varchar(8)    | NO   |     | Unknown                                      |       |
| QoS                    | int(1)        | NO   |     | 0                                            |       |
| 4mA_ss1                | varchar(8)    | NO   |     | 4000                                         |       |
| 20mA_ss1               | varchar(8)    | NO   |     | 20000                                        |       |
| 4mA_ss2                | varchar(8)    | NO   |     | 4000                                         |       |
| 20mA_ss2               | varchar(8)    | NO   |     | 20000                                        |       |
| 4mA_ss3                | varchar(8)    | NO   |     | 4000                                         |       |
| 20mA_ss3               | varchar(8)    | NO   |     | 20000                                        |       |
| 4mA_ss4                | varchar(8)    | NO   |     | 4000                                         |       |
| 20mA_ss4               | varchar(8)    | NO   |     | 20000                                        |       |
| 4mA_ss5                | varchar(8)    | NO   |     | 4000                                         |       |
| 20mA_ss5               | varchar(8)    | NO   |     | 20000                                        |       |
| 4mA_ss6                | varchar(8)    | NO   |     | 4000                                         |       |
| 20mA_ss6               | varchar(8)    | NO   |     | 20000                                        |       |
| 4mA_ss7                | varchar(8)    | NO   |     | 4000                                         |       |
| StartLogging           | datetime      | YES  |     | NULL                                         |       |
| EnableLogging          | varchar(8)    | NO   |     | DISABLE                                      |       |
| EnableLogging2         | varchar(8)    | YES  |     | DISABLE                                      |       |
| 4mA_ss8                | varchar(8)    | NO   |     | 4000                                         |       |
| CMDCODE1               | varchar(128)  | YES  |     | NOCMD                                        |       |
| Reserve                | varchar(16)   | YES  |     | NULL                                         |       |
| formula1               | varchar(1024) | YES  |     | NULL                                         |       |
| formula2               | varchar(1024) | YES  |     | NULL                                         |       |
| klr                    | text          | YES  |     | NULL                                         |       |
| mmO2                   | text          | YES  |     | NULL                                         |       |
| Nm3                    | text          | YES  |     | NULL                                         |       |
| khoiluong              | text          | YES  |     | NULL                                         |       |
| muclong                | text          | YES  |     | NULL                                         |       |
| klr_2                  | text          | YES  |     | NULL                                         |       |
| mmO2_2                 | text          | YES  |     | NULL                                         |       |
| Nm3_2                  | text          | YES  |     | NULL                                         |       |
| khoiluong_2            | text          | YES  |     | NULL                                         |       |
| muclong_2              | text          | YES  |     | NULL                                         |       |
| sample_time            | int(4)        | NO   |     | 1                                            |       |
| SendingRate            | int(2)        | NO   |     | 1                                            |       |
| LastReport             | datetime      | YES  |     | NULL                                         |       |
| AI1_name               | varchar(16)   | YES  |     | Sensor1                                      |       |
| AI2_name               | varchar(16)   | NO   |     | Sensor2                                      |       |
| AI3_name               | char(16)      | NO   |     | Sensor3                                      |       |
| AI4_name               | char(16)      | NO   |     | Sensor4                                      |       |
| AI5_name               | char(16)      | NO   |     | Sensor5                                      |       |
| AI6_name               | char(16)      | NO   |     | Sensor6                                      |       |
| AI7_name               | varchar(24)   | YES  |     | NULL                                         |       |
| AI8_name               | varchar(24)   | YES  |     | NULL                                         |       |
| AI1_Unit               | varchar(8)    | YES  |     | NULL                                         |       |
| AI2_Unit               | varchar(8)    | YES  |     | NULL                                         |       |
| AI3_Unit               | varchar(8)    | YES  |     | NULL                                         |       |
| AI4_Unit               | varchar(8)    | YES  |     | NULL                                         |       |
| AI5_Unit               | varchar(8)    | YES  |     | NULL                                         |       |
| AI6_Unit               | varchar(8)    | YES  |     | NULL                                         |       |
| AI7_Unit               | varchar(8)    | YES  |     | NULL                                         |       |
| AI8_Unit               | varchar(8)    | YES  |     | NULL                                         |       |
| TempInside             | float         | NO   |     | 0                                            |       |
| TempOutside            | float         | NO   |     | 0                                            |       |
| MainPower              | float         | NO   |     | 0                                            |       |
| Battery                | float         | NO   |     | 0                                            |       |
| GSMSignal              | int(4)        | NO   |     | 99                                           |       |
| CID_LAC                | varchar(50)   | NO   |     | 0/0                                          |       |
| MNC                    | varchar(50)   | NO   |     | UNKNOWN                                      |       |
| DeviceEventCode        | varchar(4)    | NO   |     | NEC                                          |       |
| AlarmEnable            | varchar(10)   | NO   |     | ENABLE                                       |       |
| AlarmEnable2           | varchar(10)   | YES  |     | DISABLE                                      |       |
| DailyLatched           | varchar(16)   | NO   |     | D_M_000                                      |       |
| MonthlyLatched         | varchar(16)   | NO   |     | D_M_000                                      |       |
| MonthlyTotal           | bigint(20)    | YES  |     | NULL                                         |       |
| AlertEmailList2        | varchar(256)  | YES  |     | NULL                                         |       |
| AlertEmailList         | varchar(256)  | YES  |     | NULL                                         |       |
| Sent_flag              | varchar(16)   | YES  |     | 0                                            |       |
| Sent_flag2             | varchar(16)   | YES  |     | 0                                            |       |
| OnlineStatus           | varchar(3)    | NO   |     | ONL                                          |       |
| BatteryStatus          | varchar(3)    | NO   |     | OK                                           |       |
| HWRev                  | varchar(8)    | YES  |     | NULL                                         |       |
| NSX                    | date          | YES  |     | NULL                                         |       |
| DataServer             | varchar(32)   | YES  |     | NULL                                         |       |
| BatteryUsage           | float         | YES  |     | NULL                                         |       |
| BatteryUsageDate       | date          | YES  |     | NULL                                         |       |
| BatteryInstallDate     | date          | YES  |     | NULL                                         |       |
| alarm_muclong_L        | float         | YES  |     | -1                                           |       |
| alarm_muclong_H        | float         | YES  |     | 1000000000                                   |       |
| alarm_muclong_L2       | float         | YES  |     | -1                                           |       |
| alarm_muclong_H2       | float         | YES  |     | 1000000000                                   |       |
| UnBox                  | varchar(1)    | YES  |     | NULL                                         |       |
| offline_count          | int(11)       | YES  |     | 0                                            |       |
| total_downtime_minutes | int(11)       | YES  |     | 0                                            |       |
| last_offline_start     | datetime      | YES  |     | NULL                                         |       |
| last_offline_end       | datetime      | YES  |     | NULL                                         |       |
| uptime_percentage      | decimal(5,2)  | YES  |     | 100.00                                       |       |
| last_stats_update      | datetime      | YES  |     | NULL                                         |       |

## Table: sensor_data

The `sensor_data` table stores historical sensor readings and telemetry data. It has 22 columns, with `Idx` as the primary key (auto-increment).

| Field             | Type        | Null | Key | Default | Extra          |
|-------------------|-------------|------|-----|---------|----------------|
| Idx               | bigint(64)  | NO   | PRI | NULL    | auto_increment |
| deviceID          | varchar(6)  | NO   | MUL | NULL    |                |
| HWSerial          | varchar(8)  | YES  |     | NULL    |                |
| status            | varchar(48) | NO   |     | NULL    |                |
| sensor1           | double      | YES  |     | NULL    |                |
| sensor2           | double      | YES  |     | NULL    |                |
| sensor3           | double      | YES  |     | NULL    |                |
| sensor4           | double      | YES  |     | NULL    |                |
| sensor5           | varchar(12) | NO   |     | NULL    |                |
| sensor6           | varchar(12) | NO   |     | NULL    |                |
| sensor7           | varchar(12) | NO   |     | NULL    |                |
| sensor8           | varchar(12) | NO   |     | NULL    |                |
| TempInside        | varchar(4)  | YES  |     |         |                |
| TempOutside       | varchar(4)  | YES  |     | 0       |                |
| MainPower         | float       | YES  |     | 0       |                |
| Battery           | float       | YES  |     | 0       |                |
| SensorPowerStatus | varchar(4)  | NO   |     | OF      |                |
| GSMSignal         | int(4)      | NO   |     | 99      |                |
| Current_timestamp | datetime    | NO   |     | NULL    |                |
| date_time         | datetime    | NO   |     | NULL    |                |
| UnBox             | varchar(2)  | NO   |     | NULL    |                |
| Alert             | varchar(32) | YES  |     | NULL    |                |