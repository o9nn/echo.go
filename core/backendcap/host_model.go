package backendcap

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	gib                 = 1024 * 1024 * 1024
	maxModelProbeFiles  = 128
	maxGGUFMetadataKeys = 4096
)

// HostMemoryProbe describes the host memory envelope available to local inference.
type HostMemoryProbe struct {
	TotalBytes     uint64     `json:"total_bytes"`
	AvailableBytes uint64     `json:"available_bytes"`
	Tier           MemoryTier `json:"tier"`
	Reason         string     `json:"reason"`
}

// ProbeHostMemory inspects the host memory surface without importing native backends.
func ProbeHostMemory() HostMemoryProbe {
	fields, err := readMemInfo("/proc/meminfo")
	if err != nil {
		return HostMemoryProbe{Tier: MemoryConstrained, Reason: "host memory probe unavailable: " + err.Error()}
	}

	total := fields["MemTotal"] * 1024
	available := fields["MemAvailable"] * 1024
	if available == 0 {
		available = fields["MemFree"] * 1024
	}

	tier := memoryTierFromBytes(total)
	return HostMemoryProbe{
		TotalBytes:     total,
		AvailableBytes: available,
		Tier:           tier,
		Reason:         fmt.Sprintf("host reports %.1f GiB total and %.1f GiB available", bytesToGiB(total), bytesToGiB(available)),
	}
}

// HostMemoryTier returns the current host memory tier for scheduler decisions.
func HostMemoryTier() MemoryTier {
	return ProbeHostMemory().Tier
}

func readMemInfo(path string) (map[string]uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]uint64)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		name := strings.TrimSpace(parts[0])
		valueFields := strings.Fields(parts[1])
		if len(valueFields) == 0 {
			continue
		}
		value, err := strconv.ParseUint(valueFields[0], 10, 64)
		if err == nil {
			result[name] = value
		}
	}
	return result, nil
}

func memoryTierFromBytes(bytes uint64) MemoryTier {
	switch {
	case bytes >= 32*gib:
		return MemoryStress
	case bytes >= 8*gib:
		return MemoryStandard
	default:
		return MemoryConstrained
	}
}

func modelMemoryTier(bytes uint64) MemoryTier {
	switch {
	case bytes >= 16*gib:
		return MemoryStress
	case bytes >= 4*gib:
		return MemoryStandard
	default:
		return MemoryConstrained
	}
}

func bytesToGiB(bytes uint64) float64 {
	return float64(bytes) / float64(gib)
}

