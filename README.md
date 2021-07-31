# yogo
go develop tools

### usages
```
	timer := yogo.DailyRangeIntervalTimer{
		Name:       "DailyRangeTimer",
		Interval:   yogo.GetSecond(1),
		Start:      "18:55:00",
		End:        "20:00:00",
		Fn: func() {
			yogo.PrintSomething("running...")
		},
	}
```