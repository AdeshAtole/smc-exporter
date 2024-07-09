package main

import (
	"bufio"
	"bytes"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
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
	}
)

func (l *FanCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- actualFanSpeedDesc
	ch <- maxFanSpeedDesc
	ch <- minFanSpeedDesc
	ch <- targetFanSpeedDesc
}

func newFanCollector() *FanCollector {
	return &FanCollector{}
}

func getSMCValues() map[string]float64 {
	cmd := exec.Command("/Applications/Stats.app/Contents/Resources/smc", "list", "-f")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil
	}

	values := make(map[string]float64)
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 2 {
			key := strings.Trim(parts[0], "[]")
			value, err := strconv.ParseFloat(parts[1], 64)
			if err == nil {
				values[key] = value
			}
		}
	}

	return values
}

func (l *FanCollector) Collect(ch chan<- prometheus.Metric) {
	//for _, chip := range gosensors.GetDetectedChips() {
	//	chipName := chip.String()
	//	adaptorName := chip.AdapterName()
	//	for _, feature := range chip.GetFeatures() {
	//		if strings.HasPrefix(feature.Name, "fan") {
	//			ch <- prometheus.MustNewConstMetric(fanspeedDesc,
	//				prometheus.GaugeValue,
	//				feature.GetValue(),
	//				feature.GetLabel(), chipName, adaptorName)
	//		} else if strings.HasPrefix(feature.Name, "temp") {
	//			ch <- prometheus.MustNewConstMetric(temperatureDesc,
	//				prometheus.GaugeValue,
	//				feature.GetValue(),
	//				feature.GetLabel(), chipName, adaptorName)
	//		} else if strings.HasPrefix(feature.Name, "in") {
	//			ch <- prometheus.MustNewConstMetric(voltageDesc,
	//				prometheus.GaugeValue,
	//				feature.GetValue(),
	//				feature.GetLabel(), chipName, adaptorName)
	//		} else if strings.HasPrefix(feature.Name, "power") {
	//			ch <- prometheus.MustNewConstMetric(powerDesc,
	//				prometheus.GaugeValue,
	//				feature.GetValue(),
	//				feature.GetLabel(), chipName, adaptorName)
	//		}
	//	}
	//}

	//ch <- prometheus.MustNewConstMetric(maxFanSpeedDesc,
	//	prometheus.GaugeValue,
	//	100,
	//	"0")
	//ch <- prometheus.MustNewConstMetric(minFanSpeedDesc,
	//	prometheus.GaugeValue,
	//	0,
	//	"0")
	//ch <- prometheus.MustNewConstMetric(actualFanSpeedDesc,
	//	prometheus.GaugeValue,
	//	49,
	//	"0")
	//ch <- prometheus.MustNewConstMetric(targetFanSpeedDesc,
	//	prometheus.GaugeValue,
	//	50,
	//	"0")

	values := getSMCValues()
	if values == nil {
		return
	}

	fanCountValue := values["FNum"]
	ch <- prometheus.MustNewConstMetric(fanCount, prometheus.GaugeValue, fanCountValue)

	for i := 0; i < int(fanCountValue); i++ {
		index := strconv.Itoa(i)
		ch <- prometheus.MustNewConstMetric(maxFanSpeedDesc, prometheus.GaugeValue, values["F"+index+"Mx"], index)
		ch <- prometheus.MustNewConstMetric(minFanSpeedDesc, prometheus.GaugeValue, values["F"+index+"Mn"], index)
		ch <- prometheus.MustNewConstMetric(actualFanSpeedDesc, prometheus.GaugeValue, values["F"+index+"Ac"], index)
		ch <- prometheus.MustNewConstMetric(targetFanSpeedDesc, prometheus.GaugeValue, values["F"+index+"Tg"], index)
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