// ModelFileMetadata captures scheduler-relevant metadata discovered from a GGUF model file.
type ModelFileMetadata struct {
	Path                 string                 `json:"path"`
	Name                 string                 `json:"name"`
	Architecture         string                 `json:"architecture,omitempty"`
	FileSizeBytes        uint64                 `json:"file_size_bytes"`
	ContextLength        int                    `json:"context_length,omitempty"`
	Quantization         string                 `json:"quantization,omitempty"`
	EstimatedMemoryBytes uint64                 `json:"estimated_memory_bytes"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// ProbeModelFile reads a GGUF model file and returns it as a schedulable native capability.
func ProbeModelFile(path string) (Capability, error) {
	metadata, err := ReadGGUFModelMetadata(path)
	if err != nil {
		return Capability{}, err
	}

	host := ProbeHostMemory()
	available := true
	if host.AvailableBytes > 0 && metadata.EstimatedMemoryBytes > uint64(float64(host.AvailableBytes)*0.85) {
		available = false
	}

	reason := fmt.Sprintf("GGUF model %s requires approximately %.1f GiB; %s", metadata.Name, bytesToGiB(metadata.EstimatedMemoryBytes), host.Reason)
	if !available {
		reason = "model present but host memory appears insufficient: " + reason
	}

	return Capability{
		Name:                 "model:" + metadata.Name,
		Kind:                 BackendNativeCPU,
		Available:            available,
		Native:               true,
		Offline:              true,
		StressGrade:          modelMemoryTier(metadata.EstimatedMemoryBytes) == MemoryStress,
		MemoryTier:           modelMemoryTier(metadata.EstimatedMemoryBytes),
		ModelPath:            metadata.Path,
		ContextLength:        metadata.ContextLength,
		Quantization:         metadata.Quantization,
		EstimatedMemoryBytes: metadata.EstimatedMemoryBytes,
		Reason:               reason,
		Guidance:             "Use this specific local GGUF model when its context length and estimated memory footprint fit the workload and host memory envelope.",
	}, nil
}

// DiscoverModelCapabilities probes explicit GGUF files or directories and returns model capabilities.
func DiscoverModelCapabilities(paths []string) []Capability {
	files := discoverGGUFFiles(paths)
	caps := make([]Capability, 0, len(files))
	for _, file := range files {
		cap, err := ProbeModelFile(file)
		if err == nil {
			caps = append(caps, cap)
		}
	}
	sort.SliceStable(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
	return caps
}

func discoverGGUFFiles(paths []string) []string {
	seen := make(map[string]bool)
	files := make([]string, 0)
	for _, raw := range paths {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			if strings.EqualFold(filepath.Ext(path), ".gguf") && !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
			continue
		}
		_ = filepath.WalkDir(path, func(candidate string, entry os.DirEntry, err error) error {
			if err != nil || len(files) >= maxModelProbeFiles {
				return nil
			}
			if entry.IsDir() {
				name := entry.Name()
				if strings.HasPrefix(name, ".") && candidate != path {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.EqualFold(filepath.Ext(candidate), ".gguf") && !seen[candidate] {
				seen[candidate] = true
				files = append(files, candidate)
			}
			return nil
		})
	}
	sort.Strings(files)
	return files
}

// SnapshotWithModelPaths returns the normal backend snapshot enriched with specific local model files.
func SnapshotWithModelPaths(paths []string) []Capability {
	caps := append([]Capability{}, Snapshot()...)
	modelCaps := DiscoverModelCapabilities(paths)
	if !cgoEnabled {
		for i := range modelCaps {
			modelCaps[i].Available = false
			modelCaps[i].StressGrade = false
			modelCaps[i].Reason = "native model unavailable: cgo is disabled; " + modelCaps[i].Reason
		}
	}
	caps = append(caps, modelCaps...)
	sort.SliceStable(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
	return caps
}

// ReadGGUFModelMetadata reads the lightweight GGUF metadata needed for model routing.
func ReadGGUFModelMetadata(path string) (ModelFileMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return ModelFileMetadata{}, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return ModelFileMetadata{}, err
	}

	var magic [4]byte
	if _, err := io.ReadFull(file, magic[:]); err != nil {
		return ModelFileMetadata{}, err
	}
	if string(magic[:]) != "GGUF" {
		return ModelFileMetadata{}, fmt.Errorf("%s is not a GGUF file", path)
	}

	var version uint32
	if err := binary.Read(file, binary.LittleEndian, &version); err != nil {
		return ModelFileMetadata{}, err
	}

	var tensorCount uint64
	var kvCount uint64
	if err := binary.Read(file, binary.LittleEndian, &tensorCount); err != nil {
		return ModelFileMetadata{}, err
	}
	if err := binary.Read(file, binary.LittleEndian, &kvCount); err != nil {
		return ModelFileMetadata{}, err
	}
	if kvCount > maxGGUFMetadataKeys {
		return ModelFileMetadata{}, fmt.Errorf("GGUF metadata key count %d exceeds safety limit", kvCount)
	}

	metadata := make(map[string]interface{})
	for i := uint64(0); i < kvCount; i++ {
		key, err := readGGUFString(file)
		if err != nil {
			return ModelFileMetadata{}, fmt.Errorf("read GGUF key %d: %w", i, err)
		}
		var valueType uint32
		if err := binary.Read(file, binary.LittleEndian, &valueType); err != nil {
			return ModelFileMetadata{}, err
		}
		value, err := readGGUFValue(file, valueType)
		if err != nil {
			return ModelFileMetadata{}, fmt.Errorf("read GGUF metadata %s: %w", key, err)
		}
		metadata[key] = value
	}

	name := stringFromMetadata(metadata, "general.name")
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	arch := stringFromMetadata(metadata, "general.architecture")
	context := intFromMetadata(metadata, arch+".context_length", "llama.context_length", "context_length")
	quantization := quantizationFromMetadata(metadata)
	fileSize := uint64(stat.Size())
	estimate := estimateModelMemory(fileSize, context)

	metadata["gguf.version"] = uint64(version)
	metadata["gguf.tensor_count"] = tensorCount
	metadata["gguf.metadata_kv_count"] = kvCount

	return ModelFileMetadata{
		Path:                 path,
		Name:                 name,
		Architecture:         arch,
		FileSizeBytes:        fileSize,
		ContextLength:        context,
		Quantization:         quantization,
		EstimatedMemoryBytes: estimate,
		Metadata:             metadata,
	}, nil
}

func readGGUFString(r io.Reader) (string, error) {
	var length uint64
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	if length > 16*1024*1024 {
		return "", fmt.Errorf("GGUF string length %d exceeds safety limit", length)
	}
	buf := make([]byte, length)
	_, err := io.ReadFull(r, buf)
	return string(buf), err
}

func readGGUFValue(r io.Reader, valueType uint32) (interface{}, error) {
	switch valueType {
	case 0:
		var v uint8
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 1:
		var v int8
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 2:
		var v uint16
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 3:
		var v int16
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 4:
		var v uint32
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 5:
		var v int32
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 6:
		var v float32
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 7:
		var raw uint8
		if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
			return false, err
		}
		return raw != 0, nil
	case 8:
		return readGGUFString(r)
	case 9:
		var elemType uint32
		var length uint64
		if err := binary.Read(r, binary.LittleEndian, &elemType); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return nil, err
		}
		if length > 1_000_000 {
			return nil, fmt.Errorf("GGUF array length %d exceeds safety limit", length)
		}
		values := make([]interface{}, 0, int(math.Min(float64(length), 32)))
		for i := uint64(0); i < length; i++ {
			value, err := readGGUFValue(r, elemType)
			if err != nil {
				return nil, err
			}
			if i < 32 {
				values = append(values, value)
			}
		}
		return values, nil
	case 10:
		var v uint64
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 11:
		var v int64
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 12:
		var v float64
		return v, binary.Read(r, binary.LittleEndian, &v)
	default:
		return nil, fmt.Errorf("unsupported GGUF metadata type %d", valueType)
	}
}

func stringFromMetadata(metadata map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := metadata[key]; ok {
			if s, ok := value.(string); ok {
				return s
			}
		}
	}
	return ""
}

func intFromMetadata(metadata map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if key == ".context_length" {
			continue
		}
		if value, ok := metadata[key]; ok {
			switch v := value.(type) {
			case uint8:
				return int(v)
			case int8:
				return int(v)
			case uint16:
				return int(v)
			case int16:
				return int(v)
			case uint32:
				return int(v)
			case int32:
				return int(v)
			case uint64:
				if v <= uint64(^uint(0)>>1) {
					return int(v)
				}
			case int64:
				if v >= 0 && uint64(v) <= uint64(^uint(0)>>1) {
					return int(v)
				}
			}
		}
	}
	return 0
}

func quantizationFromMetadata(metadata map[string]interface{}) string {
	fileType := intFromMetadata(metadata, "general.file_type")
	if label, ok := ggufFileTypes[fileType]; ok {
		return label
	}
	if fileType != 0 {
		return fmt.Sprintf("file_type_%d", fileType)
	}
	return "unknown"
}

func estimateModelMemory(fileSize uint64, contextLength int) uint64 {
	estimate := uint64(float64(fileSize) * 1.20)
	if contextLength > 0 {
		estimate += uint64(contextLength) * 512 * 1024
	}
	if estimate < fileSize {
		return fileSize
	}
	return estimate
}

var ggufFileTypes = map[int]string{
	0:  "F32",
	1:  "F16",
	2:  "Q4_0",
	3:  "Q4_1",
	6:  "Q5_0",
	7:  "Q5_1",
	8:  "Q8_0",
	9:  "Q8_1",
	10: "Q2_K",
	11: "Q3_K_S",
	12: "Q3_K_M",
	13: "Q3_K_L",
	14: "Q4_K_S",
	15: "Q4_K_M",
	16: "Q5_K_S",
	17: "Q5_K_M",
	18: "Q6_K",
	19: "IQ2_XXS",
	20: "IQ2_XS",
	21: "Q2_K_S",
	22: "IQ3_XS",
	23: "IQ3_XXS",
	24: "IQ1_S",
	25: "IQ4_NL",
	26: "IQ3_S",
	27: "IQ3_M",
	28: "IQ2_S",
	29: "IQ2_M",
	30: "IQ4_XS",
	31: "IQ1_M",
}
