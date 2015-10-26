package vtclean

import (
	"bytes"
	"regexp"
	"strconv"
)

// see regex.txt for a slightly separated version of this regex
var vt100re = regexp.MustCompile(`^\033([\[\]]([\d\?]+)?(;[\d\?]+)*)?(.)`)
var vt100exc = regexp.MustCompile(`^\033(\[[^a-zA-Z0-9@\?]+|[\(\)]).`)

func Clean(line string, color bool) string {
	var edit = lineEdit{buf: make([]byte, len(line))}
	lineb := []byte(line)

	hadColor := false
	for i := 0; i < len(lineb); {
		c := lineb[i]
		switch c {
		case '\b':
			edit.Move(-1)
		case '\033':
			// set terminal title
			if bytes.HasPrefix(lineb[i:], []byte("\x1b]0;")) {
				pos := bytes.Index(lineb[i:], []byte("\a"))
				if pos != -1 {
					i += pos + 1
					continue
				}
			}
			if m := vt100exc.Find(lineb[i:]); m != nil {
				i += len(m)
			} else if m := vt100re.FindSubmatch(lineb[i:]); m != nil {
				i += len(m[0])
				num := string(m[2])
				n, err := strconv.Atoi(num)
				if err != nil || n > 10000 {
					n = 1
				}
				switch m[4][0] {
				case 'm':
					if color {
						hadColor = true
						edit.Write(m[0])
					}
				case '@':
					edit.Insert(bytes.Repeat([]byte{' '}, n))
				case 'C':
					edit.Move(n)
				case 'D':
					edit.Move(-n)
				case 'P':
					edit.Delete(n)
				case 'K':
					switch num {
					case "", "0":
						edit.ClearRight()
					case "1":
						edit.ClearLeft()
					case "2":
						edit.Clear()
					}
				}
			} else {
				i += 1
			}
			continue
		default:
			if c == '\n' || c >= ' ' {
				edit.Write([]byte{c})
			}
		}
		i += 1
	}
	out := edit.Bytes()
	if hadColor {
		out = append(out, []byte("\033[0m")...)
	}
	return string(out)
}
