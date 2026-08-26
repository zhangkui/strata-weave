# 出题包 strata-weave-bug-014

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | strata-weave-bug-014 |
| task_type | diagnosis |
| bug_category | nil相关问题 |
| repro_determinism | stable-panic |

## 仓库与环境（v2：repo_url）
分支模型: orphan-redgreen
绿测分支: bug14_green
红测分支: bug14_red
基线提交: 4a558907f0d0dab03399e5dda2c503be7a65381e

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/strata-weave/tree/bug14_green |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| green_branch 语义 | **初始化时的未修复问题基线，不是修复完成代码** |
| agent workspace | **固定：`F:\skillxm12\strata-weave-bug-014-green`**（**执行端按需 clone** `bug14_green` 后 `git remote remove origin`；create 端不预建） |
| 准备命令 | 见下方详细步骤 |
| go_version | go1.22 windows/amd64 (GOTOOLCHAIN=local) |
| harness | <执行端环境，**执行后填**> |

#### 准备命令详情（固定工作区 `F:\skillxm12`，本机 Git Bash 执行）
> 交付快照=单分支单提交：clone 后必须移除 remote（模型不可见其他分支/答案）。
```bash
git clone --single-branch -b bug14_green https://github.com/zhangkui/strata-weave.git F:\skillxm12/strata-weave-bug-014-green
cd F:\skillxm12/strata-weave-bug-014-green
git remote remove origin
git log --oneline --all   # 自检：必须只有 1 行（G1）
export GOTOOLCHAIN=local
go version          # 必须为 <go 版本>
go build ./...      # 初始必须可编译
go test ./...       # 复现/回归
```

## 用户需求（user_query）
请定位新采集样本尚未有测年结果时发生 panic 的原因。不要修改代码，只分析空结果在服务层和模型校验层之间的运行时崩溃路径。

## 验证命令（verify_cmds）
> ⚠️ **go test 格式（2026-08-22 用户规则，D-054 v2）**：回归测试文件在出题侧 `cases/bug-00N/regression/`，create 阶段不提交测试到 green/red；collect 在 test 轨迹验收后才把对应公开测试落位到验证环境。verify_cmds 每行 = 一条纯 go test 命令 `go test [flags] <包路径> -count=N -run '^TestBugNN_XXX$'`（并发题目标测试行加 `-race`，D-008）。**禁止 heredoc（cat > <<'GOEOF'）/docker run/bash *.sh/cp/base64**，验收方在测试落位后的干净 checkout 直接执行即可复现红绿。示例（**直接照抄形态，不带任何前缀标签**）：
```bash
go test ./cases/bug-014/regression -count=1 -run '^TestBug014_CollectedSampleDoesNotPanicWithoutDatingResult$'
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：服务层将无实验室信息的样本结果置空，模型校验却无条件解引用结果字段。
生产文件/符号：`internal/service/service.go` 的 `CreateSample`，以及 `internal/model/validation.go` 的 `ValidateSample`。
调用链：CreateSample -> ValidateSample -> DatingResult.Method。
失效原因：新采集样本在正常尚未测年的阶段触发空指针崩溃。
证据：创建无测年结果样本的回归用例捕获到 panic。 |
| gold_patch |  |

## 成功标准（success_criteria）
目标行为：诊断准确指出空结果的解引用路径。
边界：不把真正非法的样本标签和父遗物误判为崩溃根因。
合法场景：尚未测年的已采集样本可被正常创建。
验证标准：仅分析复现结果，不修改生产代码。

## docker 打包验证（阶段 5.5，交付前必跑）
- 构建：`./build_benzhi_docker.sh strata-weave-bug-14 linux/amd64`，再 `linux/arm64` 各一次（QEMU 模拟慢属正常）；**镜像名 `benzhi/strata-weave-bug-14:latest`（对齐 Go 项目打包规范模板）**
- 容器内：`go build ./...` 通过且无 `downloading` 字样；`go version` 工具链完整；直接运行对应公开回归测试，结果与本地校准一致
- ⚠️ `-race` 只在原生架构跑（QEMU 模拟架构报 VMA range 误伤）；测试代码禁止 `t.Context()`（go.mod 1.22 用 `context.Background()`）
- <实测结果（按 bug 逐条，每 bug 一行）：bug-014：amd64 构建 OK / arm64 构建 OK / 容器内复现退出码 <与 success_criteria 一致>。少任何 bug = 未完成>

## 轨迹收集清单（执行完成后，go-bug-collect 用，自动化）
- 你需带回：**模型修复总结**（复制模型最终回复）+ **模型执行轨迹 JSONL 路径**（单题 1 份）+ **两次 verify_cmds 验证轨迹**（2026-08-17 起：用字节提供的模型 glm52-coding 在两个新 session 分别执行修复前/修复后 verify_cmds，轨迹只含执行 verify_cmds 的过程，上传后带回 trajectory_url + session_id + result；bugfix 需 pre_fix(red) + post_fix(green) 两份，diagnosis 只需 pre_fix(red) 一份）；顺带说一句执行端是 claude 还是 codex
- ⚠️ **轨迹必须含题面（2026-08-17 v2.2 验收硬性）**：模型轨迹首条 user 文本事件必须 = user_query 原文——run_task.py 已改 `-p` 位置参数传题面 + `--append-system-prompt` 传前缀/要求，新轨迹天然满足；手动执行时同样只粘题面原文。轨迹无 user 题面或首轮带前缀 = 不合格，reset 重跑。
- 其余自动：session_id / generator_model / harness 由 go-bug-collect 从轨迹自动提取拼装；**verify_result = pre_fix/post_fix JSON**（bugfix 双项 / diagnosis 仅 pre_fix）
