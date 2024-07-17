package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

var TEMPERATURE_KEYS = []string{"TAOL", "TB0T", "TB1T", "TB2T", "TCDX", "TCHP", "TCMb", "TCMz", "TD00", "TD01", "TD02", "TD03", "TD04", "TD10", "TD11", "TD12", "TD13", "TD14", "TD20", "TD21", "TD22", "TD23", "TD24", "TDBP", "TDEL", "TDER", "TDTC", "TDTP", "TDVx", "TDeL", "TDeR", "TED0", "TED1", "TFD0", "TFD1", "TG0B", "TG0C", "TG0H", "TG0V", "TG1B", "TG2B", "TH0a", "TH0b", "TH0x", "TMVR", "TPD0", "TPD1", "TPD2", "TPD3", "TPD4", "TPD5", "TPD6", "TPD7", "TPD8", "TPD9", "TPDX", "TPDa", "TPDb", "TPDc", "TPDd", "TPDe", "TPDf", "TPMP", "TPSP", "TR0Z", "TR1d", "TR2d", "TRD0", "TRD1", "TRD2", "TRD3", "TRD4", "TRD5", "TRD6", "TRD7", "TRD8", "TRD9", "TRDX", "TRDa", "TRDb", "TRDc", "TRDd", "TRDe", "TRDf", "TS0P", "TSCP", "TSG1", "TSG2", "TVA0", "TVD0", "TVH2", "TVH3", "TVHC", "TVHD", "TVHE", "TVHF", "TVHG", "TVHH", "TVHO", "TVMD", "TVS0", "TVS1", "TVS2", "TVS3", "TVSx", "TW0P", "Ta04", "Ta05", "Ta08", "Ta09", "Ta0C", "Ta0D", "Ta0G", "Ta0H", "Ta0K", "Ta0L", "TaLP", "TaLT", "TaLW", "TaRF", "TaRT", "TaRW", "TaTP", "Te04", "Te05", "Te06", "Te0G", "Te0H", "Te0I", "Te0P", "Te0Q", "Te0R", "Te0S", "Te0T", "Te0U", "Te0V", "Tf10", "Tf11", "Tf12", "Tf14", "Tf15", "Tf16", "Tf18", "Tf19", "Tf1A", "Tf1C", "Tf1D", "Tf1E", "Tf20", "Tf21", "Tf22", "Tf24", "Tf25", "Tf26", "Tf28", "Tf29", "Tf2C", "Tf2D", "Tf2E", "Tg00", "Tg01", "Tg04", "Tg05", "Tg0C", "Tg0D", "Tg0K", "Tg0L", "Tg0U", "Tg0V", "Tg0u", "Tg0v", "Tg12", "Tg13", "Tg1A", "Tg1B", "Tg1k", "Tg1l", "Th00", "Th01", "Th02", "Th04", "Th05", "Th06", "Th08", "Th09", "Th0A", "Th0C", "Th0D", "Th0E", "Tp04", "Tp05", "Tp06", "Tp0C", "Tp0D", "Tp0E", "Tp0K", "Tp0L", "Tp0M", "Tp0R", "Tp0S", "Tp0T", "Tp0U", "Tp0V", "Tp0W", "Tp0a", "Tp0b", "Tp0c", "Tp0g", "Tp0h", "Tp0i", "Tp0m", "Tp0n", "Tp0o", "Tp0u", "Tp0v", "Tp0w", "Tp0y", "Tp0z", "Tp10", "Tp16", "Tp17", "Tp18", "Tp1E", "Tp1F", "Tp1G", "Tp1I", "Tp1J", "Tp1K", "Tp1Q", "Tp1R", "Tp1S", "Tp3O", "Tp3P", "Tp3S", "Tp3T", "Tp3W", "Tp3X", "Ts00", "Ts01", "Ts02", "Ts0C", "Ts0D", "Ts0E", "Ts0K", "Ts0L", "Ts0M", "Ts0P", "Ts0Y", "Ts0Z", "Ts0a", "Ts0h", "Ts0i", "Ts1P", "Tz11", "Tz12", "Tz13", "Tz14", "Tz15", "Tz16", "Tz17", "Tz18", "Tz1j"}

var TEMPERATURE_DESC = make(map[string]*prometheus.Desc)

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

func (l *TemperatureCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range TEMPERATURE_DESC {
		ch <- desc
	}
}

func newFanCollector() *FanCollector {
	conn, result := gosmc.SMCOpen("AppleSMC")
	if result != 0 {
		log.Fatalf("Failed to open SMC connection: %d", result)
	}
	return &FanCollector{connection: conn}
}

func newTemperatureCollector() *TemperatureCollector {
	conn, result := gosmc.SMCOpen("AppleSMC")
	if result != 0 {
		log.Fatalf("Failed to open SMC connection: %d", result)
	}
	return &TemperatureCollector{connection: conn}
}

type (
	TemperatureCollector struct {
		connection uint
	}
)

func readSMCValue(connection uint, key string) (float64, error) {
	value, result := gosmc.SMCReadKey(connection, key)
	if result != 0 {
		return 0, fmt.Errorf("failed to read SMC key %s: %d", key, result)
	}

	// ref https://github.com/exelban/stats/blob/v2.11.2/SMC/smc.swift#L211 for parsing the data

	var floatValue float64
	var dataType = strings.TrimSpace(string(value.DataType[0:4]))

	switch dataType {
	case gosmc.TypeFLT:
		floatValue = float64(*(*float32)(unsafe.Pointer(&value.Bytes[0])))
	case gosmc.TypeUI8:
		floatValue = float64(*(*uint8)(unsafe.Pointer(&value.Bytes[0])))
	default:

		return 0, fmt.Errorf("unsupported data type %s for key %s", dataType, key)
	}
	return floatValue, nil
}

func (l *FanCollector) Collect(ch chan<- prometheus.Metric) {

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
	//testTemperature(l.connection)
}

func (l *TemperatureCollector) Collect(ch chan<- prometheus.Metric) {
	for _, key := range TEMPERATURE_KEYS {

		value, err := readSMCValue(l.connection, key)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(TEMPERATURE_DESC[key], prometheus.GaugeValue, value)
		} else {
			log.Printf("Failed to read %s: %v", key, err)
		}
	}

}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9259", "Address on which to expose metrics and web interface.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)

	for _, key := range TEMPERATURE_KEYS {
		TEMPERATURE_DESC[key] = prometheus.NewDesc(
			"smc_temp_"+key,
			key,
			[]string{},
			nil)
	}

	flag.Parse()

	fanCollector := newFanCollector()
	var temperatureCollector = newTemperatureCollector()
	prometheus.MustRegister(fanCollector)
	prometheus.MustRegister(temperatureCollector)

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
