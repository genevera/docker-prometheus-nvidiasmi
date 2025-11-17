package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

const LISTEN_ADDRESS = ":9202"
const NVIDIA_SMI_PATH = "/usr/bin/nvidia-smi"

var testMode string

type processInfo struct {
	Pid         string `xml:"pid"`
	Type        string `xml:"type"`
	ProcessName string `xml:"process_name"`
	UsedMemory  string `xml:"used_memory"`
}
type gpuInfo struct {
	Id                       string `xml:"id,attr"`
	ProductName              string `xml:"product_name"`
	ProductBrand             string `xml:"product_brand"`
	DisplayMode              string `xml:"display_mode"`
	DisplayActive            string `xml:"display_active"`
	PersistenceMode          string `xml:"persistence_mode"`
	AccountingMode           string `xml:"accounting_mode"`
	AccountingModeBufferSize string `xml:"accounting_mode_buffer_size"`
	DriverModel              struct {
		CurrentDM string `xml:"current_dm"`
		PendingDM string `xml:"pending_dm"`
	} `xml:"driver_model"`
	Serial         string `xml:"serial"`
	UUID           string `xml:"uuid"`
	MinorNumber    string `xml:"minor_number"`
	VbiosVersion   string `xml:"vbios_version"`
	MultiGPUBoard  string `xml:"multigpu_board"`
	BoardId        string `xml:"board_id"`
	GPUPartNumber  string `xml:"gpu_part_number"`
	InfoRomVersion struct {
		ImgVersion string `xml:"img_version"`
		OemObject  string `xml:"oem_object"`
		EccObject  string `xml:"ecc_object"`
		PwrObject  string `xml:"pwr_object"`
	} `xml:"inforom_version"`
	GPUOperationMode struct {
		Current string `xml:"current_gom"`
		Pending string `xml:"pending_gom"`
	} `xml:"gpu_operation_mode"`
	GPUVirtualizationMode struct {
		VirtualizationMode string `xml:"virtualization_mode"`
		HostVGPUMode       string `xml:"host_vgpu_mode"`
	} `xml:"gpu_virtualization_mode"`
	IBMNPU struct {
		RelaxedOrderingMode string `xml:"relaxed_ordering_mode"`
	} `xml:"ibmnpu"`
	PCI struct {
		Bus         string `xml:"pci_bus"`
		Device      string `xml:"pci_device"`
		Domain      string `xml:"pci_domain"`
		DeviceId    string `xml:"pci_device_id"`
		BusId       string `xml:"pci_bus_id"`
		SubSystemId string `xml:"pci_sub_system_id"`
		GPULinkInfo struct {
			PCIeGen struct {
				Max     string `xml:"max_link_gen"`
				Current string `xml:"current_link_gen"`
			} `xml:"pcie_gen"`
			LinkWidth struct {
				Max     string `xml:"max_link_width"`
				Current string `xml:"current_link_width"`
			} `xml:"link_widths"`
		} `xml:"pci_gpu_link_info"`
		BridgeChip struct {
			Type string `xml:"bridge_chip_type"`
			Fw   string `xml:"bridge_chip_fw"`
		} `xml:"pci_bridge_chip"`
		ReplayCounter         string `xml:"replay_counter"`
		ReplayRolloverCounter string `xml:"replay_rollover_counter"`
		TxUtil                string `xml:"tx_util"`
		RxUtil                string `xml:"rx_util"`
	} `xml:"pci"`
	FanSpeed         string `xml:"fan_speed"`
	PerformanceState string `xml:"performance_state"`
	FbMemoryUsage    struct {
		Total string `xml:"total"`
		Used  string `xml:"used"`
		Free  string `xml:"free"`
	} `xml:"fb_memory_usage"`
	Bar1MemoryUsage struct {
		Total string `xml:"total"`
		Used  string `xml:"used"`
		Free  string `xml:"free"`
	} `xml:"bar1_memory_usage"`
	ComputeMode string `xml:"compute_mode"`
	Utilization struct {
		GPUUtil     string `xml:"gpu_util"`
		MemoryUtil  string `xml:"memory_util"`
		EncoderUtil string `xml:"encoder_util"`
		DecoderUtil string `xml:"decoder_util"`
	} `xml:"utilization"`
	EncoderStats struct {
		SessionCount   string `xml:"session_count"`
		AverageFPS     string `xml:"average_fps"`
		AverageLatency string `xml:"average_latency"`
	} `xml:"encoder_stats"`
	FBCStats struct {
		SessionCount   string `xml:"session_count"`
		AverageFPS     string `xml:"average_fps"`
		AverageLatency string `xml:"average_latency"`
	} `xml:"fbc_stats"`
	Temperature struct {
		GPUTemp                string `xml:"gpu_temp"`
		GPUTempMaxThreshold    string `xml:"gpu_temp_max_threshold"`
		GPUTempSlowThreshold   string `xml:"gpu_temp_slow_threshold"`
		GPUTempMaxGpuThreshold string `xml:"gpu_temp_max_gpu_threshold"`
		MemoryTemp             string `xml:"memory_temp"`
		GPUTempMaxMemThreshold string `xml:"gpu_temp_max_mem_threshold"`
	} `xml:"temperature"`
	PowerReadings struct {
		PowerState          string `xml:"power_state"`
		AveragePowerDraw    string `xml:"average_power_draw"`
		InstantPowerDraw    string `xml:"instant_power_draw"`
		CurrentPowerLimit   string `xml:"current_power_limit"`
		DefaultPowerLimit   string `xml:"default_power_limit"`
		RequestedPowerLimit string `xml:"requested_power_limit"`
		MinPowerLimit       string `xml:"min_power_limit"`
		MaxPowerLimit       string `xml:"max_power_limit"`
	} `xml:"gpu_power_readings"`
	Clocks struct {
		GraphicsClock string `xml:"graphics_clock"`
		SmClock       string `xml:"sm_clock"`
		MemClock      string `xml:"mem_clock"`
		VideoClock    string `xml:"video_clock"`
	} `xml:"clocks"`
	MaxClocks struct {
		GraphicsClock string `xml:"graphics_clock"`
		SmClock       string `xml:"sm_clock"`
		MemClock      string `xml:"mem_clock"`
		VideoClock    string `xml:"video_clock"`
	} `xml:"max_clocks"`
	ClockPolicy struct {
		AutoBoost        string `xml:"auto_boost"`
		AutoBoostDefault string `xml:"auto_boost_default"`
	} `xml:"clock_policy"`
	Processes struct {
		ProcessInfo []processInfo `xml:"process_info"`
	} `xml:"processes"`
}

