// Package penalty 是中途退出惩罚：按累犯次数给出阶梯禁赛时长，长期不犯会降级。
// 缺陷：阶梯不看累犯次数、没有衰减、同一场重复上报也计数、禁赛从不生效、
// 每次查询全扫玩家表。
package penalty

type record struct {
	strikes int
	lastAt  int
	until   int
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

// Leave 记录一次中途退出，返回本次的禁赛时长。
func (g *Guard) Leave(player, matchID string, atMs int) int {
	item := g.players[player]
	if item == nil {
		item = &record{}
		g.players[player] = item
	}
	item.strikes++
	item.lastAt = atMs
	if len(g.tiers) == 0 {
		return 0
	}
	return g.tiers[0]
}

// Banned 判断玩家此刻是否在禁赛中。
func (g *Guard) Banned(player string, atMs int) bool {
	item := g.players[player]
	if item == nil {
		return false
	}
	return item.until > atMs
}

// Strikes 返回玩家当前的累犯次数。
func (g *Guard) Strikes(player string, atMs int) int {
	g.scanned += len(g.players)
	item := g.players[player]
	if item == nil {
		return 0
	}
	return item.strikes
}

// Scanned 是累计扫描的玩家数（规模观测）。
func (g *Guard) Scanned() int { return g.scanned }
