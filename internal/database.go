package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	MQTT struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
		Protocol string `yaml:"protocol"`
		Topic    string `yaml:"topic"`
	} `yaml:"mqtt"`
	SQL struct {
		Host   string `yaml:"host"`
		Port   int    `yaml:"port"`
		User   string `yaml:"user"`
		Pass   string `yaml:"pass"`
		Dbname string `yaml:"dbname"`
	} `yaml:"sql"`
}

func SetupDatabase(config Config) *sql.DB {
	log.Println("Connecting to database...")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.SQL.User, config.SQL.Pass, config.SQL.Host, config.SQL.Port, config.SQL.Dbname)
	var db *sql.DB
	for {
		var err error
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("Database connection failed: %v. Retrying in 10 seconds...", err)
			time.Sleep(10 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			log.Printf("Database ping failed: %v. Retrying in 10 seconds...", err)
			db.Close()
			time.Sleep(10 * time.Second)
			continue
		}
		log.Println("Connected to database successfully.")
		// Tối ưu connection pool
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(30 * 24 * time.Hour) // Tăng lên 1 tháng để đồng bộ với MQTT
		break
	}
	return db
}

func ProcessTelemetryMessage(db *sql.DB, msg mqtt.Message) {
	payload := msg.Payload()
	originalPayload := string(payload) // Keep original for logging

	var data []map[string]interface{}
	var err error

	if strings.Contains(originalPayload, "Channel:") {
		// Old format: fix JSON by adding quotes around keys
		re := regexp.MustCompile(`(\w+):`)
		fixedPayload := re.ReplaceAllStringFunc(string(payload), func(match string) string {
			key := strings.TrimSuffix(match, ":")
			return "\"" + key + "\":"
		})
		log.Printf("Fixed old format payload: %s", fixedPayload)
		err = json.Unmarshal([]byte(fixedPayload), &data)
	} else {
		// New format: direct parse
		log.Printf("Payload: %s", originalPayload)
		err = json.Unmarshal(payload, &data)
	}

	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		WriteHTMLLog(msg.Topic(), originalPayload, "unknown", 0, 0, 0, 0, time.Now().Format("2006-01-02 15:04:05"), "PARSE ERROR", "N/A", 0, 0, 0, 0, "")
		return
	}

	parts := strings.Split(msg.Topic(), "/")
	if len(parts) < 3 {
		log.Printf("Invalid topic format: %s", msg.Topic())
		WriteHTMLLog(msg.Topic(), originalPayload, "unknown", 0, 0, 0, 0, time.Now().Format("2006-01-02 15:04:05"), "INVALID TOPIC", "N/A", 0, 0, 0, 0, "")
		return
	}
	deviceID := parts[2]

	for _, item := range data {
		tsVal, ok := item["ts"]
		if !ok {
			continue
		}
		ts, ok := tsVal.(float64)
		if !ok {
			continue
		}
		valuesVal, ok := item["values"]
		if !ok {
			continue
		}
		values, ok := valuesVal.(map[string]interface{})
		if !ok {
			continue
		}
		chVal, hasCh := values["pressure1"]
		if !hasCh {
			chVal, hasCh = values["Channel"] // Backward compatibility for old format
		}
		l1Val, hasL1 := values["level1"]
		p2Val, hasP2 := values["pressure2"]
		l2Val, hasL2 := values["level2"]
		unBoxVal, hasUnBox := values["UnBox"]

		// If only UnBox is present, process it
		if hasUnBox && !hasCh && !hasL1 && !hasP2 && !hasL2 {
			unBoxStr, ok := unBoxVal.(string)
			if !ok || len(unBoxStr) == 0 {
				continue
			}
			unBox := string(unBoxStr[0]) // Take first character

			dt := time.Unix(int64(ts)/1000, (int64(ts)%1000)*int64(time.Millisecond))
			dtStr := dt.Format("2006-01-02 15:04:05")
			log.Printf("Parsed UnBox: %s, timestamp=%s, deviceID=%s", unBox, dtStr, deviceID)

			var insertResult, updateResult string

			sqlQuery := fmt.Sprintf("INSERT INTO sensor_data (deviceID, status, sensor1, sensor2, sensor3, sensor4, sensor5, sensor6, sensor7, sensor8, SensorPowerStatus, GSMSignal, `Current_timestamp`, date_time, UnBox) VALUES ('%s', 'active', 0, 0, 0, 0, 0, 0, 0, 0, 'ON', 0, NOW(), '%s', '%s')", deviceID, dtStr, unBox)
			log.Printf("Executing SQL: %s", sqlQuery)
			_, err := db.Exec(sqlQuery)
			if err != nil {
				log.Printf("Error executing SQL for sensor_data: %v", err)
				insertResult = "INSERT sensor_data: FAILED - " + err.Error()
			} else {
				log.Printf("Inserted sensor data for device %s at %s", deviceID, dtStr)
				insertResult = "INSERT sensor_data: SUCCESS"
			}

			// Update rdas_dev table
			updateQuery := fmt.Sprintf("UPDATE rdas_dev SET status = NOW(), LatestData = '%s', Type = 'MQTT', UnBox = '%s' WHERE devID = '%s'", dtStr, unBox, deviceID)
			log.Printf("Executing UPDATE SQL: %s", updateQuery)
			_, err = db.Exec(updateQuery)
			if err != nil {
				log.Printf("Error executing UPDATE for rdas_dev: %v", err)
				updateResult = "UPDATE rdas_dev: FAILED - " + err.Error()
			} else {
				log.Printf("Updated rdas_dev for device %s", deviceID)
				updateResult = "UPDATE rdas_dev: SUCCESS"
			}

			// Log to HTML file
			WriteHTMLLog(msg.Topic(), originalPayload, deviceID, 0, 0, 0, 0, dtStr, insertResult, updateResult, 0, 0, 0, 0, unBox)
			continue // Skip to next item
		}

		if !hasCh || !hasL1 || !hasP2 || !hasL2 {
			continue
		}
		sensor1, ok := chVal.(float64)
		if !ok {
			continue
		}
		sensor2, ok := l1Val.(float64)
		if !ok {
			continue
		}
		sensor3, ok := p2Val.(float64)
		if !ok {
			continue
		}
		sensor4, ok := l2Val.(float64)
		if !ok {
			continue
		}

		// Check for UnBox
		var unBox string = "C"
		if unBoxVal, hasUnBox := values["UnBox"]; hasUnBox {
			if unBoxStr, ok := unBoxVal.(string); ok && len(unBoxStr) > 0 {
				unBox = string(unBoxStr[0])
			}
		}

		// Apply calibration and rounding compatible with WEBLOG
		sensor1 = math.Round(sensor1/1000*100) / 100
		sensor2 = math.Round(sensor2/1000*100) / 100
		sensor3 = math.Round(sensor3/1000*10000) / 10000
		sensor4 = math.Round(sensor4/1000*10000) / 10000

		dt := time.Unix(int64(ts)/1000, (int64(ts)%1000)*int64(time.Millisecond))
		dtStr := dt.Format("2006-01-02 15:04:05")
		log.Printf("Parsed values: sensor1=%.2f, sensor2=%.2f, sensor3=%.2f, sensor4=%.2f, timestamp=%s, deviceID=%s", sensor1, sensor2, sensor3, sensor4, dtStr, deviceID)

		var insertResult, updateResult string

		sqlQuery := fmt.Sprintf("INSERT INTO sensor_data (deviceID, status, sensor1, sensor2, sensor3, sensor4, sensor5, sensor6, sensor7, sensor8, SensorPowerStatus, GSMSignal, `Current_timestamp`, date_time, UnBox) VALUES ('%s', 'active', %f, %f, %f, %f, 0, 0, 0, 0, 'ON', 0, NOW(), '%s', '%s')", deviceID, sensor1, sensor2, sensor3, sensor4, dtStr, unBox)
		log.Printf("Executing SQL: %s", sqlQuery)
		_, err := db.Exec(sqlQuery)
		if err != nil {
			log.Printf("Error executing SQL for sensor_data: %v", err)
			insertResult = "INSERT sensor_data: FAILED - " + err.Error()
		} else {
			log.Printf("Inserted sensor data for device %s at %s", deviceID, dtStr)
			insertResult = "INSERT sensor_data: SUCCESS"
		}

		// Update rdas_dev table
		var updates []string
		updates = append(updates, fmt.Sprintf("status = NOW(), LatestData = '%s', Current_ss1 = %f, Current_ss2 = %f, Current_ss3 = %f, Current_ss4 = %f, Type = 'MQTT'", dtStr, sensor1, sensor2, sensor3, sensor4))
		if hasUnBox {
			updates = append(updates, fmt.Sprintf("UnBox = '%s'", unBox))
		}
		updateQuery := fmt.Sprintf("UPDATE rdas_dev SET %s WHERE devID = '%s'", strings.Join(updates, ", "), deviceID)
		log.Printf("Executing UPDATE SQL: %s", updateQuery)
		_, err = db.Exec(updateQuery)
		if err != nil {
			log.Printf("Error executing UPDATE for rdas_dev: %v", err)
			updateResult = "UPDATE rdas_dev: FAILED - " + err.Error()
		} else {
			log.Printf("Updated rdas_dev for device %s", deviceID)
			updateResult = "UPDATE rdas_dev: SUCCESS"
		}

		// Log to HTML file
		WriteHTMLLog(msg.Topic(), originalPayload, deviceID, sensor1, sensor2, sensor3, sensor4, dtStr, insertResult, updateResult, 0, 0, 0, 0, unBox)
	}
}

