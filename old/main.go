package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
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

func main() {
	log.Println("Starting MQTT Subscriber...")

	// Read config
	log.Println("Loading configuration from config/config.yaml...")
	configFile, err := os.Open("../config/config.yaml")
	if err != nil {
		log.Printf("Error opening config file: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)

	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		log.Printf("Error decoding config: %v. Retrying in 5 seconds...", err)
		configFile.Close()
		time.Sleep(5 * time.Second)

	}
	log.Println("Config loaded successfully.")

	// Database setup
	log.Println("Connecting to database...")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.SQL.User, config.SQL.Pass, config.SQL.Host, config.SQL.Port, config.SQL.Dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Connected to database successfully.")
	defer db.Close()

	// MQTT setup
	log.Println("Configuring MQTT client...")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s:%d", config.MQTT.Protocol, config.MQTT.Host, config.MQTT.Port))
	opts.SetUsername(config.MQTT.User)
	opts.SetPassword(config.MQTT.Pass)
	opts.SetClientID("go-mqtt-subscriber")

	client := mqtt.NewClient(opts)
	log.Println("Connecting to MQTT broker...")
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("MQTT connection failed:", token.Error())
	}
	log.Println("Connected to MQTT broker successfully.")
	defer client.Disconnect(250)

	// Subscribe to topic
	log.Printf("Subscribing to topic: %s", config.MQTT.Topic)
	token := client.Subscribe(config.MQTT.Topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
		if strings.Contains(msg.Topic(), "telemetry") {
			processTelemetryMessage(db, msg)
		} else if strings.Contains(msg.Topic(), "attributes") {
			processAttributesMessage(db, msg)
		}
	})
	token.Wait()
	if token.Error() != nil {
		log.Fatal("Subscription failed:", token.Error())
	}
	log.Println("Subscription successful. Waiting for messages...")

	// Keep running
	for {
		time.Sleep(1 * time.Second)
	}
}

func processTelemetryMessage(db *sql.DB, msg mqtt.Message) {
	payload := msg.Payload()
	originalPayload := string(payload) // Keep original for logging

	// Fix the JSON by adding quotes around keys
	re := regexp.MustCompile(`(\w+):`)
	fixedPayload := re.ReplaceAllStringFunc(string(payload), func(match string) string {
		key := strings.TrimSuffix(match, ":")
		return "\"" + key + "\":"
	})
	log.Printf("Fixed payload: %s", fixedPayload)
	var data []map[string]interface{}
	err := json.Unmarshal([]byte(fixedPayload), &data)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	parts := strings.Split(msg.Topic(), "/")
	if len(parts) < 3 {
		log.Printf("Invalid topic format: %s", msg.Topic())
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
		chVal, hasCh := values["Channel"]
		l1Val, hasL1 := values["level1"]
		p2Val, hasP2 := values["pressure2"]
		l2Val, hasL2 := values["level2"]
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

		// Apply calibration and rounding compatible with WEBLOG
		sensor1 = math.Round(sensor1/1000*100) / 100
		sensor2 = math.Round(sensor2/1000*100) / 100
		sensor3 = math.Round(sensor3/1000*10000) / 10000
		sensor4 = math.Round(sensor4/1000*10000) / 10000

		dt := time.Unix(int64(ts)/1000, (int64(ts)%1000)*int64(time.Millisecond))
		dtStr := dt.Format("2006-01-02 15:04:05")
		log.Printf("Parsed values: sensor1=%.2f, sensor2=%.2f, sensor3=%.2f, sensor4=%.2f, timestamp=%s, deviceID=%s", sensor1, sensor2, sensor3, sensor4, dtStr, deviceID)

		var insertResult, updateResult string

		sqlQuery := fmt.Sprintf("INSERT INTO sensor_data (deviceID, status, sensor1, sensor2, sensor3, sensor4, sensor5, sensor6, sensor7, sensor8, SensorPowerStatus, GSMSignal, `Current_timestamp`, date_time, UnBox) VALUES ('%s', 'active', %f, %f, %f, %f, 0, 0, 0, 0, 'ON', 0, '%s', NOW(), 'NO')", deviceID, sensor1, sensor2, sensor3, sensor4, dtStr)
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
		updateQuery := fmt.Sprintf("UPDATE rdas_dev SET status = NOW(), LatestData = '%s', Current_ss1 = %f, Current_ss2 = %f, Current_ss3 = %f, Current_ss4 = %f, Type = 'MQTT' WHERE devID = '%s'", dtStr, sensor1, sensor2, sensor3, sensor4, deviceID)
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
		writeHTMLLog(msg.Topic(), originalPayload, deviceID, sensor1, sensor2, sensor3, sensor4, dtStr, insertResult, updateResult, 0, 0, 0, 0, "")
	}
}

