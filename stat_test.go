package kebench

import (
	"fmt"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

func TestCpu(t *testing.T) {
	per1, _ := cpu.Percent(time.Second, true)
	per2, _ := cpu.Percent(time.Second, false)
	fmt.Println(per1, per2)
}