type NvidiaSmiLog struct {
	DriverVersion string    `xml:"driver_version"`
	CudaVersion   string    `xml:"cuda_version"`
	AttachedGPUs  string    `xml:"attached_gpus"`
	GPU           []gpuInfo `xml:"gpu"`
}

func formatVersion(value string) string {
	r := regexp.MustCompile(`(?P<version>\d+\.\d+).*`)
	match := r.FindStringSubmatch(value)
	version := "0"
	if len(match) > 0 {
		version = match[1]
	}
	return version
}

func formatValue(key string, meta string, value string) string {
	result := key
	if meta != "" {
		result += "{" + meta + "}"
	}
	return result + " " + value + "\n"
}

func filterUnit(s string) string {
	if s == "N/A" {
		return "0"
	}
	r := regexp.MustCompile(`(?P<value>[\d\.]+) (?P<power>[KMGT]?[i]?)(?P<unit>.*)`)
	match := r.FindStringSubmatch(s)
	if len(match) == 0 {
		return "0"
	}

	result := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	power := result["power"]
	if value, err := strconv.ParseFloat(result["value"], 32); err == nil {
		switch power {
		case "K":
			value *= 1000
		case "M":
			value *= 1000 * 1000
		case "G":
			value *= 1000 * 1000 * 1000
		case "T":
			value *= 1000 * 1000 * 1000 * 1000
		case "Ki":
			value *= 1024
		case "Mi":
			value *= 1024 * 1024
		case "Gi":
			value *= 1024 * 1024 * 1024
		case "Ti":
			value *= 1024 * 1024 * 1024 * 1024
		}
		return fmt.Sprintf("%g", value)
	}
	return "0"
}

func filterNumber(value string) string {
	if value == "N/A" {
		return "0"
	}
	r := regexp.MustCompile("[^0-9.]")
	return r.ReplaceAllString(value, "")
}

func formatMetaGpu(gpu *gpuInfo) string {
	return fmt.Sprintf("id=\"%v\",uuid=\"%v\",name=\"%v\"", gpu.Id, gpu.UUID, gpu.ProductName)
}

func formatMetaProcess(gpu *gpuInfo, process *processInfo) string {
	return fmt.Sprintf("%v,process_pid=\"%v\",process_type=\"%v\"", formatMetaGpu(gpu), process.Pid, process.Type)
}

func writeMetric(writer io.Writer, key string, meta string, value string) {
	_, err := io.WriteString(writer, formatValue(key, meta, value))
	if err != nil {
		panic(fmt.Sprintf("could not write metric %v:%v", key, err))
	}
}

