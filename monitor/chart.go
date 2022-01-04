package monitor

import (
	"fmt"
	"github.com/wcharczuk/go-chart/v2"
	"os"
	"sort"
	"time"
)

const (
	height, width                    = 1080, 1800
	zeroY, timeoutY, healthY, labelY = 0, 1100, 200, 1020
)

var (
	timeoutLabel = "timeout"

	chartStyle = chart.Style{
		Padding: chart.Box{
			Left:   50,
			Top:    50,
			Bottom: 50,
			Right:  50,
		},
	}

	eleStyle = chart.Style{
		FontSize:    15,
		StrokeWidth: 15,
	}

	latencyStyle = chart.Style{
		StrokeColor: chart.ColorBlue,
		StrokeWidth: 2,
		FillColor:   chart.ColorBlue.WithAlpha(100),
	}

	timeoutStyle = chart.Style{
		StrokeColor: chart.ColorRed,
		StrokeWidth: 2,
		FillColor:   chart.ColorRed.WithAlpha(100),
	}

	labelStyle = chart.Style{
		FontSize: 15,
		Padding: chart.Box{
			Left: 5,
		},
	}

	safetyStyle = chart.Style{
		StrokeWidth:     3,
		StrokeColor:     chart.ColorAlternateGreen,
		StrokeDashArray: []float64{5.0, 5.0},
	}

	annoStyle = chart.Style{
		FontSize:    12,
		StrokeColor: chart.ColorRed,
	}

	ytick = []chart.Tick{
		{Value: 0, Label: "0 ms"},
		{Value: 50, Label: "50 ms"},
		{Value: 100, Label: "100 ms"},
		{Value: 200, Label: "200 ms"},
		{Value: 300, Label: "300 ms"},
		{Value: 400, Label: "400 ms"},
		{Value: 500, Label: "500 ms"},
		{Value: 600, Label: "600 ms"},
		{Value: 700, Label: "700 ms"},
		{Value: 800, Label: "800 ms"},
		{Value: 900, Label: "900 ms"},
		{Value: 1000, Label: "1000 ms"},
		{Value: 1100, Label: "Timeout"},
		{Value: 1200, Label: ""},
	}

	reportStyle = chart.Style{
		StrokeWidth: 0,
		DotWidth:    0,
		FontSize:    13,
		StrokeColor: chart.ColorWhite,
		Padding: chart.Box{
			Left: -5,
		},
	}
)

func timeFormat(v interface{}) string {
	t := v.(*time.Time)
	return t.Format("15:04:05")
}

func getTimeNano(t *time.Time) float64 {
	return float64(t.UnixNano())
}

func calculateDraw(startTime *time.Time, count int) (int, int, bool) {
	// 计算时间轴区间
	var step, min int
	timeLabel := false

	switch {
	case count > 43200: // 12小时 1小时
		step = 3600
		min = 60 - startTime.Minute()
		if min <= 20 {
			min += 60
		}
		timeLabel = true
	case 43200 >= count && count > 10800: // 3小时 30分钟
		step = 1800
		min = 30 - startTime.Minute()
		if min <= 15 {
			min += 30
		}
		timeLabel = true
	case 10800 >= count && count > 3600: // 1小时 10分钟
		step = 600
		min = 10 - startTime.Minute()%10
		if min <= 5 {
			min += 10
		}
	case 3600 >= count && count > 1800: // 半小时 5分钟
		step = 300
		min = 5 - startTime.Minute()%5
		if min == 1 {
			min += 5
		}
	case 1800 >= count && count > 600: // 10分钟 2分钟
		step = 120
		min = 2 - startTime.Minute()%2
	case 600 >= count && count > 300: // 5分钟 1分钟
		step = 60
		min = 1
	default:
		step = 30
		min = 0
	}

	return step, min, timeLabel
}

