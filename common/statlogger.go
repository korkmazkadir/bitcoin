package common

import (
	"fmt"
	"log"
	"time"
)

type EventType int

const (
	BlockReceived EventType = iota
	HopCount
	ProcessingTime
	EndOfRound
)

func (e EventType) String() string {
	switch e {
	case BlockReceived:
		return "BLOCK_RECEIVED"
	case HopCount:
		return "HOP_COUNT"
	case ProcessingTime:
		return "PROCESSING_TIME"
	case EndOfRound:
		return "END_OF_ROUND"
	default:
		panic(fmt.Errorf("undefined enum value %d", e))
	}
}

type Event struct {
	Round       int
	Type        EventType
	ElapsedTime int
	BlockHash   string
}

type StatList struct {
	IPAddress  string
	PortNumber int
	NodeID     int
	Events     []Event
}

type StatLogger struct {
	round      int
	roundStart time.Time
	nodeID     int

	events []Event
}

func NewStatLogger(nodeID int) *StatLogger {
	return &StatLogger{nodeID: nodeID}
}

func (s *StatLogger) NewRound(round int) {
	s.round = round
	s.roundStart = time.Now()
}

func (s *StatLogger) LogBlockReceived(round int, elapsedTime int, hopCount int) {
	s.events = append(s.events, Event{Round: round, Type: BlockReceived, ElapsedTime: int(elapsedTime)})
	s.events = append(s.events, Event{Round: round, Type: HopCount, ElapsedTime: hopCount})
}

func (s *StatLogger) LogProcessingTime(elapsedTime int) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, s.round, "PROCESSING_TIME", elapsedTime)
	s.events = append(s.events, Event{Round: s.round, Type: ProcessingTime, ElapsedTime: elapsedTime})
}

func (s *StatLogger) LogEndOfRound(macroblockHash []byte) {
	elapsedTime := time.Since(s.roundStart).Milliseconds()
	log.Printf("stats\t%d\t%d\t%s\t%d\t%x\t", s.nodeID, s.round, "END_OF_ROUND", elapsedTime, macroblockHash)
	s.events = append(s.events, Event{Round: s.round, Type: EndOfRound, ElapsedTime: int(elapsedTime), BlockHash: fmt.Sprintf("%x", macroblockHash)})
}

func (s *StatLogger) GetEvents() []Event {
	return s.events
}
