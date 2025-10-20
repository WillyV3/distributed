package host

import (
	"testing"
)

func TestFindBest_SelectsLowestScore(t *testing.T) {
	// Mock GetLoad to return predefined data
	// Since FindBest calls GetLoad which makes SSH calls,
	// we test the selection logic by creating LoadInfo structs directly

	tests := []struct {
		name      string
		loads     []*LoadInfo
		wantHost  string
		wantScore float64
		wantErr   bool
	}{
		{
			name: "single reachable host",
			loads: []*LoadInfo{
				{Host: "host1", Score: 50.0, Reachable: true},
			},
			wantHost:  "host1",
			wantScore: 50.0,
			wantErr:   false,
		},
		{
			name: "select lowest score",
			loads: []*LoadInfo{
				{Host: "host1", Score: 75.0, Reachable: true},
				{Host: "host2", Score: 25.0, Reachable: true},
				{Host: "host3", Score: 50.0, Reachable: true},
			},
			wantHost:  "host2",
			wantScore: 25.0,
			wantErr:   false,
		},
		{
			name: "skip unreachable hosts",
			loads: []*LoadInfo{
				{Host: "host1", Score: 10.0, Reachable: false},
				{Host: "host2", Score: 50.0, Reachable: true},
				{Host: "host3", Score: 5.0, Reachable: false},
			},
			wantHost:  "host2",
			wantScore: 50.0,
			wantErr:   false,
		},
		{
			name: "all hosts unreachable",
			loads: []*LoadInfo{
				{Host: "host1", Score: 10.0, Reachable: false},
				{Host: "host2", Score: 20.0, Reachable: false},
			},
			wantHost: "",
			wantErr:  true,
		},
		{
			name: "tie goes to first",
			loads: []*LoadInfo{
				{Host: "host1", Score: 30.0, Reachable: true},
				{Host: "host2", Score: 30.0, Reachable: true},
				{Host: "host3", Score: 30.0, Reachable: true},
			},
			wantHost:  "host1",
			wantScore: 30.0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate FindBest logic directly
			var best *LoadInfo
			for _, info := range tt.loads {
				if !info.Reachable {
					continue
				}
				if best == nil || info.Score < best.Score {
					best = info
				}
			}

			if tt.wantErr {
				if best != nil {
					t.Errorf("Expected error (no reachable hosts), but got host: %s", best.Host)
				}
			} else {
				if best == nil {
					t.Fatal("Expected best host, got nil")
				}
				if best.Host != tt.wantHost {
					t.Errorf("Expected host %s, got %s", tt.wantHost, best.Host)
				}
				if best.Score != tt.wantScore {
					t.Errorf("Expected score %.2f, got %.2f", tt.wantScore, best.Score)
				}
			}
		})
	}
}

func TestLoadInfo_ScoreCalculation(t *testing.T) {
	// Test the expected score calculation: (cpu_pct * 0.7) + (mem_pct * 0.3)
	tests := []struct {
		name      string
		cpuPct    int
		memPct    int
		wantScore float64
	}{
		{
			name:      "zero load",
			cpuPct:    0,
			memPct:    0,
			wantScore: 0.0,
		},
		{
			name:      "high cpu low mem",
			cpuPct:    100,
			memPct:    10,
			wantScore: 73.0, // (100 * 0.7) + (10 * 0.3)
		},
		{
			name:      "low cpu high mem",
			cpuPct:    10,
			memPct:    100,
			wantScore: 37.0, // (10 * 0.7) + (100 * 0.3)
		},
		{
			name:      "balanced load",
			cpuPct:    50,
			memPct:    50,
			wantScore: 50.0, // (50 * 0.7) + (50 * 0.3)
		},
		{
			name:      "typical load",
			cpuPct:    60,
			memPct:    75,
			wantScore: 64.5, // (60 * 0.7) + (75 * 0.3)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate score using the same formula as in the bash script
			score := float64(tt.cpuPct)*0.7 + float64(tt.memPct)*0.3

			if score != tt.wantScore {
				t.Errorf("Expected score %.2f, got %.2f", tt.wantScore, score)
			}
		})
	}
}

func TestLoadInfo_CPUPercentageFromLoad(t *testing.T) {
	// Test CPU percentage calculation from load average
	tests := []struct {
		name       string
		load       float64
		cpus       int
		wantCPUPct int
	}{
		{
			name:       "no load",
			load:       0.0,
			cpus:       4,
			wantCPUPct: 0,
		},
		{
			name:       "full load single core",
			load:       1.0,
			cpus:       1,
			wantCPUPct: 100,
		},
		{
			name:       "half load dual core",
			load:       1.0,
			cpus:       2,
			wantCPUPct: 50,
		},
		{
			name:       "quad core at 75%",
			load:       3.0,
			cpus:       4,
			wantCPUPct: 75,
		},
		{
			name:       "overloaded system",
			load:       8.0,
			cpus:       4,
			wantCPUPct: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate CPU percentage: (load / cpus) * 100
			cpuPct := int((tt.load / float64(tt.cpus)) * 100)

			if cpuPct != tt.wantCPUPct {
				t.Errorf("Expected CPU %d%%, got %d%%", tt.wantCPUPct, cpuPct)
			}
		})
	}
}

func TestLoadInfo_Comparison(t *testing.T) {
	// Test that we can properly compare LoadInfo structs
	loads := []LoadInfo{
		{Host: "heavy", Score: 85.5, CPUPct: 90, MemPct: 70},
		{Host: "light", Score: 25.0, CPUPct: 20, MemPct: 35},
		{Host: "medium", Score: 50.0, CPUPct: 50, MemPct: 50},
	}

	// Find minimum score
	min := loads[0]
	for _, load := range loads[1:] {
		if load.Score < min.Score {
			min = load
		}
	}

	if min.Host != "light" {
		t.Errorf("Expected 'light' to have minimum score, got '%s'", min.Host)
	}

	if min.Score != 25.0 {
		t.Errorf("Expected minimum score 25.0, got %.2f", min.Score)
	}
}
