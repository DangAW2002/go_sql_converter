#!/bin/bash

# Script to install and manage systemd service for MQTT converter

SERVICE_NAME="mqtt-converter"
SERVICE_FILE="mqtt-converter.service"
SYSTEMD_PATH="/etc/systemd/system"

install_service() {
    echo "Installing systemd service..."
    
    # Build the program first
    cd ~/go_sql_converter
    echo "Building program..."
    go mod tidy
    mkdir -p build
    go build -o build/main ./cmd
    
    if [ $? -ne 0 ]; then
        echo "Build failed. Aborting service installation."
        exit 1
    fi
    
    # Copy service file
    sudo cp scripts/$SERVICE_FILE $SYSTEMD_PATH/
    
    # Reload systemd
    sudo systemctl daemon-reload
    
    # Enable service (auto-start on boot)
    sudo systemctl enable $SERVICE_NAME
    
    echo "Service installed and enabled for auto-start."
    echo "Use the following commands to manage the service:"
    echo "  Start:   sudo systemctl start $SERVICE_NAME"
    echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
    echo "  Status:  sudo systemctl status $SERVICE_NAME"
    echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
}

uninstall_service() {
    echo "Uninstalling systemd service..."
    
    # Stop and disable service
    sudo systemctl stop $SERVICE_NAME 2>/dev/null || true
    sudo systemctl disable $SERVICE_NAME 2>/dev/null || true
    
    # Remove service file
    sudo rm -f $SYSTEMD_PATH/$SERVICE_FILE
    
    # Reload systemd
    sudo systemctl daemon-reload
    
    echo "Service uninstalled."
}

start_service() {
    sudo systemctl start $SERVICE_NAME
    echo "Service started."
}

stop_service() {
    sudo systemctl stop $SERVICE_NAME
    echo "Service stopped."
}

restart_service() {
    sudo systemctl restart $SERVICE_NAME
    echo "Service restarted."
}

show_status() {
    sudo systemctl status $SERVICE_NAME
}

show_logs() {
    sudo journalctl -u $SERVICE_NAME -f
}

case "$1" in
    install)
        install_service
        ;;
    uninstall)
        uninstall_service
        ;;
    start)
        start_service
        ;;
    stop)
        stop_service
        ;;
    restart)
        restart_service
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    *)
        echo "Usage: $0 {install|uninstall|start|stop|restart|status|logs}"
        echo ""
        echo "  install   - Install and enable service for auto-start"
        echo "  uninstall - Remove service"
        echo "  start     - Start service"
        echo "  stop      - Stop service"
        echo "  restart   - Restart service"
        echo "  status    - Show service status"
        echo "  logs      - Show service logs"
        exit 1
        ;;
esac