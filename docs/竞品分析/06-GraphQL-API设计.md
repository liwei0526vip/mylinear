## 六、GraphQL API 设计

> 参考 Linear 官方 Schema：https://api.linear.app/graphql

### 6.1 核心 Schema 定义

```graphql
# ============ 核心类型 ============

type Issue {
  id: ID!
  team: Team!
  number: Int!           # 团队内序号，如 123
  identifier: String!    # 完整 ID，如 ENG-123
  title: String!
  description: String
  descriptionData: DocumentContent
  status: WorkflowState!
  priority: Priority!
  assignee: User
  project: Project
  milestone: Milestone
  cycle: Cycle
  parent: Issue
  children: IssueConnection!
  relations: IssueRelationConnection!
  labels: LabelConnection!
  comments: CommentConnection!
  attachments: AttachmentConnection!
  subscribers: UserConnection!
  dueDate: Date
  estimate: Int
  slaDueAt: DateTime
  createdAt: DateTime!
  updatedAt: DateTime!
  completedAt: DateTime
  cancelledAt: DateTime
  creator: User!
  activity: ActivityConnection!
}

type Project {
  id: ID!
  name: String!
  description: String
  status: ProjectStatus!
  priority: Priority!
  lead: User
  team: Team
  teams: TeamConnection!
  members: UserConnection!
  milestones: MilestoneConnection!
  issues: IssueConnection!
  startDate: Date
  targetDate: Date
  progress: Float
  createdAt: DateTime!
  updatedAt: DateTime!
  completedAt: DateTime
}

type Cycle {
  id: ID!
  team: Team!
  number: Int!
  name: String
  description: String
  startDate: Date!
  endDate: Date!
  cooldownEndDate: Date
  progress: Float
  issues: IssueConnection!
  createdAt: DateTime!
}

type Team {
  id: ID!
  name: String!
  key: String!
  workspace: Workspace!
  parent: Team
  children: TeamConnection!
  issues: IssueConnection!
  projects: ProjectConnection!
  cycles: CycleConnection!
  workflowStates: WorkflowStateConnection!
  labels: LabelConnection!
  members: TeamMemberConnection!
  timezone: String
  createdAt: DateTime!
}

type User {
  id: ID!
  email: String!
  name: String!
  displayName: String
  avatarUrl: String
  role: UserRole!
  teams: TeamConnection!
  assignedIssues: IssueConnection!
  createdIssues: IssueConnection!
  createdAt: DateTime!
}

type WorkflowState {
  id: ID!
  team: Team!
  name: String!
  type: WorkflowStateType!
  color: String
  position: Float!
  isDefault: Boolean!
}

type Label {
  id: ID!
  name: String!
  description: String
  color: String!
  parent: Label
  children: LabelConnection!
  isArchived: Boolean!
}

type Comment {
  id: ID!
  issue: Issue!
  user: User!
  body: String!
  bodyData: DocumentContent
  parent: Comment
  children: CommentConnection!
  createdAt: DateTime!
  updatedAt: DateTime!
  editedAt: DateTime
}

# ============ 枚举类型 ============

enum Priority {
  NONE
  URGENT
  HIGH
  MEDIUM
  LOW
}

enum ProjectStatus {
  PLANNED
  IN_PROGRESS
  PAUSED
  COMPLETED
  CANCELLED
}

enum WorkflowStateType {
  BACKLOG
  UNSTARTED
  STARTED
  COMPLETED
  CANCELLED
}

enum UserRole {
  ADMIN
  MEMBER
  GUEST
  TEAM_OWNER
  WORKSPACE_OWNER
}

enum IssueRelationType {
  BLOCKED_BY
  BLOCKING
  RELATED
  DUPLICATE
}

# ============ 连接类型（分页） ============

type IssueConnection {
  edges: [IssueEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type IssueEdge {
  node: Issue!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

# ============ 输入类型 ============

input IssueCreateInput {
  teamId: ID!
  title: String!
  description: String
  statusId: ID
  priority: Priority
  assigneeId: ID
  projectId: ID
  milestoneId: ID
  cycleId: ID
  parentId: ID
  labelIds: [ID!]
  estimate: Int
  dueDate: Date
}

input IssueUpdateInput {
  title: String
  description: String
  statusId: ID
  priority: Priority
  assigneeId: ID
  projectId: ID
  milestoneId: ID
  cycleId: ID
  parentId: ID
  labelIds: [ID!]
  estimate: Int
  dueDate: Date
}

input IssueFilterInput {
  team: TeamFilterInput
  status: WorkflowStateFilterInput
  priority: PriorityFilterInput
  assignee: UserFilterInput
  project: ProjectFilterInput
  cycle: CycleFilterInput
  labels: LabelFilterInput
  search: String
}

input IssueBatchUpdateInput {
  ids: [ID!]!
  input: IssueUpdateInput!
}
```

