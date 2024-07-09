package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/panotza/gosmc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	maxFanSpeedDesc = prometheus.NewDesc(
		"smc_max_fan_speed_rpm",
		"max fan speed (rotations per minute).",
		[]string{"index"},
		nil)

	minFanSpeedDesc = prometheus.NewDesc(
		"smc_min_fan_speed_rpm",
		"min fan speed (rotations per minute).",
		[]string{"index"},
		nil)

	targetFanSpeedDesc = prometheus.NewDesc(
		"smc_target_fan_speed_rpm",
		"target fan speed (rotations per minute).",
		[]string{"index"},
		nil)

	actualFanSpeedDesc = prometheus.NewDesc(
		"smc_actual_fan_speed_rpm",
		"actual fan speed (rotations per minute).",
		[]string{"index"},
		nil)

	fanCount = prometheus.NewDesc(
		"smc_fan_count",
		"number of fans",
		[]string{},
		nil)
)

type (
	FanCollector struct {
		connection uint
	}
)

func (l *FanCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- actualFanSpeedDesc
	ch <- maxFanSpeedDesc
	ch <- minFanSpeedDesc
	ch <- targetFanSpeedDesc
	ch <- fanCount
}

func newFanCollector() *FanCollector {
	conn, result := gosmc.SMCOpen("AppleSMC")
	if result != 0 {
		log.Fatalf("Failed to open SMC connection: %d", result)
	}
	return &FanCollector{connection: conn}
}

func readSMCValue(connection uint, key string) (float64, error) {
	value, result := gosmc.SMCReadKey(connection, key)
	if result != 0 {
		return 0, fmt.Errorf("failed to read SMC key %s: %d", key, result)
	}

	var floatValue float64
	switch value.DataSize {
	case 4:
		floatValue = float64(*(*float32)(unsafe.Pointer(&value.Bytes[0])))
	case 8:
		floatValue = *(*float64)(unsafe.Pointer(&value.Bytes[0]))
	case 1:
		floatValue = float64(*(*uint8)(unsafe.Pointer(&value.Bytes[0])))
	default:

		return 0, fmt.Errorf("unexpected data size %d for key %s", value.DataSize, key)
	}
	return floatValue, nil
}

func (l *FanCollector) Collect(ch chan<- prometheus.Metric) {
	//defer gosmc.SMCClose(l.connection)

	fanCountValue, err := readSMCValue(l.connection, "FNum")
	if err != nil {
		log.Printf("Failed to read fan count: %v", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(fanCount, prometheus.GaugeValue, fanCountValue)

	for i := 0; i < int(fanCountValue); i++ {
		index := strconv.Itoa(i)
		maxFanSpeed, err := readSMCValue(l.connection, "F"+index+"Mx")
		if err == nil {
			ch <- prometheus.MustNewConstMetric(maxFanSpeedDesc, prometheus.GaugeValue, maxFanSpeed, index)
		} else {
			log.Printf("Failed to read max fan speed for fan %d: %v", i, err)
		}

		minFanSpeed, err := readSMCValue(l.connection, "F"+index+"Mn")
		if err == nil {
			ch <- prometheus.MustNewConstMetric(minFanSpeedDesc, prometheus.GaugeValue, minFanSpeed, index)
		} else {
			log.Printf("Failed to read min fan speed for fan %d: %v", i, err)
		}

		actualFanSpeed, err := readSMCValue(l.connection, "F"+index+"Ac")
		if err == nil {
			ch <- prometheus.MustNewConstMetric(actualFanSpeedDesc, prometheus.GaugeValue, actualFanSpeed, index)
		} else {
			log.Printf("Failed to read actual fan speed for fan %d: %v", i, err)
		}

		targetFanSpeed, err := readSMCValue(l.connection, "F"+index+"Tg")
		if err == nil {
			ch <- prometheus.MustNewConstMetric(targetFanSpeedDesc, prometheus.GaugeValue, targetFanSpeed, index)
		} else {
			log.Printf("Failed to read target fan speed for fan %d: %v", i, err)
		}
	}
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9259", "Address on which to expose metrics and web interface.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)

	flag.Parse()

	fanCollector := newFanCollector()
	prometheus.MustRegister(fanCollector)

	http.Handle(*metricsPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>SMC Exporter for Apple Silicon Macs</title></head>
			<body>
			<h1>SMC Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	http.ListenAndServe(*listenAddress, nil)
}