func processAttributesMessage(db *sql.DB, msg mqtt.Message) {
	payload := msg.Payload()
	log.Printf("Processing attributes message: %s", string(payload))

	var attributes map[string]interface{}
	err := json.Unmarshal(payload, &attributes)
	if err != nil {
		log.Printf("Error parsing attributes JSON: %v", err)
		return
	}

	parts := strings.Split(msg.Topic(), "/")
	if len(parts) < 3 {
		log.Printf("Invalid topic format for attributes: %s", msg.Topic())
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
	writeHTMLLog(msg.Topic(), string(payload), deviceID, 0, 0, 0, 0, time.Now().Format("2006-01-02 15:04:05"), "N/A", updateResult, mainPower, gsmSignal, sampleTime, sendingRate, unBox)
}

func writeHTMLLog(topic, originalPayload, deviceID string, sensor1, sensor2, sensor3, sensor4 float64, timestamp, insertResult, updateResult string, mainPower float64, gsmSignal int, sampleTime int, sendingRate int, unBox string) {
	// Create log directory
	logDir := "/opt/lampp/htdocs/DAQ/LOG"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Error creating log directory: %v", err)
		return
	}

	// Generate filename: LOG_12_09_2025-T24000.html
	now := time.Now()
	dateStr := now.Format("02_01_2006") // DD_MM_YYYY
	filename := fmt.Sprintf("LOG_%s-%s.html", dateStr, deviceID)
	filePath := filepath.Join(logDir, filename)

	// Check if file exists to determine if we need to write header
	fileExists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileExists = false
	}

	// Open file for append
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()

	// Write HTML header if new file
	if !fileExists {
		header := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Device Log - %s</title>
    <style>
        table { border-collapse: collapse; width: 100%%; font-family: Arial, sans-serif; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; vertical-align: top; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .timestamp { font-weight: bold; color: #333; }
        .success { color: green; }
        .error { color: red; }
        .raw-data { max-width: 40%%; word-wrap: break-word; white-space: pre-wrap; }
        .processed-data { max-width: 30%%; }
        .db-ops { max-width: 30%%; }
    </style>
</head>
<body>
    <h1>Device Log - %s</h1>
    <h2>Date: %s</h2>
    <table>
        <tr>
            <th>Timestamp</th>
            <th>Device ID</th>
            <th>MQTT Topic</th>
            <th style="width: 40%%;">Raw Data</th>
            <th style="width: 30%%;">Processed Data</th>
            <th style="width: 30%%;">Database Operations</th>
        </tr>
`, deviceID, deviceID, now.Format("02/01/2006"))
		file.WriteString(header)
		log.Printf("Created new HTML log file for device %s", deviceID)
	}

	// Create data row
	currentTime := now.Format("2006-01-02 15:04:05")

	// Format numbers without unnecessary decimals
	formatNumber := func(val float64) string {
		if val == float64(int(val)) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%.3f", val)
	}

	processedData := fmt.Sprintf(`Timestamp: %s<br>
DeviceID: %s<br>
Type: MQTT<br>
Current_ss1: %s<br>
Current_ss2: %s<br>
Current_ss3: %s<br>
Current_ss4: %s<br>
SensorPowerStatus: ON<br>
GSMSignal: 0<br>
UnBox: NO`,
		timestamp, deviceID,
		formatNumber(sensor1), formatNumber(sensor2),
		formatNumber(sensor3), formatNumber(sensor4))

	if strings.Contains(topic, "attributes") {
		// For attributes, show the attributes data
		processedData = fmt.Sprintf(`Timestamp: %s<br>
DeviceID: %s<br>
Type: MQTT<br>
MainPower: %s<br>
GSMSignal: %d<br>
sample_time: %d<br>
SendingRate: %d<br>
UnBox: %s`,
			timestamp, deviceID,
			formatNumber(mainPower), gsmSignal, sampleTime, sendingRate, unBox)
	}

	// Determine CSS class based on operation results
	insertClass := "success"
	updateClass := "success"
	if strings.Contains(insertResult, "FAILED") {
		insertClass = "error"
	}
	if strings.Contains(updateResult, "FAILED") {
		updateClass = "error"
	}

	databaseOps := fmt.Sprintf(`<span class="%s">%s</span><br><span class="%s">%s</span>`,
		insertClass, insertResult, updateClass, updateResult)

	row := fmt.Sprintf(`        <tr>
            <td class="timestamp">%s</td>
            <td>%s</td>
            <td>%s</td>
            <td class="raw-data">%s</td>
            <td class="processed-data">%s</td>
            <td class="db-ops">%s</td>
        </tr>
`, currentTime, deviceID, topic, originalPayload, processedData, databaseOps)

	file.WriteString(row)
	log.Printf("Written HTML log for device %s", deviceID)
}
