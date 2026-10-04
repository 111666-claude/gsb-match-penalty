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

func TestTierEscalatesAndCapsAtLast(t *testing.T) {
	guard := New([]int{5000, 30000, 120000}, 1000000)
	want := []int{5000, 30000, 120000, 120000, 120000}
	for index, w := range want {
		if got := guard.Leave("p", string(rune('a'+index)), index); got != w {
			t.Fatalf("第 %d 次退出应该是 %d，得到 %d", index+1, w, got)
		}
	}
}

func TestBanWindowStartsAtLeaveAndEndsAtDeadline(t *testing.T) {
	guard := New([]int{5000, 30000}, 50000)
	guard.Leave("p", "m1", 1000)
	if guard.Banned("p", 999) {
		t.Fatal("退出时刻之前不该被禁赛")
	}
	if !guard.Banned("p", 1000) {
		t.Fatal("退出时刻起应该处于禁赛中")
	}
	if !guard.Banned("p", 5999) {
		t.Fatal("截止时刻之前应该处于禁赛中")
	}
	if guard.Banned("p", 6000) {
		t.Fatal("到达截止时刻就不该再算禁赛")
	}
}

func TestDecayDropsOneLevelPerInterval(t *testing.T) {
	guard := New([]int{5000, 30000, 120000}, 50000)
	guard.Leave("p", "m1", 0)
	guard.Leave("p", "m2", 10)
	guard.Leave("p", "m3", 20)
	if got := guard.Strikes("p", 50019); got != 3 {
		t.Fatalf("未满 decayMs 不该衰减，得到 %d", got)
	}
	if got := guard.Strikes("p", 50020); got != 2 {
		t.Fatalf("满一个 decayMs 应该降一级，得到 %d", got)
	}
	if got := guard.Strikes("p", 50020); got != 2 {
		t.Fatalf("同一时刻重复查询结果应该相同，得到 %d", got)
	}
	if got := guard.Strikes("p", 100020); got != 1 {
		t.Fatalf("满两个 decayMs 应该降两级，得到 %d", got)
	}
}

func TestDecayFloorsAtZero(t *testing.T) {
	guard := New([]int{5000}, 50000)
	guard.Leave("p", "m1", 0)
	if got := guard.Strikes("p", 1000000); got != 0 {
		t.Fatalf("衰减最低为 0，得到 %d", got)
	}
	if got := guard.Strikes("p", 2000000); got != 0 {
		t.Fatalf("衰减后不该变成负数，得到 %d", got)
	}
}

func TestLeaveSettlesDecayBeforeCounting(t *testing.T) {
	guard := New([]int{5000, 30000, 120000}, 50000)
	guard.Leave("p", "m1", 0)
	guard.Leave("p", "m2", 10)
	if got := guard.Leave("p", "m3", 100000); got != 30000 {
		t.Fatalf("衰减一级后再计次应该落第二档，得到 %d", got)
	}
}

func TestDuplicateMatchIsCountedOnce(t *testing.T) {
	guard := New([]int{5000, 30000}, 50000)
	guard.Leave("p", "m1", 0)
	guard.Leave("p", "m1", 10)
	guard.Leave("p", "m1", 20)
	if got := guard.Strikes("p", 20); got != 1 {
		t.Fatalf("同一 matchID 只计一次，得到 %d", got)
	}
	if guard.Banned("p", 6000) {
		t.Fatal("重复上报不该重设禁赛窗口")
	}
	if !guard.Banned("p", 4999) {
		t.Fatal("首次上报的禁赛窗口应该保留")
	}
}

func TestDuplicateDoesNotBlockOtherMatches(t *testing.T) {
	guard := New([]int{5000, 30000}, 50000)
	guard.Leave("p", "m1", 0)
	guard.Leave("p", "m1", 10)
	if got := guard.Leave("p", "m2", 20); got != 30000 {
		t.Fatalf("新 matchID 应该正常计次升档，得到 %d", got)
	}
}

func TestSameSequenceReplaysIdentically(t *testing.T) {
	run := func() (int, int, bool) {
		guard := New([]int{5000, 30000, 120000}, 50000)
		last := 0
		for index, at := range []int{0, 10, 20, 100000, 100010} {
			last = guard.Leave("p", string(rune('a'+index)), at)
		}
		return last, guard.Strikes("p", 100010), guard.Banned("p", 100010)
	}
	firstDuration, firstStrikes, firstBanned := run()
	secondDuration, secondStrikes, secondBanned := run()
	if firstDuration != secondDuration || firstStrikes != secondStrikes || firstBanned != secondBanned {
		t.Fatal("同一串退出序列重复执行结果应该相同")
	}
}

func TestScannedDoesNotScaleWithPlayerCount(t *testing.T) {
	guard := New([]int{5000, 30000}, 50000)
	for index := 0; index < 3000; index++ {
		guard.Leave(string(rune(index)), "m", 0)
	}
	for index := 0; index < 3000; index++ {
		guard.Strikes("p", 0)
	}
	if got := guard.Scanned(); got > 6000 {
		t.Fatalf("Scanned 不该随玩家数乘查询次数增长，得到 %d", got)
	}
}
