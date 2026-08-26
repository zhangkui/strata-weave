# 出题包 strata-weave-bug-006

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | strata-weave-bug-006 |
| task_type | bugfix |
| bug_category | concurrency并发问题 |
| repro_determinism | stable-assertion |

## 仓库与环境（v2：repo_url）
分支模型: orphan-redgreen
绿测分支: bug6_green
红测分支: bug6_red
基线提交: c6076fafed0e842afa30c43e0e7e0c6cb255b42e

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/strata-weave/tree/bug6_green |
| base_branch | 见上方分支字段和基线提交 |
| green_branch 语义 | **初始化时的未修复问题基线，不是修复完成代码** |
| agent workspace | **固定：`F:\skillxm12\strata-weave-bug-006-green`**（**执行端按需 clone** `bug6_green` 后 `git remote remove origin`；create 端不预建） |
| 准备命令 | 见下方详细步骤 |
| go_version | go1.22 windows/amd64 (GOTOOLCHAIN=local) |
| harness | <执行端环境，**执行后填**> |

#### 准备命令详情（固定工作区 `F:\skillxm12`，本机 Git Bash 执行）
> 交付快照=单分支单提交：clone 后必须移除 remote（模型不可见其他分支/答案）。
```bash
git clone --single-branch -b bug6_green https://github.com/zhangkui/strata-weave.git F:\skillxm12/strata-weave-bug-006-green
cd F:\skillxm12/strata-weave-bug-006-green
git remote remove origin
git log --oneline --all   # 自检：必须只有 1 行（G1）
export GOTOOLCHAIN=local
go version          # 必须为 <go 版本>
go build ./...      # 初始必须可编译
go test ./...       # 复现/回归
```

## 用户需求（user_query）
两名考古员同时提交相反方向的地层先后关系，环检测和写入之间存在并发竞态，Harris 矩阵可能形成环。请修复关系校验与提交的原子性。

## 验证命令（verify_cmds）
> ⚠️ **go test 格式**：目标测试在初始化时已提交到 Red/Green 分支；后续流程不得从 `cases/` 注入、复制或提交测试。verify_cmds 每行 = 一条纯 go test 命令 `go test [flags] <包路径> -count=N -run '^TestBugNN_XXX$'`（并发题目标测试行加 `-race`）。**禁止 heredoc（cat > <<'GOEOF'）/docker run/bash *.sh/cp/base64**；验收方在干净 checkout 直接执行即可复现红绿。示例（**直接照抄形态，不带任何前缀标签**）：
```bash
go test -race ./internal/regression -count=20 -run '^TestBug006_ConcurrentRelationsMustNotCloseCycle$'
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause |  |
| gold_patch |  |

## 成功标准（success_criteria）
目标行为：相反方向的并发关系不会形成地层环。
边界：同层自引用和跨探方关系仍必须拒绝。
合法场景：互不冲突的关系可以顺序加入矩阵。
验证标准：修复后验证命令成功；修复前应稳定接受两个相反关系。

## docker 打包验证（阶段 5.5，交付前必跑）
- 构建：`./build_benzhi_docker.sh strata-weave-bug-6 linux/amd64`，再 `linux/arm64` 各一次（QEMU 模拟慢属正常）；**镜像名 `benzhi/strata-weave-bug-6:latest`（对齐 Go 项目打包规范模板）**
- 容器内：`go build ./...` 通过且无 `downloading` 字样；`go version` 工具链完整；直接运行对应公开回归测试，结果与本地校准一致
- ⚠️ `-race` 只在原生架构跑（QEMU 模拟架构报 VMA range 误伤）；测试代码禁止 `t.Context()`（go.mod 1.22 用 `context.Background()`）
- <实测结果（按 bug 逐条，每 bug 一行）：bug-006：amd64 构建 OK / arm64 构建 OK / 容器内复现退出码 <与 success_criteria 一致>。少任何 bug = 未完成>

## 轨迹收集清单（执行完成后，go-bug-collect 用，自动化）
- 你需带回：**模型修复总结**（复制模型最终回复）+ **模型执行轨迹 JSONL 路径**（单题 1 份）+ **两次 verify_cmds 验证轨迹**（2026-08-17 起：用字节提供的模型 glm52-coding 在两个新 session 分别执行修复前/修复后 verify_cmds，轨迹只含执行 verify_cmds 的过程，上传后带回 trajectory_url + session_id + result；bugfix 需 pre_fix(red) + post_fix(green) 两份，diagnosis 只需 pre_fix(red) 一份）；顺带说一句执行端是 claude 还是 codex
- ⚠️ **轨迹必须含题面（2026-08-17 v2.2 验收硬性）**：模型轨迹首条 user 文本事件必须 = user_query 原文——run_task.py 已改 `-p` 位置参数传题面 + `--append-system-prompt` 传前缀/要求，新轨迹天然满足；手动执行时同样只粘题面原文。轨迹无 user 题面或首轮带前缀 = 不合格，reset 重跑。
- 其余自动：session_id / generator_model / harness 由 go-bug-collect 从轨迹自动提取拼装；**verify_result = pre_fix/post_fix JSON**（bugfix 双项 / diagnosis 仅 pre_fix）
