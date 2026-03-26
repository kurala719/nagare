package utils

import (
	"bytes"
	"sort"

	"github.com/wcharczuk/go-chart/v2"
)

// GeneratePieChart generates a pie chart and returns the PNG bytes
func GeneratePieChart(title string, values map[string]float64, noDataText string) ([]byte, error) {
	if len(values) == 0 {
		if noDataText == "" {
			noDataText = "No Data"
		}
		values = map[string]float64{noDataText: 1}
	}
	keys := make([]string, 0, len(values))
	for label := range values {
		keys = append(keys, label)
	}
	sort.Strings(keys)

	var slices []chart.Value
	for i, label := range keys {
		val := values[label]
		c := chart.GetDefaultColor(i + 1).WithAlpha(220)
		slices = append(slices, chart.Value{
			Value: val,
			Label: label,
			Style: chart.Style{
				FillColor:   c,
				StrokeColor: c,
			},
		})
	}

	pie := chart.PieChart{
		Title:  title,
		Width:  460,
		Height: 360,
		Values: slices,
	}

	buffer := bytes.NewBuffer([]byte{})
	err := pie.Render(chart.PNG, buffer)
	return buffer.Bytes(), err
}

// GenerateLineChart generates a line chart for trends
func GenerateLineChart(title string, xValues []string, yValues []float64, xLabel, yLabel string) ([]byte, error) {
	if len(yValues) == 0 {
		yValues = []float64{0}
	}
	if len(xValues) == 0 {
		xValues = []string{"1"}
	}
	points := len(yValues)
	if len(xValues) < points {
		points = len(xValues)
	}
	if points == 0 {
		points = 1
		yValues = []float64{0}
		xValues = []string{"1"}
	}

	var x []float64
	var y []float64
	ticks := make([]chart.Tick, 0, points)
	for i := 0; i < points; i++ {
		val := yValues[i]
		x = append(x, float64(i))
		y = append(y, val)
		ticks = append(ticks, chart.Tick{Value: float64(i), Label: xValues[i]})
	}

	graph := chart.Chart{
		Title:  title,
		Width:  520,
		Height: 320,
		XAxis: chart.XAxis{
			Name:  xLabel,
			Ticks: ticks,
		},
		YAxis: chart.YAxis{
			Name: yLabel,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: chart.GetDefaultColor(1).WithAlpha(255),
					StrokeWidth: 2,
					FillColor:   chart.GetDefaultColor(1).WithAlpha(35),
				},
				XValues: x,
				YValues: y,
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	return buffer.Bytes(), err
}

// GenerateBarChart generates a bar chart
func GenerateBarChart(title string, values map[string]float64, noDataText string) ([]byte, error) {
	if len(values) == 0 {
		if noDataText == "" {
			noDataText = "No Data"
		}
		values = map[string]float64{noDataText: 0}
	}
	keys := make([]string, 0, len(values))
	for label := range values {
		keys = append(keys, label)
	}
	sort.Slice(keys, func(i, j int) bool {
		if values[keys[i]] == values[keys[j]] {
			return keys[i] < keys[j]
		}
		return values[keys[i]] > values[keys[j]]
	})

	var bars []chart.Value
	for i, label := range keys {
		val := values[label]
		c := chart.GetDefaultColor(i + 1).WithAlpha(220)
		bars = append(bars, chart.Value{
			Value: val,
			Label: label,
			Style: chart.Style{
				FillColor:   c,
				StrokeColor: c,
			},
		})
	}

	bc := chart.BarChart{
		Title: title,
		Width: 560,
		Background: chart.Style{
			Padding: chart.Box{
				Top:   40,
				Left:  20,
				Right: 20,
			},
		},
		Height:   320,
		BarWidth: 42,
		Bars:     bars,
	}

	buffer := bytes.NewBuffer([]byte{})
	err := bc.Render(chart.PNG, buffer)
	return buffer.Bytes(), err
}
