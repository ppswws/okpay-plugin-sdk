# Query 池设计（PROCESSING 收敛）

## 1. 目标

本设计用于补齐 `Submit -> PROCESSING` 之后的收敛闭环，确保：

1. 有回调时可快速完成。
2. 无回调或回调延迟时可通过 `Query` 收敛终态。
3. `ORDER / REFUND / TRANSFER` 统一进入可审计、可扩展的状态机。

---

## 2. 结论（职责边界）

采用“**payment 负责业务状态机，manager 负责执行调度**”的分层：

1. `payment` 负责业务规则：
- 入池/出池规则
- 重试策略与超时策略
- 幂等与终态落库
- 告警判定标准

2. `manager` 负责执行能力：
- 定时拉任务
- 并发调用插件 `Query`
- 结果回传给 `payment`
- 统计指标采集与任务执行观测

该分层保证“业务规则单点收敛在 payment”，避免规则散落在执行器里。

---

## 3. 业务触发规则

### 3.1 入池

满足以下条件进入 Query 池：

1. `Submit(BizType)` 返回 `BizResult.state = PROCESSING`。
2. 业务类型是 `ORDER / REFUND / TRANSFER`（`BALANCE` 不入池）。
3. 对应业务记录存在且未终态。

### 3.2 出池

出现以下任一条件出池：

1. `Query` 返回 `SUCCEEDED`。
2. `Query` 返回 `FAILED`。
3. 渠道回调触发 `CompleteBiz` 并写成终态。
4. 超过最大重试/超时窗口，进入“人工处理或失败兜底”策略。

---

## 4. 统一状态机

对 `ORDER / REFUND / TRANSFER` 统一使用以下收敛流程：

1. `Submit` 返回 `SUCCEEDED/FAILED`：直接终态，不入池。
2. `Submit` 返回 `PROCESSING`：写库后入池。
3. 池任务触发 `Query`：
- `SUCCEEDED` -> 终态成功
- `FAILED` -> 终态失败
- `PROCESSING` -> 继续排队
4. 任意时刻如先收到回调并成功 `CompleteBiz`，则任务幂等退出。

---

## 5. 幂等与并发控制

### 5.1 幂等键

幂等键建议统一为：

1. `ORDER`：`trade_no`
2. `REFUND`：`refund_no`
3. `TRANSFER`：`trade_no`

### 5.2 并发原则

1. 同一幂等键同一时间只允许一个 Query 执行实例。
2. 回调与 Query 竞争时，以“状态已终态”为短路条件。
3. 终态写入必须条件更新（例如 `where status in (processing...)`）。

---

## 6. 调度模型（manager 执行）

建议采用“延迟队列 + worker pool”：

1. `payment` 写入任务（含下一次执行时间）。
2. `manager` 按 `next_run_at` 拉取可执行任务。
3. `manager` 并发执行并上报结果。
4. `payment` 根据结果决定“终态/重排/超时”。

重试间隔建议采用指数退避 + 上限封顶（可按业务类型配置）。

---

## 7. 策略配置（由 payment 解释）

按 `BizType` 单独配置：

1. `initial_delay`：首次查询延迟
2. `max_attempts`：最大重试次数
3. `max_window`：最大查询窗口（如 30m/2h）
4. `backoff`：退避策略（线性/指数）
5. `timeout_action`：超时动作（标记异常/失败/人工队列）

manager 只执行，不解释这些业务含义。

---

## 8. 结果回传契约

manager 执行 Query 后，将结果回传给 payment，最小需要：

1. `biz_type`
2. `biz_no`
3. `state`
4. `api_biz_no`
5. `channel_code`
6. `channel_msg`
7. `trace(req_ms/req_body/resp_body)`
8. `attempt`、`executed_at`（执行维度元数据）

payment 负责将该结果映射到业务记录并更新池任务状态。

---

## 9. 观测与审计

建议最少指标：

1. 入池总量、出池总量
2. 按 `BizType` 的处理中数量
3. 平均收敛时长、P95 收敛时长
4. 超时数量、重试次数分布
5. 渠道维度成功率/失败率/长处理率

审计要求：

1. 每次 Query 的请求/响应摘要可追溯。
2. 每次状态迁移有明确来源（submit/query/callback/manual）。

---

## 10. 与当前规约关系

本设计不改动 `payment/plugin/new.md` 已定规约，只补齐运行机制：

1. 插件侧继续使用 `Submit/Query` 返回 `BizResult`。
2. 回调侧继续在 `Handle` 中调用 `CompleteBiz`。
3. Query 池仅负责“PROCESSING 到终态”的调度与收敛。

---

## 11. 分阶段落地建议

### 阶段 A（最小可用）

1. `Submit=PROCESSING` 入池。
2. 固定间隔重试 Query。
3. 达到终态即出池。

### 阶段 B（可运营）

1. 按 BizType 区分策略。
2. 增加超时动作与告警。
3. 增加渠道维度统计。

### 阶段 C（可扩展）

1. 支持手工重试/重驱动。
2. 支持优先级队列与熔断。
3. 支持策略热更新。