func ProcessAttributesMessage(db *sql.DB, msg mqtt.Message) {
	payload := msg.Payload()
	log.Printf("Processing attributes message: %s", string(payload))

	var attributes map[string]interface{}
	err := json.Unmarshal(payload, &attributes)
	if err != nil {
		log.Printf("Error parsing attributes JSON: %v", err)
		WriteHTMLLog(msg.Topic(), string(payload), "unknown", 0, 0, 0, 0, time.Now().Format("2006-01-02 15:04:05"), "N/A", "PARSE ERROR", 0, 0, 0, 0, "")
		return
	}

	parts := strings.Split(msg.Topic(), "/")
	if len(parts) < 3 {
		log.Printf("Invalid topic format for attributes: %s", msg.Topic())
		WriteHTMLLog(msg.Topic(), string(payload), "unknown", 0, 0, 0, 0, time.Now().Format("2006-01-02 15:04:05"), "N/A", "INVALID TOPIC", 0, 0, 0, 0, "")
		return
	}
	deviceID := parts[2]

	// Initialize attribute values
	var mainPower float64
	var gsmSignal int
	var sampleTime int
	var sendingRate int
	var unBox string

	// Map attributes to database fields
	var updates []string

	if mainPowerStr, ok := attributes["main_power"].(string); ok {
		if mp, err := strconv.ParseFloat(mainPowerStr, 64); err == nil {
			mainPower = mp
			updates = append(updates, fmt.Sprintf("MainPower = %f", mainPower))
		} else {
			log.Printf("Error parsing main_power '%s' to float: %v", mainPowerStr, err)
		}
	}

	if gs, ok := attributes["GSM_Signal"].(float64); ok {
		gsmSignal = int(gs)
		updates = append(updates, fmt.Sprintf("GSMSignal = %d", gsmSignal))
	}

	if sr, ok := attributes["SamplingRate"].(float64); ok {
		sampleTime = int(sr)
		updates = append(updates, fmt.Sprintf("sample_time = %d", sampleTime))
	}

	if sdr, ok := attributes["SendingRate"].(float64); ok {
		sendingRate = int(sdr)
		updates = append(updates, fmt.Sprintf("SendingRate = %d", sendingRate))
	}

	if ub, ok := attributes["UnBox"].(string); ok && len(ub) > 0 {
		unBox = string(ub[0])
		updates = append(updates, fmt.Sprintf("UnBox = '%s'", unBox))
	}

	if len(updates) == 0 {
		log.Printf("No valid attributes to update for device %s", deviceID)
		WriteHTMLLog(msg.Topic(), string(payload), deviceID, 0, 0, 0, 0, time.Now().Format("2006-01-02 15:04:05"), "N/A", "NO VALID ATTRIBUTES", 0, 0, 0, 0, "")
		return
	}

	updates = append(updates, "Type = 'MQTT'")

	updateQuery := fmt.Sprintf("UPDATE rdas_dev SET %s WHERE devID = '%s'", strings.Join(updates, ", "), deviceID)
	log.Printf("Executing UPDATE SQL for attributes: %s", updateQuery)
	_, err = db.Exec(updateQuery)
	var updateResult string
	if err != nil {
		log.Printf("Error executing UPDATE for rdas_dev attributes: %v", err)
		updateResult = "UPDATE rdas_dev: FAILED - " + err.Error()
	} else {
		log.Printf("Updated rdas_dev attributes for device %s", deviceID)
		updateResult = "UPDATE rdas_dev: SUCCESS"
	}

	// Log to HTML file
	WriteHTMLLog(msg.Topic(), string(payload), deviceID, 0, 0, 0, 0, time.Now().Format("2006-01-02 15:04:05"), "N/A", updateResult, mainPower, gsmSignal, sampleTime, sendingRate, unBox)
}
