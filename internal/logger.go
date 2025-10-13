package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func WriteHTMLLog(topic, originalPayload, deviceID string, sensor1, sensor2, sensor3, sensor4 float64, timestamp, insertResult, updateResult, alertResult string, mainPower float64, gsmSignal int, sampleTime int, sendingRate int, unBox string) {
	// Wrap raw data if longer than 100 chars
	wrappedPayload := wrapText(originalPayload, 100)
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
UnBox: %s`,
		timestamp, deviceID,
		formatNumber(sensor1), formatNumber(sensor2),
		formatNumber(sensor3), formatNumber(sensor4), unBox)

	if alertResult != "N/A" {
		processedData += fmt.Sprintf(`<br>Alert: %s`, alertResult)
	}

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
		if alertResult != "N/A" {
			processedData += fmt.Sprintf(`<br>Alert: %s`, alertResult)
		}
	}

	// Determine CSS class based on operation results
	insertClass := "success"
	updateClass := "success"
	alertClass := "success"
	if strings.Contains(insertResult, "FAILED") {
		insertClass = "error"
	}
	if strings.Contains(updateResult, "FAILED") {
		updateClass = "error"
	}
	if strings.Contains(alertResult, "FAILED") {
		alertClass = "error"
	}

	databaseOps := fmt.Sprintf(`<span class="%s">%s</span><br><span class="%s">%s</span>`,
		insertClass, insertResult, updateClass, updateResult)
	if alertResult != "N/A" {
		databaseOps += fmt.Sprintf(`<br><span class="%s">%s</span>`, alertClass, alertResult)
	}

	row := fmt.Sprintf(`        <tr>
            <td class="timestamp">%s</td>
            <td>%s</td>
            <td>%s</td>
            <td class="raw-data">%s</td>
            <td class="processed-data">%s</td>
            <td class="db-ops">%s</td>
        </tr>
`, currentTime, deviceID, topic, wrappedPayload, processedData, databaseOps)

	file.WriteString(row)
	log.Printf("Written HTML log for device %s", deviceID)
}

func wrapText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	var result strings.Builder
	for i := 0; i < len(text); i += maxLen {
		end := i + maxLen
		if end > len(text) {
			end = len(text)
		}
		result.WriteString(text[i:end])
		if end < len(text) {
			result.WriteString("<br>")
		}
	}
	return result.String()
}
