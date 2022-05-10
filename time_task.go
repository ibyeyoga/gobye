package yogo

import (
	"fmt"
	"log"
	"time"
)

const ONE_DAY = 24 * time.Hour

//每日定时任务定时器
type DailyTimer struct {
	Name   string
	Hour   int
	Minute int
	Second int
	tr     *time.Timer
	Fn     func()
}

//每日时间范围间隔定时器
type DailyRangeIntervalTimer struct {
	Name             string
	Interval         time.Duration
	Start            string
	End              string
	startTr          *time.Timer
	endTr            *time.Timer
	intervalTk       *time.Ticker
	startTimeChan    chan time.Time
	endTimeChan      chan time.Time
	Fn               func(timer *DailyRangeIntervalTimer)
	FnExecLimit      int
	fnTodayExecCount int
}

func GetHour(hour int) time.Duration {
	return time.Duration(hour) * time.Hour
}

func GetMinute(minute int) time.Duration {
	return time.Duration(minute) * time.Minute
}

func GetSecond(second int) time.Duration {
	return time.Duration(second) * time.Second
}

func PrintSomething(extra string) {
	fmt.Println("Hello world!" + extra)
}

func (timer DailyTimer) RunTask() {
	timer.tr = time.NewTimer(timer.getNextTickDuration())
	go func() {
		for {
			timer.WaitAndFlush()
			timer.Fn()
		}
	}()
}

/**
获取下一次启动闹钟的时间
*/
func (timer DailyTimer) getNextTickDuration() time.Duration {
	now, nextTick := getTodayNowAndSpecialTime(timer.Hour, timer.Minute, timer.Second)
	if nextTick.Before(now) {
		nextTick = nextTick.Add(ONE_DAY)
	}
	return nextTick.Sub(now)
}

/**
获得今天时间和指定时间
*/
func getTodayNowAndSpecialTime(Hour int, Minute int, Second int) (time.Time, time.Time) {
	//获取当前时间
	now := time.Now()
	specialTime := time.Date(now.Year(), now.Month(), now.Day(), Hour, Minute, Second, 0, time.Local)
	return now, specialTime
}

/**
通过字符串获取时间
*/
func GetTimeByString(timeStr string) (time.Time) {
	strTime, _ := time.Parse("15:04:05", timeStr)

	hour := strTime.Hour()
	minute := strTime.Minute()
	second := strTime.Second()
	spTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(),
		hour, minute, second, 0, time.Local);
	return spTime
}

/**
获取今天时间和指定时间
*/
func getTodayNowAndSpecialDuration(Hour int, Minute int, Second int) time.Duration {
	now, specialTime := getTodayNowAndSpecialTime(Hour, Minute, Second)
	if specialTime.Before(now) {
		specialTime = specialTime.Add(time.Hour * 24)
	}
	return specialTime.Sub(now)
}

/**
等待并刷新定时器
*/
func (timer DailyTimer) WaitAndFlush() {
	<-timer.tr.C
	timer.tr.Reset(timer.getNextTickDuration())
}

/**
执行定时器任务
*/
func (timer DailyRangeIntervalTimer) RunTask() {
	startTime := GetTimeByString(timer.Start)
	endTime := GetTimeByString(timer.End)
	now := time.Now()

	if startTime.Before(now) && endTime.After(now) {
		//时间范围内
		timer.initEndTr(endTime.Sub(now))
	} else {
		for ; startTime.Before(now); {
			startTime = startTime.Add(ONE_DAY)
		}
	}
	timer.initStartTr(startTime.Sub(now))

	timer.startTimeChan = make(chan time.Time, 0)
	timer.endTimeChan = make(chan time.Time, 0)
	go func() {
		for {
			select {
			case n := <-timer.startTimeChan:
				//处理开始逻辑
				endTime := GetTimeByString(timer.End)
				for ; endTime.Before(n); {
					endTime = endTime.Add(ONE_DAY)
				}

				timer.startTicker()
				timer.initEndTr(endTime.Sub(n))
			case n := <-timer.endTimeChan:
				//处理结束逻辑
				if timer.intervalTk != nil {
					//关闭ticker
					timer.intervalTk.Stop()
				}

				log.Println("已结束循环执行任务，时间:" + n.String())

				//设置开始定时器
				startTime := GetTimeByString(timer.Start)
				for ; startTime.Before(n); {
					startTime = startTime.Add(ONE_DAY)
				}

				timer.initStartTr(startTime.Sub(n))
			}
		}
	}()
}

/*
初始化开始定时器
*/
func (timer *DailyRangeIntervalTimer) initStartTr(startDuration time.Duration) {
	if timer.startTr != nil {
		timer.startTr.Reset(startDuration)
	} else {
		timer.startTr = time.NewTimer(startDuration)
	}
	log.Println("下次启动时间：" + GetTimeByString(timer.Start).Add(ONE_DAY).String())
	timer.fnTodayExecCount = 0
	go func() {
		timer.startTimeChan <- <-timer.startTr.C
	}()
}

/*
开始间隔定时器
*/
func (timer *DailyRangeIntervalTimer) startTicker() *time.Ticker {
	if timer.intervalTk == nil {
		timer.intervalTk = time.NewTicker(timer.Interval)
	} else {
		timer.intervalTk.Reset(timer.Interval)
	}
	go func() {
		for {
			execTime := <-timer.intervalTk.C
			fmt.Println("exec task at " + execTime.String())
			if timer.FnExecLimit == 0 || timer.fnTodayExecCount < timer.FnExecLimit {
				//timer.fnTodayExecCount++
				timer.Fn(timer)
			}
		}
	}()
	return timer.intervalTk
}

/*
初始化范围结束计时器
*/
func (timer *DailyRangeIntervalTimer) initEndTr(endDuration time.Duration) {
	if timer.endTr != nil {
		timer.endTr.Reset(endDuration)
	} else {
		timer.endTr = time.NewTimer(endDuration)
	}
	timer.fnTodayExecCount = 0;
	go func() {
		timer.endTimeChan <- <-timer.endTr.C
	}()
}

func (timer *DailyRangeIntervalTimer) AddExecCount() {
	timer.fnTodayExecCount++
}
