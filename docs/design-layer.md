# 为什么有 design 层？

## 一句话

`internal/design` **不是业务层**，而是 **与业务无关的设计模式基础设施**（纯泛型框架），对齐 Java 原项目的：

```
cn.bugstack.wrench.design.framework
  ├── link  （责任链）
  └── tree  （策略树 / 路由树）
```

小傅哥工程里这块叫「扳手」公共组件：领域服务 **复用** 链/树骨架，而不是每个业务包手写 next 指针。

## 放在哪一层？

```
domain 业务 ──依赖──► design（纯抽象，无业务）
                         │
                    只依赖标准库 context
```

| 包 | 内容 | 是否含业务 |
|----|------|-----------|
| `design/chain` | `LogicHandler` / `BaseHandler` / `LinkedList` | 否 |
| `design/tree` | `StrategyHandler` / `AbstractStrategyRouter` | 否 |

**允许 domain 依赖 design**，因为：

1. 无 IO、无中间件、无 DTO、无 Entity 引用  
2. 属于「通用语言中的模式骨架」，类似标准库扩展  
3. 与 Java：`domain` 依赖 `wrench.design.framework` 一致  

**禁止 design 依赖 domain / api / infrastructure。**

## 谁在用？

| 业务 | 使用方式 |
|------|----------|
| 锁单规则 | `design/chain` → ActivityUsability → UserTakeLimit → TeamStock |
| 结算规则 | `design/chain` → SC → OutTradeNo → Settable → End |
| 退单规则 | `design/chain` → Data → Unique → RefundOrder |
| 试算流程 | 当前用领域内 `trial/node.Chain` 手写路由（语义等同策略树；`design/tree` 预留对齐 Java AbstractMultiThreadStrategyRouter） |

## 和 api 层区别

| | design | api |
|--|--------|-----|
| 目的 | 模式复用 | 对外契约 |
| 依赖方 | domain | trigger 实现 |
| 有无业务 | 无 | 无（仅 DTO/接口） |

## 面试怎么讲

> 「design 层对应原项目 wrench 设计模式框架，把责任链、策略树抽成与业务无关的泛型组件。交易锁单/结算/退单的规则过滤都挂在这条链上，扩展规则只加节点，符合开闭原则。这不是第四个业务层，而是领域实现模式的共享内核。」
