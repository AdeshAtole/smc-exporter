
# SMC Prometheus Exporter for Apple Silicon Macs

This Prometheus exporter exposes metrics from the System Management Controller (SMC) on Apple Silicon Macs. It provides information about fan speeds and various temperature sensors.

Despite the System Management Controller (SMC) hardware being exclusive to Intel Macs, the API retains the same name, and this project adheres to that convention.

## Features

- Exposes fan metrics:
  - Fan count
  - Maximum fan speed
  - Minimum fan speed
  - Actual fan speed
  - Target fan speed
- Exposes temperature metrics from numerous sensors



## Running the exporter

1. Clone the repository:
```
git clone https://github.com/yourusername/smc-prometheus-exporter.git
```

2. Navigate to the project directory:
```
cd smc-prometheus-exporter
```

3. Build the exporter:
```
go build
```

## Usage

Run the exporter:

```
./smc-exporter
```

By default, the exporter listens on `:9259` and exposes metrics at `/metrics`.

### Command-line flags

- `--web.listen-address`: The address to listen on for HTTP requests (default: ":9259")
- `--web.telemetry-path`: Path under which to expose metrics (default: "/metrics")

## Metrics

The exporter provides the following metrics:

- `smc_fan_count`: Number of fans
- `smc_max_fan_speed_rpm`: Maximum fan speed in RPM
- `smc_min_fan_speed_rpm`: Minimum fan speed in RPM
- `smc_target_fan_speed_rpm`: Target fan speed in RPM
- `smc_actual_fan_speed_rpm`: Actual fan speed in RPM
- `smc_temp_*`: Various temperature sensors (e.g., `smc_temp_TCMB`, `smc_temp_TA0P`, etc.)

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'smc'
    static_configs:
      - targets: ['localhost:9259']
```


## Acknowledgments

This exporter uses the [gosmc](https://github.com/panotza/gosmc) to interact with the SMC.

Thanks to [stats](https://github.com/exelban/stats) and hackers around the internet for the SMC keys.

