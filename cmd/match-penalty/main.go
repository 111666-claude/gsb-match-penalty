// Command match-penalty 跑中途退出惩罚样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/penalty"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("match-penalty", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "tier", "tier / decay / dup / ban / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "tier":
		guard := penalty.New([]int{5000, 30000, 120000}, 50000)
		guard.Leave("p", "m1", 0)
		guard.Leave("p", "m2", 10)
		fmt.Fprintf(stdout, "third=%d\n", guard.Leave("p", "m3", 20))
	case "decay":
		guard := penalty.New([]int{5000, 30000, 120000}, 50000)
		guard.Leave("p", "m1", 0)
		guard.Leave("p", "m2", 10)
		guard.Leave("p", "m3", 20)
		fmt.Fprintf(stdout, "strikes=%d\n", guard.Strikes("p", 100000))
	case "dup":
		guard := penalty.New([]int{5000, 30000}, 50000)
		guard.Leave("p", "m1", 0)
		guard.Leave("p", "m1", 10)
		fmt.Fprintf(stdout, "strikes=%d\n", guard.Strikes("p", 10))
	case "ban":
		guard := penalty.New([]int{5000, 30000}, 50000)
		guard.Leave("p", "m1", 0)
		fmt.Fprintf(stdout, "banned=%v\n", guard.Banned("p", 1000))
	case "scan":
		guard := penalty.New([]int{5000, 30000}, 50000)
		for index := 0; index < 3000; index++ {
			guard.Leave(fmt.Sprintf("p-%d", index), fmt.Sprintf("m-%d", index), 0)
		}
		for index := 0; index < 3000; index++ {
			guard.Strikes("p-0", 0)
		}
		fmt.Fprintf(stdout, "scanned=%d\n", guard.Scanned())
	default:
		fmt.Fprintln(stderr, "需要 --sample tier|decay|dup|ban|scan")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
