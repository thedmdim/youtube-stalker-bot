package stats

import "time"

const DateFormat string = "2006/01/02"

type Storage struct {
	Days map[string]*Stat
}

func NewStorage() *Storage {
	today := time.Now()
	days := make(map[string]*Stat)
	for i:=0; i>-3; i-- {
		days[today.AddDate(0,0,i).Format(DateFormat)] = new(Stat)
	}
	return &Storage{days}
}

func (s *Storage) Today() *Stat {
	today := time.Now()
	if stat, ok := s.Days[today.Format(DateFormat)]; ok {
		return stat
	} else {
		stat := new(Stat)
		s.Days[today.Format(DateFormat)] = stat

		daym0 := today.Format(DateFormat)
		daym1 := today.AddDate(0,0,-1).Format(DateFormat)
		daym2 := today.AddDate(0,0,-2).Format(DateFormat)

		for k := range s.Days {
			if k!=daym2 || k!=daym1 || k!=daym0 {
				delete(s.Days, k)
			}
		}
		return stat
	}
}


type Stat struct {
	Clicks     int
	ApiQueries int
}