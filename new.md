# 插件规约重构方案（最终版）

## 1. 背景与目标
本次重构的唯一目标是：

1. 缩减并整理内核 <-> 插件规约。
2. 消除位置参数式返回（如 `"", "", ...`）带来的可读性与审计歧义。
3. 在不降低现有功能可用性的前提下完成协议统一。

当前功能“基本可用”，但规约分裂、返回语义不统一，已经影响维护与审计。

## 2. 强约束（必须遵守）

1. 本次重构不考虑旧插件兼容。
2. 项目未上线，插件将跟随新规约统一升级。
3. 重构完成后，旧实现/死代码/冗余代码必须删除，不允许保留“过渡遗留”。
4. 不允许新增会引发审计歧义的协议字段或语义注释。
5. 不允许继续使用位置参数式返回 helper。

## 3. 新的固定 RPC 规约

保留并统一为 4 个固定入口：

1. `Info`
2. `Handle`（页面/回调入口）
3. `Submit`（发起业务）
4. `Query`（查询业务）

删除旧入口：

- `Create`
- `Refund`
- `Transfer`
- `Balance`
- `InvokeFunc`

说明：`Handle` 负责页面渲染与回调处理；`Submit/Query` 负责业务状态流转。

## 4. 统一业务类型

使用枚举 `BizType`：

- `ORDER`
- `REFUND`
- `TRANSFER`
- `BALANCE`

约束：

1. 内核必须做白名单强校验，仅允许上述类型。
2. 插件必须按 `BizType` 分派处理逻辑。
3. 代码中禁止使用数字字面量判断业务类型，必须使用枚举常量。

## 5. 统一请求结构

`Submit` 与 `Query` 使用同一请求结构：

- `ctx`：`InvokeContext`（沿用现有快照，不做拆分）
- `biz_type`
- `biz_no`（`balance` 可为空）

说明：

1. 不新增 `ParentBizNo`、`Amount`、`Extra` 等冗余字段。
2. 插件从 `InvokeContext` 获取订单/退款/转账上下文。

## 6. 统一返回结构（核心）

`Submit` 与 `Query` 统一返回 `BizResult`，字段语义如下：

- `state`：`FAILED / PROCESSING / SUCCEEDED`
- `api_biz_no`
- `channel_code`
- `channel_msg`
- `trace.req_ms`
- `trace.req_body`
- `trace.resp_body`
- `balance`（仅 `BALANCE` 使用）

约束：

1. 页面相关字段不进入 `BizResult`。
2. 页面协议继续使用现有 `PageResponse`，避免影响 `orderPageEcho`。
3. 渠道业务失败通过 `BizResult.state=FAILED` 表达，不通过 RPC error 表达。
4. RPC error 仅用于程序级错误（参数缺失、插件异常、序列化/验签实现异常等）。

## 7. 页面与回调规约

1. `Handle` 返回 `PageResponse`，沿用当前页面规约，不新增/不删除页面类型。
2. 回调处理仍在插件内完成验签与映射。
3. 插件回调后调用内核统一回写接口（见下一节）。

## 8. 插件 -> 内核回写收敛

回写接口统一为：`CompleteBiz`。

必备字段：

- `biz_type`
- `biz_no`
- `state`
- `api_biz_no`
- `channel_code`
- `channel_msg`
- `resp_body`
- `buyer`（仅 `ORDER` 使用）

删除旧回写接口：

- `CompleteOrder`
- `CompleteRefund`
- `CompleteTransfer`
- `RecordCNotify`

说明：渠道回调日志由内核统一记录，不再要求插件单独回写日志 RPC。

## 9. 状态流转规则

统一流转：

1. `Submit -> SUCCEEDED/FAILED`：直接终态。
2. `Submit -> PROCESSING`：进入轮询队列，后续执行 `Query`。
3. 收到渠道回调：插件 `Handle` 中调用 `CompleteBiz` 更新状态。
4. 即便有回调，仍允许 `Query` 佐证状态。

该规则适用于 `ORDER/REFUND/TRANSFER`。

## 10. SDK 与代码风格要求

### 10.1 SDK
新增结构化构造器，仅返回对象，且必须使用结构体命名字段入参：

- `ResultOK(BizResultInput{...})`
- `ResultPending(BizResultInput{...})`
- `ResultFail(BizResultInput{...})`
- `ResultBal(BizResultInput{...})`

### 10.2 代码风格

1. 所有返回使用命名字段初始化。
2. 禁止 `RespXxx(a, b, c, d, e, f)` 这种多位置参数 API。
3. 禁止“空字符串占位式”返回写法。

## 11. 重构实施步骤

1. 更新 `proto`：定义 `Info/Handle/Submit/Query`、`BizRequest`、`BizResult`、`CompleteBiz`。
2. 更新 `contract` 与 `host`：替换旧 RPC 接口调用链。
3. 更新 `sdk`：移除位置参数返回 helper，提供结构化结果构造器。
4. 更新内核服务层：
   - `order` 改走 `Submit(ORDER)`
   - `refund` 改走 `Submit(REFUND)`
   - `transfer` 改走 `Submit(TRANSFER)`
   - 统一查询走 `Query`
5. 增加/整理轮询机制：处理 `PROCESSING` 的 `REFUND/TRANSFER/ORDER`。
6. 更新全部插件到新接口。
7. 删除旧接口实现与所有死代码。
8. 全量编译、联调、回归。

## 12. 验收标准

### 12.1 规约层

1. 代码库不存在旧 RPC 声明与调用。
2. 代码库不存在位置参数式响应 helper。
3. `Submit/Query` 仅返回 `BizResult`。
4. `Handle` 仅返回 `PageResponse`。

### 12.2 功能层

1. 现有支付下单页面链路可用。
2. 退款与转账：同步 + 回调 + 查询均可正确收敛终态。
3. 有回调渠道可回调完成；无回调或未回调可通过 `Query` 收敛。
4. `balance` 通过 `Query(BALANCE)` 可用。

### 12.3 清理层

1. 无死代码。
2. 无双轨实现。
3. 无“临时兼容”残留。

## 13. 风险控制

1. 重构期间严禁同时维护新旧两套业务实现。
2. 每一步改造后必须执行编译与核心链路回归。
3. 变更以“规约一致性优先”，避免插件各自解释字段语义。

---

本方案为执行基线。后续实现、评审、联调均以本文件为准。
