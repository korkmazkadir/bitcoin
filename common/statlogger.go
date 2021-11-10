package common

import (
	"fmt"
	"log"
	"time"
)

type EventType int

const (
	Proposed EventType = iota
	BlockReceived
	Echo
	Accept
	EndOfRound
)

func (e EventType) String() string {
	switch e {
	case Proposed:
		return "PROPOSED"
	case BlockReceived:
		return "BLOCK_RECEIVED"
	case Echo:
		return "ECHO"
	case Accept:
		return "ACCEPT"
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

func (s *StatLogger) LogBlockReceive(elapsedTime int64, round int) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, s.round, "BLOCK_RECEIVED", elapsedTime)
	s.events = append(s.events, Event{Round: round, Type: BlockReceived, ElapsedTime: int(elapsedTime)})
}

func (s *StatLogger) LogEndOfRound() {
	elapsedTime := time.Since(s.roundStart).Milliseconds()
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, s.round, "END_OF_ROUND", elapsedTime)
	s.events = append(s.events, Event{Round: s.round, Type: EndOfRound, ElapsedTime: int(elapsedTime)})
}

func (s *StatLogger) GetEvents() []Event {
	return s.events
}
