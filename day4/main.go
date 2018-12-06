package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
)

func main() {
	if err := run(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}

type logDate struct {
	year, month, day int
}

type logTime struct {
	hour, minute int
}

type logEntry struct {
	logDate
	logTime
	guard int
	event eventType
}

type schedule struct {
	logDate
	sched [60]bool
}

type eventType uint8

const (
	eventNone eventType = iota
	eventBegin
	eventSleep
	eventWake
)

func run(r io.Reader) error {
	// read log events
	ents, err := readLogEntries(r)
	if err != nil {
		return err
	}

	// sort by date/time
	sort.Slice(ents, func(i, j int) bool {

		if ents[i].year < ents[j].year {
			return true
		} else if ents[i].year == ents[j].year {
			if ents[i].month < ents[j].month {
				return true
			} else if ents[i].month == ents[j].month {
				if ents[i].day < ents[j].day {
					return true
				} else if ents[i].day == ents[j].day {
					if ents[i].hour < ents[j].hour {
						return true
					} else if ents[i].hour == ents[j].hour {
						if ents[i].minute < ents[j].minute {
							return true
						} else if ents[i].minute == ents[j].minute {
							return ents[i].event < ents[j].event
						}
					}
				}
			}
		}

		return false
	})

	// fill in guard-on-duty
	guard := 0
	for i := range ents {
		if ents[i].guard == 0 {
			ents[i].guard = guard
		} else {
			guard = ents[i].guard
		}
	}

	// collect schedules, with optional dump
	schedules := collectSchedules(ents)

	// tabulate total guard time
	var worstGuard, mostTimeAsleep int
	for guard, scheds := range schedules {
		n := 0
		for _, sched := range scheds {
			n += sched.totalAsleep()
		}
		if mostTimeAsleep < n {
			worstGuard, mostTimeAsleep = guard, n
		}
	}
	log.Printf("Guard #%v is the worst with %v total minutes asleep", worstGuard, mostTimeAsleep)

	// for that guard, find its worst minute
	var counts [60]int
	fmt.Printf("      Date 000000000011111111112222222222333333333344444444445555555555\n")
	fmt.Printf("           012345678901234567890123456789012345678901234567890123456789\n")
	for _, sched := range schedules[worstGuard] {
		fmt.Printf("%v %s\n", sched.logDate, schedBytes(sched.sched))
		sched.incrementCounts(&counts)
	}
	var worstMinute, maxCount int
	for minute, count := range counts {
		if maxCount < count {
			worstMinute, maxCount = minute, count
		}
	}
	log.Printf("worst minute is %v, seen asleep %v times", worstMinute, maxCount)
	log.Printf("answer (part 1): %v", worstGuard*worstMinute)

	// now find the guard with the most frequent / consistent worst minute
	worstGuard, mostTimeAsleep = 0, 0
	var worstGuardMinute int
	for guard, scheds := range schedules {
		var counts [60]int
		for _, sched := range scheds {
			sched.incrementCounts(&counts)
		}
		var worstMinute, maxCount int
		for minute, count := range counts {
			if maxCount < count {
				worstMinute, maxCount = minute, count
			}
		}

		if mostTimeAsleep < maxCount {
			worstGuard, mostTimeAsleep = guard, maxCount
			worstGuardMinute = worstMinute
		}

		// log.Printf("#% 5v worst minute is %v seen %v times", guard, worstMinute, maxCount)
	}

	log.Printf("#%v worst minute is %v, seen %v times, that was the worst",
		worstGuard, worstGuardMinute, mostTimeAsleep)
	log.Printf("answer (part 2): %v", worstGuard*worstGuardMinute)

	return nil
}

func (sched schedule) incrementCounts(counts *[60]int) {
	for i, b := range sched.sched {
		if b {
			counts[i]++
		}
	}
}

func (sched schedule) totalAsleep() (n int) {
	for _, asleep := range sched.sched {
		if asleep {
			n++
		}
	}
	return n
}

func collectSchedules(ents []logEntry) map[int][]schedule {
	schedules := make(map[int][]schedule, 1024)
	dump := false

	var cur struct {
		schedule
		guard int
		awake bool
		last  logTime
	}

	if dump {
		// fmt.Printf("% 10s % 6s %s\n", "Date", "ID", "Minute")
		fmt.Printf("      Date     ID 000000000011111111112222222222333333333344444444445555555555\n")
		fmt.Printf("                  012345678901234567890123456789012345678901234567890123456789\n")
	}

	fill := func(t logTime, sleeping bool) {
		if cur.last.hour != 0 {
			if cur.last.hour != 23 {
				panic("inconceivable start hour")
			}
			cur.last = logTime{0, 0}
		}
		for ; cur.last.minute < t.minute; cur.last.minute++ {
			cur.sched[cur.last.minute] = sleeping
		}
		cur.last = t
	}

	flush := func() {
		fill(logTime{0, 60}, false)

		if dump {
			fmt.Printf("%v #% 5v %s\n", cur.logDate, cur.guard, schedBytes(cur.sched))
		}
		schedules[cur.guard] = append(schedules[cur.guard], cur.schedule)
	}

	for _, ent := range ents {
		// log.Printf("ENT: %+v", ent)
		if ent.event == eventBegin {
			if cur.guard != 0 {
				flush()
			}
			cur.logDate = ent.logDate
			cur.guard = ent.guard
			cur.awake = true
			for i := range cur.sched {
				cur.sched[i] = false
			}
			cur.last = ent.logTime
			continue
		}

		if ent.hour != 0 {
			log.Printf("IGNORE %v", ent)
			continue
		}
		switch ent.event {
		case eventSleep:
			cur.awake = false
		case eventWake:
			cur.awake = true
		default:
			log.Printf("IGNORE %v", ent)
			continue
		}
		fill(ent.logTime, cur.awake)
	}
	if cur.guard != 0 {
		flush()
	}

	return schedules
}

var logEntryPattern = regexp.MustCompile(
	`^\[(\d+)-(\d+)-(\d+) +(\d+):(\d+)\] +(.+)$`,
)

var logMessPattern = regexp.MustCompile(
	`(^falls asleep$)` +
		`|` +
		`(^wakes up$)` +
		`|` +
		`(?:^Guard #(\d+) begins shift$)`,
)

func readLogEntries(r io.Reader) ([]logEntry, error) {
	var ents []logEntry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := logEntryPattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			log.Printf("NO MATCH in %q", line)
			continue
		}

		var ent logEntry
		ent.year, _ = strconv.Atoi(parts[1])
		ent.month, _ = strconv.Atoi(parts[2])
		ent.day, _ = strconv.Atoi(parts[3])
		ent.hour, _ = strconv.Atoi(parts[4])
		ent.minute, _ = strconv.Atoi(parts[5])

		mess := parts[6]
		parts = logMessPattern.FindStringSubmatch(mess)
		if len(parts) == 0 {
			log.Printf("NO MATCH in %q", mess)
			continue
		}

		if parts[1] != "" {
			ent.event = eventSleep
		} else if parts[2] != "" {
			ent.event = eventWake
		} else if parts[3] != "" {
			ent.event = eventBegin
			ent.guard, _ = strconv.Atoi(parts[3])
		} else {
			log.Printf("BOGUS parts %q from %q", parts, mess)
			continue
		}

		ents = append(ents, ent)
	}
	return ents, sc.Err()
}

func (d logDate) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func (t logTime) String() string {
	return fmt.Sprintf("%02d:%02d", t.hour, t.minute)
}

func (ent logEntry) String() string {
	return fmt.Sprintf("[%v %v] #% 5v %v", ent.logDate, ent.logTime, ent.guard, ent.event)
}

func (event eventType) String() string {
	switch event {
	case eventBegin:
		return "Begin shift"
	case eventSleep:
		return "Sleep"
	case eventWake:
		return "Wake"
	default:
		return fmt.Sprintf("???<%v>", uint8(event))
	}
}

func schedBytes(sched [60]bool) []byte {
	var bs [60]byte
	for i, b := range sched {
		if b {
			bs[i] = '#'
		} else {
			bs[i] = '.'
		}
	}
	return bs[:]
}
