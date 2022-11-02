package stats

import (
	"time"
	"sync"
)

type Storage struct {
	sync.Mutex
	Days []*DayStat
}

func NewStorage() *Storage {
	s := new(Storage)

	for i:=0;i<3;i++{
		s.Days = append(s.Days, new(DayStat))
	}

	s.Ticker()
	return s
}

func (s *Storage) IncreaseTodaysQueries() {
	s.Lock()
	s.Days[0].ApiQueries+=1
	s.Unlock()
}

func (s *Storage) IncreaseTodaysClicks() {
	s.Lock()
	s.Days[0].Clicks+=1
	s.Unlock()
}

func (s *Storage) Ticker() {
	go func() {
		for {
			if h := time.Now().Hour(); h == 0 {
				s.Lock()
				// add new at the beggining and delete last item
				s.Days = append([]*DayStat{new(DayStat)}, s.Days[:len(s.Days)-1]...)
				s.Unlock()
				time.Sleep(time.Hour * 24)
			}
		}
	}()
}

type DayStat struct {
	Clicks     int
	ApiQueries int
}