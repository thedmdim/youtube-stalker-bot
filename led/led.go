package led

import (
	"fmt"
	"os"
)

const Red string = "/sys/devices/platform/gpio-leds/leds/F@ST2704N:red:inet/trigger"

func LedSwitch(line string) {
	// LED on = default-on
	// LED off = none
	in, _ := os.Create(Red)
	if in != nil {
		fmt.Fprint(in, line)
	}
	
	defer in.Close()
}