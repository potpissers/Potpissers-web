package main

import "time"

type abstractEvent struct {
	Message   string
	Timestamp time.Time
}

var events = func() []abstractEvent {
	var events []abstractEvent
	i, j := 0, 0
	for i < len(koths) && j < len(supplyDrops) {
		koth, drop := koths[i], supplyDrops[j]
		if koth.EndTimestamp.After(drop.EndTimestamp) {
			events = append(events, abstractEvent{koth.CapMessage, koth.EndTimestamp})
			i++
		} else {
			events = append(events, abstractEvent{drop.WinMessage, drop.EndTimestamp})
			j++
		}
	}
	for i < len(koths) {
		koth := koths[i]
		events = append(events, abstractEvent{koth.CapMessage, koth.EndTimestamp})
	}
	for j < len(supplyDrops) {
		drop := supplyDrops[j]
		events = append(events, abstractEvent{drop.WinMessage, drop.EndTimestamp})
	}
	return events
}()
