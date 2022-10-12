package stats

import "time"

type Storage struct {
	Days []*Stat
}

func NewStorage() *Storage {
	s := new(Storage)
	s.Days = make([]*Stat, 3)
	s.Ticker()
	return s
}

func (s *Storage) Ticker() {
	go func() {
		for {
			if h := time.Now().Hour(); h == 0 {
				// add new at the beggining and delete last item
				s.Days = append([]*Stat{new(Stat)}, s.Days[:len(s.Days)-1]...)
			}
		}
	}()
}

type Stat struct {
	Clicks     int
	ApiQueries int
}