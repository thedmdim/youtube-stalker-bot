package stats

import "time"

type Storage struct {
	Days map[int]*Stat
}

func NewStorage() *Storage {
	day := time.Now().Day()
	storage := make(map[int]*Stat)
	for i:=0; i<4; i++ { 
		storage[day]=new(Stat)
		day-=1
	}
	return &Storage{
		Days: storage,
	}
}

func (s *Storage) Today() *Stat {
	day := time.Now().Day()
	if stat, ok := s.Days[day]; ok {
		return stat
	} else {
		delete(s.Days, day-3)
		stat := new(Stat)
		s.Days[day] = stat
		return stat
	}
}

type Stat struct {
	Clicks     int
	ApiQueries int
}
