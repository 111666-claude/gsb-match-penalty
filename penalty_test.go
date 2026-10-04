package penalty

import "testing"

func TestFreshPlayerIsNotBanned(t *testing.T) {
	if New([]int{5000, 30000}, 50000).Banned("p", 0) {
		t.Fatal("没退出过的玩家不该被禁赛")
	}
}

func TestFirstLeaveUsesFirstTier(t *testing.T) {
	guard := New([]int{5000, 30000}, 50000)
	if got := guard.Leave("p", "m1", 0); got != 5000 {
		t.Fatalf("第一次退出应该是第一档，得到 %d", got)
	}
}

func TestUnknownPlayerHasNoStrikes(t *testing.T) {
	if New([]int{5000}, 50000).Strikes("nobody", 0) != 0 {
		t.Fatal("没退出过的玩家累犯次数应该是 0")
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if New([]int{5000}, 50000).Scanned() != 0 {
		t.Fatal("Scanned 初始应该是 0")
	}
}

func TestTierLadderClampsToLastTier(t *testing.T) {
	guard := New([]int{5000, 30000, 120000}, 50000)
	for step, want := range []int{5000, 30000, 120000, 120000, 120000} {
		if got := guard.Leave("p", string(rune('a'+step)), step*10); got != want {
			t.Fatalf("第 %d 次退出应禁赛 %d，得到 %d", step+1, want, got)
		}
	}
}

func TestDecayDropsOneLevelAfterQuietPeriod(t *testing.T) {
	guard := New([]int{5000, 30000, 120000}, 50000)
	guard.Leave("p", "m1", 0)
	guard.Leave("p", "m2", 10)
	guard.Leave("p", "m3", 20)
	if got := guard.Strikes("p", 100000); got != 2 {
		t.Fatalf("静默期达到 decayMs 应降一级到 2，得到 %d", got)
	}
	if got := guard.Strikes("p", 100000); got != 2 {
		t.Fatalf("同一时刻重复查询结果应相同，得到 %d", got)
	}
}

func TestDecayNeverDropsBelowZero(t *testing.T) {
	guard := New([]int{5000}, 50000)
	guard.Leave("p", "m1", 0)
	if got := guard.Strikes("p", 60000); got != 0 {
		t.Fatalf("累犯最低为 0，得到 %d", got)
	}
	if got := guard.Strikes("p", 120000); got != 0 {
		t.Fatalf("累犯最低为 0，得到 %d", got)
	}
}

func TestDecayAppliesBeforeNextLeave(t *testing.T) {
	guard := New([]int{5000, 30000}, 50000)
	guard.Leave("p", "m1", 0)
	guard.Leave("p", "m2", 10)
	if got := guard.Leave("p", "m3", 100000); got != 30000 {
		t.Fatalf("静默期降一级后再犯应按第 2 档 30000，得到 %d", got)
	}
}

func TestDuplicateMatchCountedOnce(t *testing.T) {
	guard := New([]int{5000, 30000}, 50000)
	first := guard.Leave("p", "m1", 0)
	second := guard.Leave("p", "m1", 10)
	if first != second {
		t.Fatalf("重复上报应返回相同结果，得到 %d 与 %d", first, second)
	}
	if got := guard.Strikes("p", 10); got != 1 {
		t.Fatalf("同一 matchID 只计一次，得到 %d", got)
	}
}

func TestDuplicateMatchDoesNotResetBan(t *testing.T) {
	guard := New([]int{5000}, 50000)
	guard.Leave("p", "m1", 0)
	guard.Leave("p", "m1", 4000)
	if !guard.Banned("p", 4999) {
		t.Fatal("截止时刻前应仍在禁赛")
	}
	if guard.Banned("p", 5000) {
		t.Fatal("重复上报不应重设禁赛截止时刻")
	}
}

func TestBanWindowEndsAtDeadline(t *testing.T) {
	guard := New([]int{5000}, 50000)
	guard.Leave("p", "m1", 1000)
	if !guard.Banned("p", 1000) {
		t.Fatal("退出时刻起算应处于禁赛")
	}
	if !guard.Banned("p", 5999) {
		t.Fatal("截止时刻前应处于禁赛")
	}
	if guard.Banned("p", 6000) {
		t.Fatal("查询时刻等于截止时刻不算禁赛")
	}
}

func TestSameSequenceSameResult(t *testing.T) {
	run := func() []int {
		guard := New([]int{5000, 30000}, 50000)
		return []int{
			guard.Leave("p", "m1", 0),
			guard.Leave("p", "m1", 5),
			guard.Leave("p", "m2", 10),
			guard.Strikes("p", 100000),
			guard.Strikes("p", 100000),
		}
	}
	first, second := run(), run()
	if len(first) != len(second) {
		t.Fatal("序列长度不一致")
	}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("第 %d 步结果不一致：%d 与 %d", index, first[index], second[index])
		}
	}
}

func TestStrikesDoesNotScanWholeTable(t *testing.T) {
	guard := New([]int{5000, 30000}, 50000)
	for index := 0; index < 3000; index++ {
		guard.Leave(string(rune(index)), string(rune(index)), 0)
	}
	for index := 0; index < 3000; index++ {
		guard.Strikes(string(rune(0)), 0)
	}
	if got := guard.Scanned(); got > 6000 {
		t.Fatalf("Scanned 不应随玩家数乘查询次数增长，得到 %d", got)
	}
}
