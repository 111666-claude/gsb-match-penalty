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
