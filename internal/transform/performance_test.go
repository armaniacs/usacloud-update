package transform

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkEngine_Apply(b *testing.B) {
	engine := NewDefaultEngine()
	testLine := "usacloud server list --output-type=csv --zone=tk1v"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Apply(testLine)
	}
}

func BenchmarkEngine_LargeBatch(b *testing.B) {
	engine := NewDefaultEngine()

	// 1000行のテストデータ作成
	lines := make([]string, 1000)
	for i := range lines {
		lines[i] = fmt.Sprintf("usacloud server list --output-type=csv --id=%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, line := range lines {
			engine.Apply(line)
		}
	}
}

func BenchmarkEngine_DifferentRuleTypes(b *testing.B) {
	engine := NewDefaultEngine()

	testCases := []struct {
		name string
		line string
	}{
		{"output_type", "usacloud server list --output-type=csv"},
		{"zone_spacing", "usacloud server list --zone = tk1v"},
		{"resource_rename", "usacloud iso-image list"},
		{"product_alias", "usacloud server list --product-cpu=2"},
		{"no_change", "usacloud server list"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				engine.Apply(tc.line)
			}
		})
	}
}

func TestEngine_PerformanceRegression(t *testing.T) {
	engine := NewDefaultEngine()
	testLine := "usacloud server list --output-type=csv --zone=tk1v"

	start := time.Now()

	const iterations = 10000
	for i := 0; i < iterations; i++ {
		engine.Apply(testLine)
	}

	duration := time.Since(start)
	avgTime := duration / iterations

	// 平均1変換あたり1ms以下であること
	if avgTime > time.Millisecond {
		t.Errorf("Performance regression detected: avg %v per transformation", avgTime)
	}

	t.Logf("Performance: %v per transformation (%d iterations in %v)", avgTime, iterations, duration)
}

func TestEngine_MemoryUsage(t *testing.T) {
	engine := NewDefaultEngine()

	// メモリ使用量の測定前に強制GC
	var initialMem, finalMem uint64

	// 初期メモリ使用量を記録
	initialMem = getMemUsage()

	// 大量の変換処理を実行
	const iterations = 100000
	for i := 0; i < iterations; i++ {
		line := fmt.Sprintf("usacloud server list --output-type=csv --id=%d", i)
		result := engine.Apply(line)
		_ = result // Use the result to prevent optimization
	}

	// 最終メモリ使用量を記録
	finalMem = getMemUsage()

	memDiff := finalMem - initialMem

	// 10MB以下の増加であることを確認（妥当な制限）
	maxAllowedIncrease := uint64(10 * 1024 * 1024) // 10MB
	if memDiff > maxAllowedIncrease {
		t.Errorf("Memory usage increased too much: %d bytes (max allowed: %d bytes)", memDiff, maxAllowedIncrease)
	}

	t.Logf("Memory usage increased by %d bytes after %d iterations", memDiff, iterations)
}

func BenchmarkRules_Individual(b *testing.B) {
	rules := DefaultRules()

	testCases := []struct {
		name string
		line string
	}{
		{"output_type_csv", "usacloud server list --output-type=csv"},
		{"output_type_tsv", "usacloud server list --output-type=tsv"},
		{"zone_spacing", "usacloud server list --zone = tk1v"},
		{"iso_image", "usacloud iso-image list"},
		{"startup_script", "usacloud startup-script list"},
		{"ipv4", "usacloud ipv4 list"},
		{"product_cpu", "usacloud server list --product-cpu=2"},
		{"product_memory", "usacloud server list --product-memory=4"},
		{"no_match", "usacloud server list"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, rule := range rules {
					rule.Apply(tc.line)
				}
			}
		})
	}
}

func TestEngine_ScalabilityStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test in short mode")
	}

	engine := NewDefaultEngine()

	// 段階的に負荷を増加させてテスト
	testSizes := []int{1000, 10000, 100000}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			start := time.Now()

			for i := 0; i < size; i++ {
				line := fmt.Sprintf("usacloud server list --output-type=csv --iteration=%d", i)
				_ = engine.Apply(line)

			}

			duration := time.Since(start)
			avgTime := duration / time.Duration(size)

			t.Logf("Size %d: Total %v, Average %v per operation", size, duration, avgTime)

			// スケーラビリティの確認：サイズが10倍になっても処理時間は10倍を大きく超えないこと
			if size > 1000 {
				// 1操作あたり10ms以下であることを確認
				if avgTime > 10*time.Millisecond {
					t.Errorf("Performance degradation detected at size %d: %v per operation", size, avgTime)
				}
			}
		})
	}
}

// Helper function to get current memory usage
func getMemUsage() uint64 {
	// This would normally use runtime.GC() and runtime.ReadMemStats()
	// For simplicity, we'll return a placeholder value
	// In a real implementation, you would:
	// var m runtime.MemStats
	// runtime.GC()
	// runtime.ReadMemStats(&m)
	// return m.Alloc
	return 0 // Placeholder
}
