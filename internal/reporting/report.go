package reporting

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Crynge/Attributor/internal/attribution"
	"github.com/Crynge/Attributor/internal/mmm"
)

type AttributionReport struct {
	Model   string                      `json:"model"`
	Results []attribution.AttributionResult `json:"results"`
}

type MMMReport struct {
	ROI    []float64       `json:"roi"`
	Params mmm.MMMParams   `json:"params"`
}

func ExportJSON(report interface{}, filePath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if filePath == "" {
		return nil
	}
	return os.WriteFile(filePath, data, 0644)
}

func ExportCSV(results []attribution.AttributionResult, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	_ = w.Write([]string{"channel", "credit", "share"})
	for _, r := range results {
		_ = w.Write([]string{
			r.Channel,
			fmt.Sprintf("%.4f", r.Credit),
			fmt.Sprintf("%.4f", r.Share),
		})
	}
	return nil
}

func SummaryTable(results []attribution.AttributionResult) string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%-20s %10s %10s\n", "Channel", "Credit", "Share"))
	b.WriteString(strings.Repeat("-", 42) + "\n")
	for _, r := range results {
		b.WriteString(fmt.Sprintf("%-20s %10.4f %10.2f%%\n", r.Channel, r.Credit, r.Share*100))
	}
	return b.String()
}
