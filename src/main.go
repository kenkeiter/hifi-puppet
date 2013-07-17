package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

var hostAddr = flag.String("host", ":8192", "host:port for client-side UI")
var assetPath = flag.String("asset_path", "src/www", "relative or absolute path to UI asset directory")
var imuPort = flag.String("imu_port", "/dev/tty.usbmodem1411", "absolute path to sensor device (e.g. /dev/tty.<port>)")

func main() {
	flag.Parse()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Instantiate the sensor and AHRS filter.
	fmt.Print("Connecting to sensor... ")
	sensor := CreateMPU9150_IMUSensor(*imuPort, 115200)
	sensor.Connect()
	fmt.Print("Done.\n")

	ahrs := NewAHRSFilter(sensor, 0.0001, 60)

	// Setup the server.
	go func(ahrs *AHRSFilter) {
		http.Handle("/", http.FileServer(http.Dir(*assetPath)))
		http.Handle("/motion", websocket.Handler(AHRSMotionServer(ahrs)))
		if err := http.ListenAndServe(*hostAddr, nil); err != nil {
			panic(err)
		}
	}(ahrs)

	for sig := range signals {
		fmt.Printf("Received %s signal; shutting down.\n", sig)
		sensor.Disconnect()
		os.Exit(0)
	}
}
