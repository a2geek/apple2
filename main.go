package main

import "go6502/apple2"

func main() {
	//romFile := "apple2/romdumps/Apple2.rom"
	romFile := "apple2/romdumps/Apple2_Plus.rom"
	//romFile := "apple2/romdumps/Apple2e.rom"
	disk2RomFile := "apple2/romdumps/DISK2.rom"
	diskImage := "../dos33.nib"

	log := false
	a := apple2.NewApple2(romFile)
	a.AddDisk2(disk2RomFile, diskImage)
	a.Run(log)
}
