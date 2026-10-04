# match-penalty

中途退出惩罚：按累犯次数给出阶梯禁赛时长，长期不犯会降级，只用 Go 标准库。

```
go run ./cmd/match-penalty --sample=tier
go run ./cmd/match-penalty --sample=decay
go run ./cmd/match-penalty --sample=dup
go run ./cmd/match-penalty --sample=ban
go run ./cmd/match-penalty --sample=scan
go test ./...
go vet ./...
```

## 口径（README 为准）

- **阶梯**：第 1 次退出禁赛 `tiers[0]`，第 2 次 `tiers[1]`，之后一律取最后一档。
- **禁赛窗口**：本次禁赛从退出时刻起算，`Banned` 在 `atMs` 小于截止时刻时为真。
- **衰减**：距上一次退出达到 `decayMs` 时累犯次数降一级（惰性结算，查询时生效），最低为 0。
- **幂等**：同一 `matchID` 只计一次，重复上报不增加累犯次数也不重设禁赛。
- **不变量**：累犯次数单调不减（除衰减），次数不少于 0；同一串退出序列重复执行结果相同。
- **代价**：`Strikes` 不许每次全扫玩家表，`Scanned` 不随玩家数乘查询次数增长；10 万玩家每秒 10 万次查询。

## 输出契约（不改格式）

```
third=..
strikes=..
banned=..
scanned=..
```

## 目录

```
penalty.go            阶梯、衰减、幂等与禁赛窗口
penalty_test.go       单元测试
cmd/match-penalty/    命令行入口
```
