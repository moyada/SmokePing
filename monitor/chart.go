package monitor

import (
	"fmt"
	"github.com/wcharczuk/go-chart/v2"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	height, width                    = 1080, 1980
	zeroY, timeoutY, labelY, healthY = 0, 1000, 1020, 200
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

	delayStyle = chart.Style{
		StrokeColor: chart.ColorBlue,
		StrokeWidth: 3,
		FillColor:   chart.ColorBlue.WithAlpha(100),
	}

	timeoutStyle = chart.Style{
		StrokeColor: chart.ColorRed,
		StrokeWidth: 3,
		FillColor:   chart.ColorRed.WithAlpha(100),
	}

	labelStyle = chart.Style{
		FontSize: 15,
		Padding: chart.Box{
			Top: -10,
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
		{Value: 100, Label: "100 ms"},
		{Value: 200, Label: "200 ms"},
		{Value: 300, Label: "300 ms"},
		{Value: 400, Label: "400 ms"},
		{Value: 500, Label: "500 ms"},
		{Value: 600, Label: "600 ms"},
		{Value: 700, Label: "700 ms"},
		{Value: 800, Label: "800 ms"},
		{Value: 900, Label: "900 ms"},
		{Value: 1000, Label: "Timeout"},
		{Value: 1200, Label: ""},
	}

	reportStyle = chart.Style{
		StrokeWidth: 0,
		DotWidth:    0,
		FontSize:    13,
		StrokeColor: chart.ColorWhite,
		Padding: chart.Box{
			Left: -10,
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

func toFileName(output, host string, start *time.Time, du *time.Duration) string {
	if output != "" && strings.LastIndexAny(output, "/") != len(output)-1 {
		output += "/"
	}
	s1 := start.Format("2006-01-02 15:04:05")
	s2 := start.Add(*du).Format("15:04:05")
	return fmt.Sprintf("%v%v %v~%v.png", output, host, s1, s2)
}

type Chart struct {
}

func (c *Chart) record(seq int, time *time.Time, delay *time.Duration) {
}

func (c *Chart) output(host string, startTime *time.Time, records *map[int]*time.Duration, report *Report, output string) error {
	delayXValues := make([]float64, 0)
	delayYValues := make([]float64, 0)

	timeoutXValues := make([]float64, 0)
	timeoutYValues := make([]float64, 0)

	HealthXValues := make([]float64, 0)
	HealthYValues := make([]float64, 0)

	labels := make([]chart.Value2, 0)

	// 排序
	var keys []int
	for key := range *records {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	var lastTimeOut = -10

	for i, key := range keys {
		timeout, exist := (*records)[key]

		t := startTime.Add(time.Second * time.Duration(key))
		x := getTimeNano(&t)

		if exist && timeout != nil {
			y := float64(timeout.Milliseconds())

			// 上个请求超时
			if i > -1 && lastTimeOut == i-1 {
				td := t.Add(-50 * time.Millisecond)
				xd := getTimeNano(&td)

				// 延长耗时折线
				delayXValues = append(delayXValues, xd)
				delayYValues = append(delayYValues, zeroY)

				// 上升耗时折线
				delayXValues = append(delayXValues, xd)
				delayYValues = append(delayYValues, y)

				tt := t.Add(-100 * time.Millisecond)
				xt := getTimeNano(&tt)

				// 延长超时折线
				timeoutXValues = append(timeoutXValues, xt)
				timeoutYValues = append(timeoutYValues, timeoutY)

				// 下降超时折线
				timeoutXValues = append(timeoutXValues, xt)
				timeoutYValues = append(timeoutYValues, zeroY)
			} else {
				delayXValues = append(delayXValues, x)
				delayYValues = append(delayYValues, y)

				timeoutXValues = append(timeoutXValues, x)
				timeoutYValues = append(timeoutYValues, zeroY)
			}
		} else {
			// 请求超时
			var xd, xt float64
			if i > -1 && lastTimeOut != i-1 {
				// 上个请求未超时

				// 延长耗时折线
				dt := t.Add(-25 * time.Millisecond)
				xd = getTimeNano(&dt)
				yl := delayYValues[len(delayYValues)-1]

				delayXValues = append(delayXValues, xd)
				delayYValues = append(delayYValues, yl)

				// 延长超时折线
				tt := t.Add(25 * time.Millisecond)
				xt = getTimeNano(&tt)

				timeoutXValues = append(timeoutXValues, xt)
				timeoutYValues = append(timeoutYValues, zeroY)
			} else {
				xd = x
				xt = x
			}
			delayXValues = append(delayXValues, xd)
			delayYValues = append(delayYValues, zeroY)

			timeoutXValues = append(timeoutXValues, xt)
			timeoutYValues = append(timeoutYValues, timeoutY)

			// 上个请求未超时，添加超时标签
			if lastTimeOut != i-1 {
				label := chart.Value2{
					XValue: x,
					YValue: labelY,
					Label:  timeoutLabel,
					Style:  annoStyle,
				}
				labels = append(labels, label)
			}

			lastTimeOut = i
		}

		HealthXValues = append(HealthXValues, x)
		HealthYValues = append(HealthYValues, healthY)
	}

	// 计算时间轴区间，5分之内每步30秒，5分以上每步1分钟
	step, half := 30, 10
	if len(*records) > 300 {
		step = 60
		half = 30
	}

	sec := startTime.Second()
	ms := step - sec%step
	if ms < half {
		ms += step
	}
	m, _ := time.ParseDuration(fmt.Sprintf("%vs", ms))

	// 描绘x轴
	endTime := time.Now()
	ticks := make([]chart.Tick, 0)

	// 起点
	ticks = append(ticks, chart.Tick{Value: float64(startTime.UnixNano()), Label: timeFormat(startTime)})

	// 整数时间点
	st := startTime.Add(m)

	for i := 0; ; i++ {
		t := st.Add(time.Second * time.Duration(step*i))
		v := float64(t.UnixNano())
		if t.After(endTime) {
			// 终点
			tick := chart.Tick{Value: v, Label: timeFormat(&endTime)}
			if endTime.Second()%step < 15 {
				ticks[len(ticks)-1] = tick
			} else {
				ticks = append(ticks, tick)
			}
			break
		}
		tick := chart.Tick{Value: v, Label: timeFormat(&t)}
		ticks = append(ticks, tick)
	}

	// 耗时折线
	delayLine := chart.ContinuousSeries{
		Name:    "Delay",
		Style:   delayStyle,
		XValues: delayXValues,
		YValues: delayYValues,
	}

	// 超时折线
	timeoutLine := chart.ContinuousSeries{
		Name:    "Timeout",
		Style:   timeoutStyle,
		XValues: timeoutXValues,
		YValues: timeoutYValues,
	}
	// 安全耗时
	safeLine := &chart.ContinuousSeries{
		Name:    "Safety (200 ms)",
		Style:   safetyStyle,
		XValues: HealthXValues,
		YValues: HealthYValues,
	}

	// 综合报告
	x := delayXValues[0] - 200000
	labels = append(labels, chart.Value2{
		XValue: x,
		YValue: 1180,
		Label:  fmt.Sprintf("AvgRtt:  %v ms", report.AvgRtt.Milliseconds()),
		Style:  reportStyle,
	})
	labels = append(labels, chart.Value2{
		XValue: x,
		YValue: 1130,
		Label:  fmt.Sprintf("MaxRtt:  %v ms", report.MaxRtt.Milliseconds()),
		Style:  reportStyle,
	})
	labels = append(labels, chart.Value2{
		XValue: x,
		YValue: 1080,
		Label:  fmt.Sprintf("MinRtt:  %v ms", report.MinRtt.Milliseconds()),
		Style:  reportStyle,
	})

	graph := chart.Chart{
		//Title: host,
		Height:     height,
		Width:      width,
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
			delayLine,
			timeoutLine,
			safeLine,
			//chart.LastValueAnnotationSeries(safeLine),
			chart.AnnotationSeries{Style: labelStyle, Annotations: labels},
		},
	}

	//for i, value := range delayXValues {
	//	fmt.Printf("%v %v\n", value, delayYValues[i])
	//}
	//fmt.Println("-------")
	//for i, value := range timeoutXValues {
	//	fmt.Printf("%v %v\n", value, timeoutYValues[i])
	//}
	//fmt.Println("-------")
	//for i, value := range HealthXValues {
	//	fmt.Printf("%v %v\n", value, HealthYValues[i])
	//}

	graph.Elements = []chart.Renderable{chart.LegendThin(&graph, eleStyle)}

	// 监听时间
	var du = time.Duration(len(keys)-1) * time.Second

	fileName := toFileName(output, host, startTime, &du)
	f, _ := os.Create(fileName)
	defer f.Close()
	return graph.Render(chart.PNG, f)
}
