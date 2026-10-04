// Package penalty 是中途退出惩罚：按累犯次数给出阶梯禁赛时长，长期不犯会降级。
package penalty

type record struct {
	strikes int
	lastAt  int
	until   int
	matches map[string]struct{}
}

// Guard 是退出惩罚台账。
type Guard struct {
	tiers   []int
	decayMs int
	players map[string]*record
	scanned int
}

// New 建台账：tiers 是累犯阶梯时长（毫秒），decayMs 是降级所需的无犯间隔。
func New(tiers []int, decayMs int) *Guard {
	return &Guard{tiers: append([]int{}, tiers...), decayMs: decayMs, players: map[string]*record{}}
}

// tier 返回累犯 strikes 次对应的禁赛时长：第 1 次取 tiers[0]，
// 第 2 次取 tiers[1]，之后一律取最后一档。
func (g *Guard) tier(strikes int) int {
	if len(g.tiers) == 0 || strikes <= 0 {
		return 0
	}
	if strikes > len(g.tiers) {
		strikes = len(g.tiers)
	}
	return g.tiers[strikes-1]
}

// settle 惰性结算衰减：距上一次退出达到 decayMs 时累犯次数降一级，最低为 0。
func (g *Guard) settle(item *record, atMs int) {
	if item.strikes > 0 && atMs-item.lastAt >= g.decayMs {
		item.strikes--
		item.lastAt = atMs
	}
}

// Leave 记录一次中途退出，返回本次的禁赛时长。
// 同一 matchID 只计一次，重复上报不增加累犯次数也不重设禁赛。
func (g *Guard) Leave(player, matchID string, atMs int) int {
	item := g.players[player]
	if item == nil {
		item = &record{matches: map[string]struct{}{}}
		g.players[player] = item
	}
	g.settle(item, atMs)
	if _, seen := item.matches[matchID]; seen {
		return g.tier(item.strikes)
	}
	item.matches[matchID] = struct{}{}
	item.strikes++
	item.lastAt = atMs
	duration := g.tier(item.strikes)
	item.until = atMs + duration
	return duration
}

// Banned 判断玩家此刻是否在禁赛中：atMs 小于截止时刻才算禁赛中。
func (g *Guard) Banned(player string, atMs int) bool {
	item := g.players[player]
	if item == nil {
		return false
	}
	return atMs < item.until
}

// Strikes 返回玩家当前的累犯次数（惰性结算衰减），只查单个玩家，不全扫表。
func (g *Guard) Strikes(player string, atMs int) int {
	g.scanned++
	item := g.players[player]
	if item == nil {
		return 0
	}
	g.settle(item, atMs)
	return item.strikes
}

// Scanned 是累计扫描的玩家数（规模观测）。
func (g *Guard) Scanned() int { return g.scanned }
