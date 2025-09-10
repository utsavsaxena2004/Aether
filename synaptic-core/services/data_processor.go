package services

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aether-project/synaptic-core/models"
)

// DataProcessor handles CSV file analysis and processing
type DataProcessor struct {
	redis *RedisService
}

// NewDataProcessor creates a new data processor with Redis
func NewDataProcessor(redis *RedisService) *DataProcessor {
	return &DataProcessor{
		redis: redis,
	}
}

// ProcessCSV analyzes a CSV file and returns a summary
func (dp *DataProcessor) ProcessCSV(file multipart.File, header *multipart.FileHeader, sessionID string) (*models.DataSummary, error) {
	log.Printf("📁 Processing CSV file: %s (%.2f KB)", header.Filename, float64(header.Size)/1024)

	// Read CSV data
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Extract headers
	headers := records[0]
	data := records[1:]

	if len(data) == 0 {
		return nil, fmt.Errorf("CSV file contains only headers")
	}

	// Analyze columns
	columns := make([]models.ColumnInfo, len(headers))
	for i, header := range headers {
		columns[i] = dp.analyzeColumn(header, data, i)
	}

	// Calculate overall statistics
	stats := dp.calculateOverallStats(data, columns)

	// Create summary
	summary := &models.DataSummary{
		FileName:    header.Filename,
		RowCount:    len(data),
		ColumnCount: len(headers),
		Columns:     columns,
		Stats:       stats,
		UploadedAt:  time.Now(),
	}

	// Store in Redis
	if dp.redis != nil {
		if err := dp.redis.SetDataSummary(sessionID, summary); err != nil {
			log.Printf("⚠️ Failed to store data summary in Redis: %v", err)
		}
	}

	log.Printf("✅ CSV processed successfully: %d rows, %d columns", len(data), len(headers))
	return summary, nil
}

// analyzeColumn analyzes a single column and determines its type and statistics
func (dp *DataProcessor) analyzeColumn(name string, data [][]string, colIndex int) models.ColumnInfo {
	if colIndex >= len(data[0]) {
		return models.ColumnInfo{
			Name:     name,
			Type:     "unknown",
			Samples:  []string{},
			NullRate: 1.0,
		}
	}

	// Collect all values for this column
	values := make([]string, 0, len(data))
	nullCount := 0

	for _, row := range data {
		if colIndex < len(row) {
			val := strings.TrimSpace(row[colIndex])
			if val == "" || strings.ToLower(val) == "null" || strings.ToLower(val) == "na" {
				nullCount++
			} else {
				values = append(values, val)
			}
		} else {
			nullCount++
		}
	}

	nullRate := float64(nullCount) / float64(len(data))

	// Determine data type and calculate statistics
	dataType, stats := dp.determineTypeAndStats(values)

	// Get sample values (up to 5)
	samples := dp.getSampleValues(values, 5)

	return models.ColumnInfo{
		Name:     name,
		Type:     dataType,
		Samples:  samples,
		NullRate: nullRate,
		Stats:    stats,
	}
}

// determineTypeAndStats determines the data type and calculates relevant statistics
func (dp *DataProcessor) determineTypeAndStats(values []string) (string, interface{}) {
	if len(values) == 0 {
		return "empty", nil
	}

	// Try to determine if it's numeric
	numericValues := make([]float64, 0, len(values))
	numericCount := 0

	for _, val := range values {
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			numericValues = append(numericValues, num)
			numericCount++
		}
	}

	// If more than 80% of values are numeric, treat as numeric
	if float64(numericCount)/float64(len(values)) > 0.8 {
		return "numeric", dp.calculateNumericStats(numericValues)
	}

	// Try to determine if it's date/time
	dateCount := 0
	for _, val := range values {
		if dp.isDateLike(val) {
			dateCount++
		}
	}

	if float64(dateCount)/float64(len(values)) > 0.6 {
		return "datetime", dp.calculateDateStats(values)
	}

	// Check if it's boolean-like
	if dp.isBooleanLike(values) {
		return "boolean", dp.calculateBooleanStats(values)
	}

	// Default to categorical/text
	return "categorical", dp.calculateCategoricalStats(values)
}

// calculateNumericStats calculates statistics for numeric columns
func (dp *DataProcessor) calculateNumericStats(values []float64) map[string]interface{} {
	if len(values) == 0 {
		return map[string]interface{}{"count": 0}
	}

	// Sort for percentiles
	sort.Float64s(values)

	// Calculate basic stats
	sum := 0.0
	min := values[0]
	max := values[len(values)-1]

	for _, val := range values {
		sum += val
	}

	mean := sum / float64(len(values))

	// Calculate standard deviation
	varSum := 0.0
	for _, val := range values {
		varSum += math.Pow(val-mean, 2)
	}
	stdDev := math.Sqrt(varSum / float64(len(values)))

	// Calculate percentiles
	q25 := dp.percentile(values, 0.25)
	median := dp.percentile(values, 0.5)
	q75 := dp.percentile(values, 0.75)

	return map[string]interface{}{
		"count":  len(values),
		"mean":   mean,
		"std":    stdDev,
		"min":    min,
		"max":    max,
		"q25":    q25,
		"median": median,
		"q75":    q75,
		"unique": dp.countUnique(values),
	}
}

