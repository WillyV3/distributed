package host

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// LoadInfo contains host load metrics
type LoadInfo struct {
	Host      string
	Load      float64
	CPUs      int
	CPUPct    int
	MemPct    int
	Score     float64
	Reachable bool
}

// CheckReachable tests if a host is reachable via SSH
func CheckReachable(host string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh",
		"-o", "ConnectTimeout=2",
		"-o", "BatchMode=yes",
		host, "exit")

	return cmd.Run() == nil
}

// GetLoad retrieves load information from a host
func GetLoad(host string) (*LoadInfo, error) {
	// Check if reachable first
	if !CheckReachable(host, 2*time.Second) {
		return &LoadInfo{
			Host:      host,
			Reachable: false,
		}, nil
	}

	// Get load metrics via SSH
	script := `
load=$(uptime | sed "s/.*load average[s]*: //" | awk "{print \$1}" | tr -d ",")

if [[ "$OSTYPE" == "darwin"* ]]; then
    cpus=$(sysctl -n hw.ncpu)
    mem_total=$(sysctl -n hw.memsize)
    page_size=$(vm_stat | grep "page size" | awk "{print \$8}")
    pages_wired=$(vm_stat | grep "Pages wired" | awk "{print \$4}" | tr -d ".")
    pages_active=$(vm_stat | grep "Pages active" | awk "{print \$3}" | tr -d ".")
    pages_compressed=$(vm_stat | grep "occupied by compressor" | awk "{print \$5}" | tr -d ".")
    mem_used=$(( (pages_wired + pages_active + pages_compressed) * page_size ))
    mem_pct=$((mem_used * 100 / mem_total))
else
    cpus=$(nproc)
    mem_info=$(free | grep Mem)
    mem_total=$(echo "$mem_info" | awk "{print \$2}")
    mem_used=$(echo "$mem_info" | awk "{print \$3}")
    mem_pct=$((mem_used * 100 / mem_total))
fi

cpu_pct=$(awk "BEGIN {printf \"%.0f\", ($load / $cpus) * 100}")
score=$(awk "BEGIN {printf \"%.2f\", ($cpu_pct * 0.7) + ($mem_pct * 0.3)}")

echo "$load|$cpus|$cpu_pct|$mem_pct|$score"
`

	cmd := exec.Command("ssh", "-o", "LogLevel=QUIET", host, script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get load: %w", err)
	}

	// Parse output: load|cpus|cpu_pct|mem_pct|score
	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) != 5 {
		return nil, fmt.Errorf("unexpected output format")
	}

	load, _ := strconv.ParseFloat(parts[0], 64)
	cpus, _ := strconv.Atoi(parts[1])
	cpuPct, _ := strconv.Atoi(parts[2])
	memPct, _ := strconv.Atoi(parts[3])
	score, _ := strconv.ParseFloat(parts[4], 64)

	return &LoadInfo{
		Host:      host,
		Load:      load,
		CPUs:      cpus,
		CPUPct:    cpuPct,
		MemPct:    memPct,
		Score:     score,
		Reachable: true,
	}, nil
}

// FindBest finds the host with the lowest load score
func FindBest(hosts []string) (*LoadInfo, error) {
	var best *LoadInfo

	for _, host := range hosts {
		info, err := GetLoad(host)
		if err != nil || !info.Reachable {
			continue
		}

		if best == nil || info.Score < best.Score {
			best = info
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no reachable hosts found")
	}

	return best, nil
}