### 6.2 Query 操作

```graphql
type Query {
  # ============ Issue 查询 ============
  issue(id: ID!): Issue
  issueByIdentifier(identifier: String!): Issue  # 如 "ENG-123"
  issues(
    filter: IssueFilterInput
    orderBy: IssueOrderByInput
    first: Int
    after: String
    last: Int
    before: String
  ): IssueConnection!

  # ============ 项目查询 ============
  project(id: ID!): Project
  projects(
    filter: ProjectFilterInput
    orderBy: ProjectOrderByInput
    first: Int
    after: String
  ): ProjectConnection!

  # ============ 团队查询 ============
  team(id: ID!): Team
  teams(
    filter: TeamFilterInput
    first: Int
    after: String
  ): TeamConnection!

  # ============ Cycle 查询 ============
  cycle(id: ID!): Cycle
  cycles(
    filter: CycleFilterInput
    first: Int
    after: String
  ): CycleConnection!

  # ============ 用户查询 ============
  viewer: User!  # 当前登录用户
  user(id: ID!): User
  users(
    filter: UserFilterInput
    first: Int
    after: String
  ): UserConnection!

  # ============ 工作区查询 ============
  workspace: Workspace!

  # ============ 搜索 ============
  searchIssues(
    query: String!
    first: Int
    after: String
  ): IssueConnection!

  # ============ 过滤器 ============
  issueFilters: IssueFilters!
}
```

### 6.3 Mutation 操作

```graphql
type Mutation {
  # ============ Issue 操作 ============
  issueCreate(input: IssueCreateInput!): IssuePayload!
  issueUpdate(id: ID!, input: IssueUpdateInput!): IssuePayload!
  issueDelete(id: ID!): DeletePayload!
  issueBatchUpdate(input: IssueBatchUpdateInput!): IssueBatchPayload!
  issueArchive(id: ID!): IssuePayload!
  issueUnarchive(id: ID!): IssuePayload!

  # ============ Issue 关系 ============
  issueRelationCreate(
    issueId: ID!
    relatedIssueId: ID!
    type: IssueRelationType!
  ): IssueRelationPayload!
  issueRelationDelete(id: ID!): DeletePayload!

  # ============ 评论 ============
  commentCreate(issueId: ID!, body: String!): CommentPayload!
  commentUpdate(id: ID!, body: String!): CommentPayload!
  commentDelete(id: ID!): DeletePayload!

  # ============ 项目 ============
  projectCreate(input: ProjectCreateInput!): ProjectPayload!
  projectUpdate(id: ID!, input: ProjectUpdateInput!): ProjectPayload!
  projectDelete(id: ID!): DeletePayload!

  # ============ Cycle ============
  cycleCreate(input: CycleCreateInput!): CyclePayload!
  cycleUpdate(id: ID!, input: CycleUpdateInput!): CyclePayload!
  cycleDelete(id: ID!): DeletePayload!

  # ============ 团队 ============
  teamCreate(input: TeamCreateInput!): TeamPayload!
  teamUpdate(id: ID!, input: TeamUpdateInput!): TeamPayload!
  teamDelete(id: ID!): DeletePayload!

  # ============ 标签 ============
  labelCreate(input: LabelCreateInput!): LabelPayload!
  labelUpdate(id: ID!, input: LabelUpdateInput!): LabelPayload!
  labelDelete(id: ID!): DeletePayload!

  # ============ 工作流 ============
  workflowStateCreate(input: WorkflowStateCreateInput!): WorkflowStatePayload!
  workflowStateUpdate(id: ID!, input: WorkflowStateUpdateInput!): WorkflowStatePayload!
  workflowStateDelete(id: ID!): DeletePayload!
}

# ============ 返回类型 ============

type IssuePayload {
  success: Boolean!
  issue: Issue
  errors: [Error!]!
}

type IssueBatchPayload {
  success: Boolean!
  issues: [Issue!]
  errors: [Error!]!
}

type DeletePayload {
  success: Boolean!
  errors: [Error!]!
}

type Error {
  message: String!
  code: String!
  path: [String!]
}
```

### 6.4 Subscription（实时订阅）