func metrics(w http.ResponseWriter, r *http.Request) {
	log.Print("Serving /metrics")

	var cmd *exec.Cmd
	if testMode == "1" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command("/bin/cat", dir+"/nvidia-smi.sample.xml")
	} else {
		cmd = exec.Command(NVIDIA_SMI_PATH, "-q", "-x")
	}

	// Execute system command
	stdout, err := cmd.Output()
	if err != nil {
		println(err.Error())
		if testMode != "1" {
			println("Something went wrong with the execution of nvidia-smi")
		}
		return
	}

	// Parse XML
	var xmlData NvidiaSmiLog
	xml.Unmarshal(stdout, &xmlData)

	// Output
	for _, GPU := range xmlData.GPU {
		meta := formatMetaGpu(&GPU)

		writeMetric(w, "nvidiasmi_driver_version", meta, formatVersion(xmlData.DriverVersion))
		writeMetric(w, "nvidiasmi_cuda_version", meta, formatVersion(xmlData.CudaVersion))
		writeMetric(w, "nvidiasmi_attached_gpus", meta, xmlData.AttachedGPUs)
		writeMetric(w, "nvidiasmi_pci_pcie_gen_max", meta, GPU.PCI.GPULinkInfo.PCIeGen.Max)
		writeMetric(w, "nvidiasmi_pci_pcie_gen_current", meta, GPU.PCI.GPULinkInfo.PCIeGen.Current)
		writeMetric(w, "nvidiasmi_pci_link_width_max_multiplicator", meta, filterNumber(GPU.PCI.GPULinkInfo.LinkWidth.Max))
		writeMetric(w, "nvidiasmi_pci_link_width_current_multiplicator", meta, filterNumber(GPU.PCI.GPULinkInfo.LinkWidth.Current))
		writeMetric(w, "nvidiasmi_pci_replay_counter", meta, GPU.PCI.ReplayRolloverCounter)
		writeMetric(w, "nvidiasmi_pci_replay_rollover_counter", meta, GPU.PCI.ReplayRolloverCounter)
		writeMetric(w, "nvidiasmi_pci_tx_util_bytes_per_second", meta, filterUnit(GPU.PCI.TxUtil))
		writeMetric(w, "nvidiasmi_pci_rx_util_bytes_per_second", meta, filterUnit(GPU.PCI.RxUtil))
		writeMetric(w, "nvidiasmi_fan_speed_percent", meta, filterUnit(GPU.FanSpeed))
		writeMetric(w, "nvidiasmi_performance_state_int", meta, filterNumber(GPU.PerformanceState))
		writeMetric(w, "nvidiasmi_fb_memory_usage_total_bytes", meta, filterUnit(GPU.FbMemoryUsage.Total))
		writeMetric(w, "nvidiasmi_fb_memory_usage_used_bytes", meta, filterUnit(GPU.FbMemoryUsage.Used))
		writeMetric(w, "nvidiasmi_fb_memory_usage_free_bytes", meta, filterUnit(GPU.FbMemoryUsage.Free))
		writeMetric(w, "nvidiasmi_bar1_memory_usage_total_bytes", meta, filterUnit(GPU.Bar1MemoryUsage.Total))
		writeMetric(w, "nvidiasmi_bar1_memory_usage_used_bytes", meta, filterUnit(GPU.Bar1MemoryUsage.Used))
		writeMetric(w, "nvidiasmi_bar1_memory_usage_free_bytes", meta, filterUnit(GPU.Bar1MemoryUsage.Free))
		writeMetric(w, "nvidiasmi_utilization_gpu_percent", meta, filterUnit(GPU.Utilization.GPUUtil))
		writeMetric(w, "nvidiasmi_utilization_memory_percent", meta, filterUnit(GPU.Utilization.MemoryUtil))
		writeMetric(w, "nvidiasmi_utilization_encoder_percent", meta, filterUnit(GPU.Utilization.EncoderUtil))
		writeMetric(w, "nvidiasmi_utilization_decoder_percent", meta, filterUnit(GPU.Utilization.DecoderUtil))
		writeMetric(w, "nvidiasmi_encoder_session_count", meta, GPU.EncoderStats.SessionCount)
		writeMetric(w, "nvidiasmi_encoder_average_fps", meta, GPU.EncoderStats.AverageFPS)
		writeMetric(w, "nvidiasmi_encoder_average_latency", meta, GPU.EncoderStats.AverageLatency)
		writeMetric(w, "nvidiasmi_fbc_session_count", meta, GPU.FBCStats.SessionCount)
		writeMetric(w, "nvidiasmi_fbc_average_fps", meta, GPU.FBCStats.AverageFPS)
		writeMetric(w, "nvidiasmi_fbc_average_latency", meta, GPU.FBCStats.AverageLatency)
		writeMetric(w, "nvidiasmi_gpu_temp_celsius", meta, filterUnit(GPU.Temperature.GPUTemp))
		writeMetric(w, "nvidiasmi_gpu_temp_max_threshold_celsius", meta, filterUnit(GPU.Temperature.GPUTempMaxThreshold))
		writeMetric(w, "nvidiasmi_gpu_temp_slow_threshold_celsius", meta, filterUnit(GPU.Temperature.GPUTempSlowThreshold))
		writeMetric(w, "nvidiasmi_gpu_temp_max_gpu_threshold_celsius", meta, filterUnit(GPU.Temperature.GPUTempMaxGpuThreshold))
		writeMetric(w, "nvidiasmi_memory_temp_celsius", meta, filterUnit(GPU.Temperature.MemoryTemp))
		writeMetric(w, "nvidiasmi_gpu_temp_max_mem_threshold_celsius", meta, filterUnit(GPU.Temperature.GPUTempMaxMemThreshold))
		writeMetric(w, "nvidiasmi_power_state_int", meta, filterNumber(GPU.PowerReadings.PowerState))
		writeMetric(w, "nvidiasmi_power_instant_draw_watts", meta, filterUnit(GPU.PowerReadings.InstantPowerDraw))
		writeMetric(w, "nvidiasmi_power_average_draw_watts", meta, filterUnit(GPU.PowerReadings.AveragePowerDraw))
		writeMetric(w, "nvidiasmi_current_power_limit_watts", meta, filterUnit(GPU.PowerReadings.CurrentPowerLimit))
		writeMetric(w, "nvidiasmi_default_power_limit_watts", meta, filterUnit(GPU.PowerReadings.DefaultPowerLimit))
		writeMetric(w, "nvidiasmi_requested_power_limit_watts", meta, filterUnit(GPU.PowerReadings.RequestedPowerLimit))
		writeMetric(w, "nvidiasmi_min_power_limit_watts", meta, filterUnit(GPU.PowerReadings.MinPowerLimit))
		writeMetric(w, "nvidiasmi_max_power_limit_watts", meta, filterUnit(GPU.PowerReadings.MaxPowerLimit))
		writeMetric(w, "nvidiasmi_clock_graphics_hertz", meta, filterUnit(GPU.Clocks.GraphicsClock))
		writeMetric(w, "nvidiasmi_clock_graphics_max_hertz", meta, filterUnit(GPU.MaxClocks.GraphicsClock))
		writeMetric(w, "nvidiasmi_clock_sm_hertz", meta, filterUnit(GPU.Clocks.SmClock))
		writeMetric(w, "nvidiasmi_clock_sm_max_hertz", meta, filterUnit(GPU.MaxClocks.SmClock))
		writeMetric(w, "nvidiasmi_clock_mem_hertz", meta, filterUnit(GPU.Clocks.MemClock))
		writeMetric(w, "nvidiasmi_clock_mem_max_hertz", meta, filterUnit(GPU.MaxClocks.MemClock))
		writeMetric(w, "nvidiasmi_clock_video_hertz", meta, filterUnit(GPU.Clocks.VideoClock))
		writeMetric(w, "nvidiasmi_clock_video_max_hertz", meta, filterUnit(GPU.MaxClocks.VideoClock))
		writeMetric(w, "nvidiasmi_clock_policy_auto_boost", meta, filterUnit(GPU.ClockPolicy.AutoBoost))
		writeMetric(w, "nvidiasmi_clock_policy_auto_boost_default", meta, filterUnit(GPU.ClockPolicy.AutoBoostDefault))

		for _, Process := range GPU.Processes.ProcessInfo {
			writeMetric(w, "nvidiasmi_process_used_memory_bytes", formatMetaProcess(&GPU, &Process), filterUnit(Process.UsedMemory))
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Print("Serving /index")
	html := `<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Nvidia SMI Exporter</title>
    </head>
    <body>
        <h1>Nvidia SMI Exporter</h1>
        <p><a href="/metrics">Metrics</a></p>
    </body>
</html>`
	io.WriteString(w, html)
}

func main() {
	testMode = os.Getenv("TEST_MODE")
	if testMode == "1" {
		log.Print("Test mode is enabled")
	}

	log.Print("Nvidia SMI exporter listening on " + LISTEN_ADDRESS)
	http.HandleFunc("/", index)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(LISTEN_ADDRESS, nil)
}