// calculateCategoricalStats calculates statistics for categorical columns
func (dp *DataProcessor) calculateCategoricalStats(values []string) map[string]interface{} {
	// Count occurrences
	counts := make(map[string]int)
	for _, val := range values {
		counts[val]++
	}

	// Find most common values
	type valueCount struct {
		Value string `json:"value"`
		Count int    `json:"count"`
	}

	valueCounts := make([]valueCount, 0, len(counts))
	for val, count := range counts {
		valueCounts = append(valueCounts, valueCount{Value: val, Count: count})
	}

	// Sort by count (descending)
	sort.Slice(valueCounts, func(i, j int) bool {
		return valueCounts[i].Count > valueCounts[j].Count
	})

	// Take top 10
	topValues := valueCounts
	if len(topValues) > 10 {
		topValues = valueCounts[:10]
	}

	return map[string]interface{}{
		"count":       len(values),
		"unique":      len(counts),
		"top_values":  topValues,
		"most_common": valueCounts[0].Value,
	}
}

// calculateBooleanStats calculates statistics for boolean columns
func (dp *DataProcessor) calculateBooleanStats(values []string) map[string]interface{} {
	trueCount := 0
	falseCount := 0

	for _, val := range values {
		val = strings.ToLower(strings.TrimSpace(val))
		if val == "true" || val == "1" || val == "yes" || val == "y" {
			trueCount++
		} else {
			falseCount++
		}
	}

	return map[string]interface{}{
		"count":       len(values),
		"true_count":  trueCount,
		"false_count": falseCount,
		"true_rate":   float64(trueCount) / float64(len(values)),
	}
}

// calculateDateStats calculates statistics for date columns
func (dp *DataProcessor) calculateDateStats(values []string) map[string]interface{} {
	dates := make([]time.Time, 0, len(values))

	for _, val := range values {
		if date := dp.parseDate(val); !date.IsZero() {
			dates = append(dates, date)
		}
	}

	if len(dates) == 0 {
		return map[string]interface{}{"count": 0}
	}

	// Sort dates
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	return map[string]interface{}{
		"count":      len(dates),
		"min_date":   dates[0].Format("2006-01-02"),
		"max_date":   dates[len(dates)-1].Format("2006-01-02"),
		"date_range": dates[len(dates)-1].Sub(dates[0]).Hours() / 24, // days
	}
}

// calculateOverallStats calculates dataset-wide statistics
func (dp *DataProcessor) calculateOverallStats(data [][]string, columns []models.ColumnInfo) map[string]interface{} {
	// Count column types
	typeCounts := make(map[string]int)
	for _, col := range columns {
		typeCounts[col.Type]++
	}

	// Calculate data quality metrics
	totalCells := len(data) * len(columns)
	nullCells := 0
	for _, col := range columns {
		nullCells += int(col.NullRate * float64(len(data)))
	}

	completeness := 1.0 - (float64(nullCells) / float64(totalCells))

	return map[string]interface{}{
		"total_cells":  totalCells,
		"null_cells":   nullCells,
		"completeness": completeness,
		"column_types": typeCounts,
		"memory_usage": dp.estimateMemoryUsage(data),
		"processed_at": time.Now().Format(time.RFC3339),
	}
}

// Helper functions

func (dp *DataProcessor) percentile(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	if len(sortedValues) == 1 {
		return sortedValues[0]
	}

	index := p * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

func (dp *DataProcessor) countUnique(values []float64) int {
	seen := make(map[float64]bool)
	for _, val := range values {
		seen[val] = true
	}
	return len(seen)
}

func (dp *DataProcessor) getSampleValues(values []string, count int) []string {
	if len(values) <= count {
		return values
	}

	samples := make([]string, count)
	step := len(values) / count
	for i := 0; i < count; i++ {
		samples[i] = values[i*step]
	}
	return samples
}

func (dp *DataProcessor) isDateLike(value string) bool {
	return !dp.parseDate(value).IsZero()
}

func (dp *DataProcessor) parseDate(value string) time.Time {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if date, err := time.Parse(format, value); err == nil {
			return date
		}
	}

	return time.Time{}
}

func (dp *DataProcessor) isBooleanLike(values []string) bool {
	boolCount := 0
	for _, val := range values {
		val = strings.ToLower(strings.TrimSpace(val))
		if val == "true" || val == "false" || val == "1" || val == "0" ||
			val == "yes" || val == "no" || val == "y" || val == "n" {
			boolCount++
		}
	}
	return float64(boolCount)/float64(len(values)) > 0.8
}

func (dp *DataProcessor) estimateMemoryUsage(data [][]string) string {
	totalChars := 0
	for _, row := range data {
		for _, cell := range row {
			totalChars += len(cell)
		}
	}

	// Rough estimate: each character takes 1 byte + overhead
	bytes := totalChars * 2 // rough overhead factor

	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}