// 描绘x轴坐标
func xAxis(startTime, endTime *time.Time, step, min int) []chart.Tick {
	ticks := make([]chart.Tick, 0)

	// 起点
	ticks = append(ticks, chart.Tick{Value: getTimeNano(startTime), Label: timeFormat(startTime)})

	// 整数时间点
	var nextTime time.Time

	sec := startTime.Second()
	if min > 0 {
		nextTime = startTime.Add(time.Duration(60-sec) * time.Second)
		nextTime = nextTime.Add(time.Duration(min-1) * time.Minute)
	} else if sec%30 < 15 {
		nextTime = startTime.Add(time.Duration(30-sec%30) * time.Second)
	} else {
		nextTime = *startTime
	}

	for i := 0; ; i++ {
		nt := nextTime.Add(time.Second * time.Duration(step*i))
		if nt.After(*endTime) {
			// 终点
			tick := chart.Tick{Value: getTimeNano(endTime), Label: timeFormat(endTime)}

			if len(ticks) < 2 {
				ticks = append(ticks, tick)
				break
			}

			switch {
			case step >= 600:
				ticks[i] = tick
			case min > 0 && (endTime.Add(time.Duration(-step) * time.Minute).Before(nt)):
				ticks[i] = tick
			case min == 0 && (endTime.Add(-time.Minute).Before(nt)):
				ticks[i] = tick
			default:
				ticks = append(ticks, tick)
			}

			break
		}
		tick := chart.Tick{Value: getTimeNano(&nt), Label: timeFormat(&nt)}
		ticks = append(ticks, tick)
	}
	return ticks
}

type LineData struct {
	xLine, yLine []float64
}

func (p *LineData) addPoint(x, y float64) {
	p.xLine = append(p.xLine, x)
	p.yLine = append(p.yLine, y)
}

func newLineData() *LineData {
	return &LineData{
		xLine: make([]float64, 0),
		yLine: make([]float64, 0),
	}
}

type Chart struct {
}

func (c *Chart) record(seq int, latency *time.Duration) {
}

