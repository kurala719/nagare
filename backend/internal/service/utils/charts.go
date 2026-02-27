package utils

import (
	"bytes"

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
	var slices []chart.Value
	for label, val := range values {
		slices = append(slices, chart.Value{Value: val, Label: label})
	}

	pie := chart.PieChart{
		Width:  512,
		Height: 512,
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
	var x []float64
	var y []float64
	for i, val := range yValues {
		x = append(x, float64(i))
		y = append(y, val)
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name: xLabel,
		},
		YAxis: chart.YAxis{
			Name: yLabel,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: chart.GetDefaultColor(0).WithAlpha(255),
					FillColor:   chart.GetDefaultColor(0).WithAlpha(100),
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
	var bars []chart.Value
	for label, val := range values {
		bars = append(bars, chart.Value{Value: val, Label: label})
	}

	bc := chart.BarChart{
		Title: title,
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   512,
		BarWidth: 60,
		Bars:     bars,
	}

	buffer := bytes.NewBuffer([]byte{})
	err := bc.Render(chart.PNG, buffer)
	return buffer.Bytes(), err
}
