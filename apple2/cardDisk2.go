package apple2

import (
	"bufio"
	"os"
)

/*
https://applesaucefdc.com/woz/reference2/
Good explanation of the softswitches and the phases:
http://yesterbits.com/media/pubs/AppleOrchard/articles/disk-ii-part-1-1983-apr.pdf

35 tracks, 16 sectors, 256 bytes
NIB: 35 tracks 6656 bytes, 232960 bytes

*/
const maxHalfTrack = 68

type cardDisk2 struct {
	cardBase
	selected int // Only 0 and 1 supported
	drive    [2]cardDisk2Drive
}

type cardDisk2Drive struct {
	diskette     *diskette16sector
	currentPhase int
	power        bool
	writeMode    bool
	halfTrack    int
	position     int
}

// type softSwitchR func(io *ioC0Page) uint8

func newCardDisk2(filename string) *cardDisk2 {
	var c cardDisk2
	c.rom = loadCardRom(filename)

	// Phase control soft switches
	// Lazy emulation. It only checks for phases on and move the head
	// up or down depending on the previous phase.
	for i := 0; i < 4; i++ {
		func(phase int) {
			c.ssr[phase<<1] = func(_ *ioC0Page) uint8 {
				//fmt.Printf("DISKII: Phase %v off\n", phase)
				return 0
			}
			c.ssr[(phase<<1)+1] = func(_ *ioC0Page) uint8 {
				//fmt.Printf("DISKII: Phase %v on\n", phase)
				halfTrack := c.drive[c.selected].halfTrack
				delta := (phase - c.drive[c.selected].currentPhase + 4) % 4
				switch delta {
				case 1: // Up
					halfTrack++
				case 2: // Illegal, let's say up
					halfTrack++
				case 3: // Down
					halfTrack--
				case 0: // No chamge
				}

				if halfTrack > maxHalfTrack {
					halfTrack = maxHalfTrack
				} else if halfTrack < 0 {
					halfTrack = 0
				}
				c.drive[c.selected].halfTrack = halfTrack
				c.drive[c.selected].currentPhase = phase
				//fmt.Printf("DISKII: Current halftrack is %v\n", halfTrack)
				return 0
			}
		}(i)
	}

	// Other soft switches
	c.ssr[0x8] = func(_ *ioC0Page) uint8 {
		c.drive[c.selected].power = false
		//fmt.Printf("DISKII: Disk %v is off\n", c.selected)
		return 0
	}
	c.ssr[0x9] = func(_ *ioC0Page) uint8 {
		c.drive[c.selected].power = true
		//fmt.Printf("DISKII: Disk %v is on\n", c.selected)
		return 0
	}
	c.ssr[0xA] = func(_ *ioC0Page) uint8 {
		c.selected = 0
		//fmt.Printf("DISKII: Disk %v selected\n", c.selected)
		return 0
	}
	c.ssr[0xB] = func(_ *ioC0Page) uint8 {
		c.selected = 1
		//fmt.Printf("DISKII: Disk %v selected\n", c.selected)
		return 0
	}

	// Q6L
	c.ssr[0xC] = func(_ *ioC0Page) uint8 {
		//fmt.Printf("DISKII: Reading\n")
		drive := &c.drive[c.selected]
		if drive.diskette == nil {
			return 0xff
		}
		track := drive.halfTrack / 2
		value, newPosition := drive.diskette.read(track, drive.position)
		drive.position = newPosition
		//fmt.Printf("DISKII: Reading value 0x%02v from track %v, position %v\n", value, track, drive.position)
		return value
	}

	c.ssw[0xC] = func(_ *ioC0Page, value uint8) {
		//fmt.Printf("DISKII: Writing the value 0x%02x\n", value)
	}

	// Q6H
	c.ssr[0xD] = func(_ *ioC0Page) uint8 {
		c.drive[c.selected].writeMode = false
		//fmt.Printf("DISKII: Sense write protection\n")
		return 0
	}

	// Q7L
	c.ssr[0xE] = func(_ *ioC0Page) uint8 {
		c.drive[c.selected].writeMode = false
		//fmt.Printf("DISKII: Set read mode\n")
		return 0
	}

	// Q7H
	c.ssr[0xF] = func(_ *ioC0Page) uint8 {
		c.drive[c.selected].writeMode = true
		//fmt.Printf("DISKII: Set write mode\n")
		return 0
	}
	return &c
}

func loadCardRom(filename string) []memoryPage {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(err)
	}

	size := stats.Size()
	bytes := make([]byte, size)
	buf := bufio.NewReader(f)
	buf.Read(bytes)

	pages := size / 256
	if (size % 256) > 0 {
		pages++
	}

	rom := make([]romPage, pages)
	for i := int64(0); i < size; i++ {
		rom[i>>8].burn(uint8(i), bytes[i])
	}

	memPages := make([]memoryPage, pages)
	for i := range rom {
		memPages[i] = &rom[i]
	}

	return memPages
}

func (d *cardDisk2Drive) insertDiskette(dt *diskette16sector) {
	d.diskette = dt
}