func (c *Chart) output(output string, startTime *time.Time, records map[int]*time.Duration, report *Report) error {
	latencyData := newLineData()
	timeoutData := newLineData()
	healthData := newLineData()

	labels := make([]chart.Value2, 0)

	total := len(records)
	lastIndex := total - 1

	// 计算时间轴区间
	step, min, _ := calculateDraw(startTime, total)

	// 排序
	var keys []int
	for key := range records {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	// 监听时间
	var duration = time.Duration(lastIndex) * time.Second

	lastTimeOut := -10

	for i, key := range keys {
		timeout, exist := (records)[key]
		// 是否最后一个节点
		lastItem := lastIndex == i

		t := startTime.Add(time.Second * time.Duration(key))
		x := getTimeNano(&t)

		healthData.addPoint(x, healthY)

		if exist && timeout != nil {
			y := float64(timeout.Milliseconds())

			// 上个请求超时
			if i > 0 && lastTimeOut == i-1 {
				var td time.Time
				if lastItem {
					td = t.Add(-40 * time.Millisecond)
				} else {
					td = t.Add(-20 * time.Millisecond)
				}
				xd := getTimeNano(&td)

				// 延长耗时折线
				latencyData.addPoint(xd, zeroY)
				// 上升耗时折线
				latencyData.addPoint(xd, y)

				if lastItem {
					// 延长耗时折线
					latencyData.addPoint(x, y)
				}

				tt := t.Add(-40 * time.Millisecond)
				xt := getTimeNano(&tt)

				// 延长超时折线
				timeoutData.addPoint(xt, timeoutY)
				// 下降超时折线
				timeoutData.addPoint(xt, zeroY)
			} else {
				latencyData.addPoint(x, y)
				timeoutData.addPoint(x, zeroY)
			}
		} else {
			// 请求超时
			var xd, xt float64
			if i > 0 && lastTimeOut != i-1 {
				// 上个请求未超时

				// 延长耗时折线
				var dt time.Time
				if lastItem {
					dt = t.Add(-40 * time.Millisecond)
				} else {
					dt = t.Add(-30 * time.Millisecond)
				}
				xd = getTimeNano(&dt)
				yl := latencyData.yLine[len(latencyData.yLine)-1]

				latencyData.addPoint(xd, yl)

				// 延长超时折线
				if lastItem {
					tt := t.Add(-20 * time.Millisecond)
					xt = getTimeNano(&tt)

					timeoutData.addPoint(xt, zeroY)
					timeoutData.addPoint(xt, timeoutY)

					xt = x
				} else {
					tt := t.Add(-10 * time.Millisecond)
					xt = getTimeNano(&tt)
					timeoutData.addPoint(xt, timeoutY)

					xt = x
				}

			} else {
				xd = x
				xt = x
			}
			latencyData.addPoint(xd, zeroY)
			timeoutData.addPoint(xt, timeoutY)

			// 上个请求未超时，添加超时标签
			//if lastTimeOut != i-1 && !lastItem {
			//	name := timeoutLabel
			//	// 3小时以上标签更换为时间点
			//	if timeLabel {
			//		name = timeFormat(&t)
			//	}
			//
			//	label := chart.Value2{
			//		XValue: x,
			//		YValue: labelY,
			//		Label:  name,
			//		Style:  annoStyle,
			//	}
			//	labels = append(labels, label)
			//}

			lastTimeOut = i
		}
	}

	// 描绘x轴
	endTime := startTime.Add(duration)
	ticks := xAxis(startTime, &endTime, step, min)

	// 耗时折线
	latencyLine := chart.ContinuousSeries{
		Name:    "Latency",
		Style:   latencyStyle,
		XValues: latencyData.xLine,
		YValues: latencyData.yLine,
	}

	// 超时折线
	timeoutLine := chart.ContinuousSeries{
		Name:    "Timeout",
		Style:   timeoutStyle,
		XValues: timeoutData.xLine,
		YValues: timeoutData.yLine,
	}
	// 安全耗时
	safeLine := &chart.ContinuousSeries{
		Name:    fmt.Sprintf("Safety (%v ms)", healthY),
		Style:   safetyStyle,
		XValues: healthData.xLine,
		YValues: healthData.yLine,
	}

	// 综合报告
	x := latencyData.xLine[0] - 200000
	labels = append(labels, chart.Value2{
		XValue: x,
		YValue: 1180,
		Label:  fmt.Sprintf("AvgRtt: %v ms", report.AvgRtt.Milliseconds()),
		Style:  reportStyle,
	})
	labels = append(labels, chart.Value2{
		XValue: x,
		YValue: 1135,
		Label:  fmt.Sprintf("MaxRtt: %v ms", report.MaxRtt.Milliseconds()),
		Style:  reportStyle,
	})
	labels = append(labels, chart.Value2{
		XValue: x,
		YValue: 1090,
		Label:  fmt.Sprintf("MinRtt: %v ms", report.MinRtt.Milliseconds()),
		Style:  reportStyle,
	})
	labels = append(labels, chart.Value2{
		XValue: x,
		YValue: 1045,
		Label:  fmt.Sprintf("Loss:  %v", fmt.Sprintf("%.2f", report.Loss)) + " %",
		Style:  reportStyle,
	})

	w := width
	if total > 1800 {
		w = total * 1
	}

	graph := chart.Chart{
		//Title: host,
		Height:     height,
		Width:      w,
		Background: chartStyle,
		XAxis: chart.XAxis{
			//ValueFormatter: timeFormat,
			Ticks:     ticks,
			TickStyle: labelStyle,
		},
		YAxis: chart.YAxis{
			Name:      "Elapsed Millis",
			NameStyle: labelStyle,
			Ticks:     ytick,
			TickStyle: labelStyle,
		},
		Series: []chart.Series{
			latencyLine,
			timeoutLine,
			safeLine,
			//chart.LastValueAnnotationSeries(safeLine),
			chart.AnnotationSeries{Style: labelStyle, Annotations: labels},
		},
	}

	graph.Elements = []chart.Renderable{chart.LegendThin(&graph, eleStyle)}

	f, _ := os.Create(output)
	defer f.Close()
	return graph.Render(chart.PNG, f)
}
