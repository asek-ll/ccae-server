package logger

var (
	TRACE Level = Level{00, [5]byte{'T', 'R', 'A', 'C', 'E'}}
	DEBUG Level = Level{10, [5]byte{'D', 'E', 'B', 'U', 'G'}}
	INFO  Level = Level{20, [5]byte{'I', 'N', 'F', 'O', ' '}}
	WARN  Level = Level{30, [5]byte{'W', 'A', 'R', 'N', ' '}}
	ERROR Level = Level{40, [5]byte{'E', 'R', 'R', 'O', 'R'}}
	FATAL Level = Level{50, [5]byte{'F', 'A', 'T', 'A', 'L'}}
	PANIC Level = Level{60, [5]byte{'P', 'A', 'N', 'I', 'C'}}
)

var levels = []Level{TRACE, DEBUG, INFO, WARN, ERROR, FATAL, PANIC}

type Level struct {
	Value int16
	name  [5]byte
}

func (l Level) Trim() string {
	if l.name[4] == ' ' {
		return string(l.name[:4])
	}
	return string(l.name[:])
}

func (l Level) Name() []byte {
	if l.name[4] == ' ' {
		return l.name[:4]
	}
	return l.name[:]
}

func (l Level) Padded() string {
	return string(l.name[:])
}

func (l Level) Braced() string {
	var buf [7]byte
	buf[0] = '['
	copy(buf[1:], l.name[:])
	if buf[5] == ' ' {
		buf[5] = ']'
		buf[6] = ' '
	} else {
		buf[6] = ']'
	}
	return string(buf[:])
}