```graphql
type Subscription {
  # Issue 变更
  issueUpdated(teamId: ID, projectId: ID): IssueUpdatePayload!

  # 评论新增
  commentCreated(issueId: ID!): Comment!

  # 通知
  notifications: NotificationPayload!

  # 工作区活动
  workspaceActivity: ActivityPayload!
}

type IssueUpdatePayload {
  issue: Issue!
  type: UpdateType!
  changedFields: [String!]
}

enum UpdateType {
  CREATED
  UPDATED
  DELETED
  ARCHIVED
}
```

### 6.5 分页实现（游标分页）

```graphql
# 查询示例
query Issues($first: Int, $after: String, $filter: IssueFilterInput) {
  issues(first: $first, after: $after, filter: $filter) {
    edges {
      node {
        id
        identifier
        title
        status {
          id
          name
        }
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

**服务端分页逻辑**：
```go
// Go 后端分页实现示例
type IssueConnection struct {
    Edges       []IssueEdge
    PageInfo    PageInfo
    TotalCount  int
}

type PageInfo struct {
    HasNextPage     bool
    HasPreviousPage bool
    StartCursor     *string
    EndCursor       *string
}

func GetIssues(first *int, after *string, filter IssueFilter) (*IssueConnection, error) {
    limit := 50
    if first != nil && *first < limit {
        limit = *first
    }

    query := db.Model(&Issue{})

    // 游标解码
    if after != nil {
        cursorID := decodeCursor(*after)
        query = query.Where("id > ?", cursorID)
    }

    // 应用过滤器
    query = applyFilter(query, filter)

    // 获取 limit + 1 条记录以判断 hasNextPage
    var issues []Issue
    query.Order("id ASC").Limit(limit + 1).Find(&issues)

    hasNextPage := len(issues) > limit
    if hasNextPage {
        issues = issues[:limit]
    }

    // 构建连接
    connection := &IssueConnection{
        TotalCount: getTotalCount(filter),
        PageInfo: PageInfo{
            HasNextPage: hasNextPage,
        },
    }

    for _, issue := range issues {
        connection.Edges = append(connection.Edges, IssueEdge{
            Node:   issue,
            Cursor: encodeCursor(issue.ID),
        })
    }

    if len(issues) > 0 {
        connection.PageInfo.StartCursor = &connection.Edges[0].Cursor
        connection.PageInfo.EndCursor = &connection.Edges[len(issues)-1].Cursor
    }

    return connection, nil
}
```

### 6.6 N+1 查询优化

使用 **DataLoader** 批量加载关联数据：

```go
// Issue DataLoader
type IssueLoader struct {
    loader *dataloader.Loader
}

func NewIssueLoader(db *gorm.DB) *IssueLoader {
    return &IssueLoader{
        loader: dataloader.NewBatchedLoader(func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
            ids := make([]string, len(keys))
            for i, key := range keys {
                ids[i] = key.String()
            }

            var issues []Issue
            db.Where("id IN ?", ids).Find(&issues)

            issueMap := make(map[string]*Issue)
            for i := range issues {
                issueMap[issues[i].ID] = &issues[i]
            }

            results := make([]*dataloader.Result, len(keys))
            for i, key := range keys {
                results[i] = &dataloader.Result{
                    Data: issueMap[key.String()],
                }
            }
            return results
        }),
    }
}
```

### 6.7 批量操作 API

```graphql
# 批量更新 Issue
mutation BatchUpdateIssues($ids: [ID!]!, $input: IssueUpdateInput!) {
  issueBatchUpdate(input: { ids: $ids, input: $input }) {
    success
    issues {
      id
      identifier
      status {
        name
      }
    }
    errors {
      message
      code
    }
  }
}
```

**服务端实现要点**：
1. **原子性**：使用数据库事务
2. **部分失败处理**：返回每个 Issue 的处理结果
3. **进度追踪**：长时间操作返回 Job ID

---

### 6.8 API 设计分析

#### GraphQL 的核心优势

- **按需查询**：前端精确指定所需字段，减少数据传输
- **游标分页（Cursor-based）**：比 offset 分页在大数据量下更稳定
- **Subscription**：原生支持实时推送（Issue 变更、评论通知）
- **批量操作**：`issueBatchUpdate` 原子事务
- **类型安全**：Schema 即文档，客户端自动类型推导

#### 关键设计要点

1. **游标分页**：应从 API 设计初期就实现游标分页，避免后期在大数据量下 offset 分页的性能问题
2. **DataLoader 模式**：解决 GraphQL 天然的 N+1 查询问题，通过批量加载关联数据保证性能
3. **批量操作原子性**：使用数据库事务保证批量更新的一致性，部分失败时返回每个 Issue 的处理结果
4. **Connection 类型统一**：所有列表查询使用 `*Connection` + `*Edge` + `PageInfo` 的统一分页模式，保持 API 一致性

---
