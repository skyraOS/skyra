# Codex App Server endpoints

Date: 2026-09-19  
Protocol snapshot: installed `codex-cli 0.154.0`, including experimental schema entries.  
Scope: App Server JSON-RPC methods, not the CLI commands, model-native tools, or the separate MCP-server API.

## Sources and completeness

The inventories and parameter declarations below come directly from the installed binary's generated protocol bindings. They are version-specific, not a claim about every Codex release or every enabled feature. Schema presence does not guarantee runtime availability.

Reproduce the snapshot with:

```sh
codex --version
codex app-server generate-ts --experimental --out /tmp/codex-protocol-full
codex app-server generate-ts --out /tmp/codex-protocol-default
```

Source types: `ClientRequest.ts`, `ServerRequest.ts`, `ClientNotification.ts`, `ServerNotification.ts`, and `v2/*Params.ts`. An **E** marks a method or field absent from the default generation but present with `--experimental`; it is not a stability guarantee for unmarked entries.

Lifecycle behavior was checked against [official OpenAI App Server documentation](https://learn.chatgpt.com/docs/app-server). In this document, exact parameter availability follows the installed binary when the rolling documentation differs.

## The two entry points

| Intent | Direct method | Required parameters | Next operation |
|---|---|---|---|
| Start a new thread | `thread/start` | None; `{}` uses configured defaults | Use returned `thread.id` with `turn/start` |
| Continue a stored or detached thread | `thread/resume` | `threadId: string` | Use `turn/start` to send the next input |
| Send another prompt to an already loaded thread | `turn/start` | `threadId: string`, `input: UserInput[]` | Consume turn/item events |
| Add input to an active turn explicitly | `turn/steer` | See declaration below | Continue consuming that turn's events |

Thread creation/resumption and model execution are separate calls. Initialize the connection with `initialize`, await its response, then send the `initialized` notification before thread/turn requests. These sequences follow the [official lifecycle documentation](https://learn.chatgpt.com/docs/app-server#lifecycle-overview).

Minimal calls after initialization:

```json
{"id": 10, "method": "thread/start", "params": {}}
```

```json
{"id": 11, "method": "thread/resume", "params": {"threadId": "<existing-thread-id>"}}
```

After either entry point, send the prompt as a new turn:

```json
{"id": 12, "method": "turn/start", "params": {"threadId": "<thread-id>", "input": [{"type": "text", "text": "Your instruction here", "text_elements": []}]}}
```

These are JSON-RPC request IDs, not Skyra MoveIDs or committed RequestIDs. The provider's `threadId` and `turnId` also remain separate identities.

## Full parameters: thread/start

All fields are optional. Nullable types retain their generated spelling. Named types are defined in the generated bindings; selected nested types are printed below.

| Parameter | Type | Required | Experimental-only field |
|---|---|---|---|
| `model` | `string \| null` | No | — |
| `modelProvider` | `string \| null` | No | — |
| `allowProviderModelFallback` | `boolean` | No | E |
| `serviceTier` | `string \| null \| null` | No | — |
| `cwd` | `string \| null` | No | — |
| `runtimeWorkspaceRoots` | `Array<AbsolutePathBuf> \| null` | No | E |
| `approvalPolicy` | `AskForApproval \| null` | No | — |
| `approvalsReviewer` | `ApprovalsReviewer \| null` | No | — |
| `sandbox` | `SandboxMode \| null` | No | — |
| `permissions` | `string \| null` | No | E |
| `config` | `{ [key in string]?: JsonValue } \| null` | No | — |
| `serviceName` | `string \| null` | No | — |
| `baseInstructions` | `string \| null` | No | — |
| `developerInstructions` | `string \| null` | No | — |
| `personality` | `Personality \| null` | No | — |
| `multiAgentMode` | `MultiAgentMode \| null` | No | E |
| `ephemeral` | `boolean \| null` | No | — |
| `historyMode` | `ThreadHistoryMode \| null` | No | E |
| `sessionStartSource` | `ThreadStartSource \| null` | No | — |
| `threadSource` | `ThreadSource \| null` | No | — |
| `projectId` | `string \| null` | No | E |
| `environments` | `Array<TurnEnvironmentParams> \| null` | No | E |
| `dynamicTools` | `Array<DynamicToolSpec> \| null` | No | E |
| `selectedCapabilityRoots` | `Array<SelectedCapabilityRoot> \| null` | No | E |
| `mockExperimentalField` | `string \| null` | No | E |
| `experimentalRawEvents` | `boolean` | No | E |

`permissions` and `sandbox` are mutually exclusive. `runtimeWorkspaceRoots` paths must be absolute. `multiAgentMode` is deprecated and ignored in this snapshot. `mockExperimentalField` is test-only; `experimentalRawEvents` is described as internal-use. `dynamicTools` is offered at thread creation.

## Full parameters: thread/resume

| Parameter | Type | Required | Experimental-only field |
|---|---|---|---|
| `threadId` | `string` | Yes | — |
| `history` | `Array<ResponseItem> \| null` | No | E |
| `path` | `string \| null` | No | E |
| `model` | `string \| null` | No | — |
| `modelProvider` | `string \| null` | No | — |
| `serviceTier` | `string \| null \| null` | No | — |
| `cwd` | `string \| null` | No | — |
| `runtimeWorkspaceRoots` | `Array<AbsolutePathBuf> \| null` | No | E |
| `approvalPolicy` | `AskForApproval \| null` | No | — |
| `approvalsReviewer` | `ApprovalsReviewer \| null` | No | — |
| `sandbox` | `SandboxMode \| null` | No | — |
| `permissions` | `string \| null` | No | E |
| `config` | `{ [key in string]?: JsonValue } \| null` | No | — |
| `baseInstructions` | `string \| null` | No | — |
| `developerInstructions` | `string \| null` | No | — |
| `personality` | `Personality \| null` | No | — |
| `excludeTurns` | `boolean` | No | — |
| `initialTurnsPage` | `ThreadResumeInitialTurnsPageParams \| null` | No | E |

Prefer `threadId`. For non-running threads, alternate source precedence is `history` > nonempty `path` > `threadId`; the `threadId` field remains required by the type. `history` is marked unstable/internal, and `path` unstable. For an already running thread, a nonempty path must match its active rollout. `permissions` and `sandbox` are mutually exclusive. This version has no `dynamicTools` parameter on resume.

## Full parameters: turn/start

| Parameter | Type | Required | Experimental-only field |
|---|---|---|---|
| `threadId` | `string` | Yes | — |
| `clientUserMessageId` | `string \| null` | No | — |
| `input` | `Array<UserInput>` | Yes | — |
| `turnTrigger` | `string \| null` | No | — |
| `toolOutput` | `TurnToolOutput \| null` | No | — |
| `responsesapiClientMetadata` | `{ [key in string]?: string } \| null` | No | E |
| `additionalContext` | `{ [key in string]?: AdditionalContextEntry } \| null` | No | E |
| `environments` | `Array<TurnEnvironmentParams> \| null` | No | E |
| `cwd` | `string \| null` | No | — |
| `runtimeWorkspaceRoots` | `Array<AbsolutePathBuf> \| null` | No | E |
| `approvalPolicy` | `AskForApproval \| null` | No | — |
| `approvalsReviewer` | `ApprovalsReviewer \| null` | No | — |
| `sandboxPolicy` | `SandboxPolicy \| null` | No | — |
| `permissions` | `string \| null` | No | E |
| `model` | `string \| null` | No | — |
| `serviceTier` | `string \| null \| null` | No | — |
| `serviceTierForTurn` | `string \| null` | No | — |
| `effort` | `ReasoningEffort \| null` | No | — |
| `summary` | `ReasoningSummary \| null` | No | — |
| `personality` | `Personality \| null` | No | — |
| `outputSchema` | `JsonValue \| null` | No | — |
| `collaborationMode` | `CollaborationMode \| null` | No | E |
| `multiAgentMode` | `MultiAgentMode \| null` | No | E |
| `cyberAccessProgram` | `CyberAccessProgram \| null` | No | E |

`permissions` and `sandboxPolicy` are mutually exclusive. `outputSchema` constrains the final response. `input` is structurally required even when other input mechanisms are present. Completion is asynchronous; the initial RPC response does not itself contain the final answer.

## Selected nested types and active-turn steering

These declarations are copied from generated local types with comments/imports removed. Referenced type names retain their protocol meanings.

```ts
type TurnSteerParams = { threadId: string, clientUserMessageId?: string | null, input: Array<UserInput>, responsesapiClientMetadata?: { [key in string]?: string } | null, additionalContext?: { [key in string]?: AdditionalContextEntry } | null, expectedTurnId: string, }; 
```

```ts
type ThreadResumeInitialTurnsPageParams = { limit?: number | null, sortDirection?: SortDirection | null, itemsView?: TurnItemsView | null, }; 
```

```ts
type DynamicToolSpec = { "type": "function" } & DynamicToolFunctionSpec | { "type": "namespace" } & DynamicToolNamespaceSpec; 
```

```ts
type DynamicToolFunctionSpec = { name: string, description: string, inputSchema: JsonValue, deferLoading?: boolean, }; 
```

```ts
type DynamicToolNamespaceSpec = { name: string, description: string, tools: Array<DynamicToolNamespaceTool>, }; 
```

```ts
type UserInput = { "type": "text", text: string, text_elements: Array<TextElement>, } | { "type": "image", detail?: ImageDetail, url: string, } | { "type": "localImage", detail?: ImageDetail, path: string, } | { "type": "audio", url: string, } | { "type": "localAudio", path: string, } | { "type": "skill", name: string, path: string, } | { "type": "mention", name: string, path: string, }; 
```

```ts
type AskForApproval = "untrusted" | "on-request" | { "granular": { sandbox_approval: boolean, rules: boolean, skill_approval: boolean, request_permissions: boolean, mcp_elicitations: boolean, } } | "never"; 
```

```ts
type ApprovalsReviewer = "user" | "auto_review" | "guardian_subagent"; 
```

```ts
type SandboxMode = "read-only" | "workspace-write" | "danger-full-access"; 
```

```ts
type Personality = "none" | "friendly" | "pragmatic"; 
```

```ts
type ThreadHistoryMode = "legacy" | "paginated"; 
```

```ts
type TurnEnvironmentParams = { environmentId: string, cwd: LegacyAppPathString, runtimeWorkspaceRoots?: Array<LegacyAppPathString> | null, }; 
```

## All client-to-server request endpoints (162)

These are the methods Skyra can call on App Server. **E** means experimental-only in schema generation.

| Method | Parameters type | E |
|---|---|---|
| `initialize` | `InitializeParams` | — |
| `server/diagnostics` | `ServerDiagnosticsParams` | E |
| `userVerification/status` | `UserVerificationStatusParams` | E |
| `userVerification/enroll` | `UserVerificationEnrollParams` | E |
| `userVerification/delete` | `UserVerificationDeleteParams` | E |
| `userVerification/verify` | `UserVerificationVerifyParams` | E |
| `thread/start` | `ThreadStartParams` | — |
| `thread/resume` | `ThreadResumeParams` | — |
| `thread/fork` | `ThreadForkParams` | — |
| `thread/archive` | `ThreadArchiveParams` | — |
| `thread/delete` | `ThreadDeleteParams` | — |
| `thread/unsubscribe` | `ThreadUnsubscribeParams` | — |
| `thread/increment_elicitation` | `ThreadIncrementElicitationParams` | E |
| `thread/decrement_elicitation` | `ThreadDecrementElicitationParams` | E |
| `thread/name/set` | `ThreadSetNameParams` | — |
| `thread/goal/set` | `ThreadGoalSetParams` | — |
| `thread/goal/get` | `ThreadGoalGetParams` | — |
| `thread/goal/clear` | `ThreadGoalClearParams` | — |
| `thread/queue/add` | `ThreadQueueAddParams` | E |
| `thread/queue/list` | `ThreadQueueListParams` | E |
| `thread/queue/update` | `ThreadQueueUpdateParams` | E |
| `thread/queue/delete` | `ThreadQueueDeleteParams` | E |
| `thread/queue/reorder` | `ThreadQueueReorderParams` | E |
| `thread/queue/start` | `ThreadQueueStartParams` | E |
| `thread/metadata/update` | `ThreadMetadataUpdateParams` | — |
| `thread/section/move` | `ThreadSectionMoveParams` | — |
| `thread/settings/update` | `ThreadSettingsUpdateParams` | E |
| `thread/memoryMode/set` | `ThreadMemoryModeSetParams` | E |
| `memory/reset` | `undefined` | E |
| `thread/unarchive` | `ThreadUnarchiveParams` | — |
| `thread/compact/start` | `ThreadCompactStartParams` | — |
| `thread/shellCommand` | `ThreadShellCommandParams` | — |
| `thread/approveGuardianDeniedAction` | `ThreadApproveGuardianDeniedActionParams` | — |
| `thread/backgroundTerminals/clean` | `ThreadBackgroundTerminalsCleanParams` | E |
| `thread/backgroundTerminals/list` | `ThreadBackgroundTerminalsListParams` | E |
| `thread/backgroundTerminals/terminate` | `ThreadBackgroundTerminalsTerminateParams` | E |
| `thread/rollback` | `ThreadRollbackParams` | — |
| `thread/revert` | `ThreadRevertParams` | — |
| `thread/list` | `ThreadListParams` | — |
| `project/list` | `ProjectListParams` | E |
| `project/read` | `ProjectReadParams` | E |
| `project/create` | `ProjectCreateParams` | E |
| `project/import` | `ProjectImportParams` | E |
| `project/update` | `ProjectUpdateParams` | E |
| `project/move` | `ProjectMoveParams` | E |
| `project/delete` | `ProjectDeleteParams` | E |
| `threadSection/list` | `ThreadSectionListParams` | — |
| `threadSection/create` | `ThreadSectionCreateParams` | — |
| `threadSection/update` | `ThreadSectionUpdateParams` | — |
| `threadSection/delete` | `ThreadSectionDeleteParams` | — |
| `thread/search` | `ThreadSearchParams` | E |
| `thread/searchOccurrences` | `ThreadSearchOccurrencesParams` | E |
| `thread/loaded/list` | `ThreadLoadedListParams` | — |
| `thread/read` | `ThreadReadParams` | — |
| `thread/turns/list` | `ThreadTurnsListParams` | — |
| `thread/items/list` | `ThreadItemsListParams` | — |
| `thread/inject_items` | `ThreadInjectItemsParams` | — |
| `skills/list` | `SkillsListParams` | — |
| `skills/extraRoots/set` | `SkillsExtraRootsSetParams` | — |
| `hooks/list` | `HooksListParams` | — |
| `marketplace/add` | `MarketplaceAddParams` | — |
| `marketplace/remove` | `MarketplaceRemoveParams` | — |
| `marketplace/upgrade` | `MarketplaceUpgradeParams` | — |
| `plugin/list` | `PluginListParams` | — |
| `plugin/search` | `PluginSearchParams` | E |
| `plugin/installed` | `PluginInstalledParams` | — |
| `plugin/reconcile` | `PluginReconcileParams` | — |
| `plugin/read` | `PluginReadParams` | — |
| `plugin/skill/read` | `PluginSkillReadParams` | — |
| `plugin/share/save` | `PluginShareSaveParams` | — |
| `plugin/share/updateTargets` | `PluginShareUpdateTargetsParams` | — |
| `plugin/share/list` | `PluginShareListParams` | — |
| `plugin/share/checkout` | `PluginShareCheckoutParams` | — |
| `plugin/share/delete` | `PluginShareDeleteParams` | — |
| `app/read` | `AppsReadParams` | — |
| `app/list` | `AppsListParams` | — |
| `app/installed` | `AppsInstalledParams` | — |
| `fs/readFile` | `FsReadFileParams` | — |
| `fs/writeFile` | `FsWriteFileParams` | — |
| `fs/createDirectory` | `FsCreateDirectoryParams` | — |
| `fs/getMetadata` | `FsGetMetadataParams` | — |
| `fs/readDirectory` | `FsReadDirectoryParams` | — |
| `fs/remove` | `FsRemoveParams` | — |
| `fs/copy` | `FsCopyParams` | — |
| `fs/watch` | `FsWatchParams` | — |
| `fs/unwatch` | `FsUnwatchParams` | — |
| `skills/config/write` | `SkillsConfigWriteParams` | — |
| `plugin/install` | `PluginInstallParams` | — |
| `plugin/uninstall` | `PluginUninstallParams` | — |
| `turn/start` | `TurnStartParams` | — |
| `turn/settings/update` | `TurnSettingsUpdateParams` | E |
| `turn/steer` | `TurnSteerParams` | — |
| `turn/interrupt` | `TurnInterruptParams` | — |
| `thread/realtime/start` | `ThreadRealtimeStartParams` | E |
| `thread/realtime/appendAudio` | `ThreadRealtimeAppendAudioParams` | E |
| `thread/realtime/appendText` | `ThreadRealtimeAppendTextParams` | E |
| `thread/realtime/appendSpeech` | `ThreadRealtimeAppendSpeechParams` | E |
| `thread/realtime/stop` | `ThreadRealtimeStopParams` | E |
| `thread/timeline/list` | `ThreadTimelineListParams` | E |
| `thread/realtime/listVoices` | `ThreadRealtimeListVoicesParams` | E |
| `review/start` | `ReviewStartParams` | — |
| `model/list` | `ModelListParams` | — |
| `modelProvider/capabilities/read` | `ModelProviderCapabilitiesReadParams` | — |
| `experimentalFeature/list` | `ExperimentalFeatureListParams` | — |
| `permissionProfile/list` | `PermissionProfileListParams` | — |
| `experimentalFeature/enablement/set` | `ExperimentalFeatureEnablementSetParams` | — |
| `remoteControl/enable` | `RemoteControlEnableParams \| null` | E |
| `remoteControl/disable` | `RemoteControlDisableParams \| null` | E |
| `remoteControl/status/read` | `undefined` | E |
| `remoteControl/pairing/start` | `RemoteControlPairingStartParams` | E |
| `remoteControl/pairing/status` | `RemoteControlPairingStatusParams` | E |
| `remoteControl/client/list` | `RemoteControlClientsListParams` | E |
| `remoteControl/client/revoke` | `RemoteControlClientsRevokeParams` | E |
| `collaborationMode/list` | `CollaborationModeListParams` | E |
| `mock/experimentalMethod` | `MockExperimentalMethodParams` | E |
| `environment/add` | `EnvironmentAddParams` | E |
| `environment/info` | `EnvironmentInfoParams` | E |
| `environment/status` | `EnvironmentStatusParams` | E |
| `mcpServer/oauth/login` | `McpServerOauthLoginParams` | — |
| `config/mcpServer/reload` | `undefined` | — |
| `mcpServerStatus/list` | `ListMcpServerStatusParams` | — |
| `mcpServer/resource/read` | `McpResourceReadParams` | — |
| `mcpServer/event/stream/start` | `McpServerEventStreamStartParams` | E |
| `mcpServer/event/stream/stop` | `McpServerEventStreamStopParams` | E |
| `mcpServer/tool/call` | `McpServerToolCallParams` | — |
| `windowsSandbox/setupStart` | `WindowsSandboxSetupStartParams` | — |
| `windowsSandbox/readiness` | `undefined` | — |
| `account/login/start` | `LoginAccountParams` | — |
| `account/bedrock/discover` | `BedrockDiscoverParams` | E |
| `account/bedrock/setup` | `BedrockSetupParams` | E |
| `account/login/cancel` | `CancelLoginAccountParams` | — |
| `account/logout` | `undefined` | — |
| `account/rateLimits/read` | `GetAccountRateLimitsParams \| undefined` | — |
| `account/rateLimitResetCredit/consume` | `ConsumeAccountRateLimitResetCreditParams` | — |
| `account/usage/read` | `GetAccountTokenUsageParams \| undefined` | — |
| `account/workspaceMessages/read` | `undefined` | — |
| `account/sendAddCreditsNudgeEmail` | `SendAddCreditsNudgeEmailParams` | — |
| `feedback/upload` | `FeedbackUploadParams` | — |
| `command/exec` | `CommandExecParams` | — |
| `command/exec/write` | `CommandExecWriteParams` | — |
| `command/exec/terminate` | `CommandExecTerminateParams` | — |
| `command/exec/resize` | `CommandExecResizeParams` | — |
| `process/spawn` | `ProcessSpawnParams` | E |
| `process/writeStdin` | `ProcessWriteStdinParams` | E |
| `process/kill` | `ProcessKillParams` | E |
| `process/resizePty` | `ProcessResizePtyParams` | E |
| `config/read` | `ConfigReadParams` | — |
| `externalAgentConfig/detect` | `ExternalAgentConfigDetectParams` | — |
| `externalAgentConfig/import` | `ExternalAgentConfigImportParams` | — |
| `externalAgentConfig/import/recordHistory` | `ExternalAgentConfigImportHistoryRecordParams` | — |
| `externalAgentConfig/import/readHistories` | `undefined` | — |
| `config/value/write` | `ConfigValueWriteParams` | — |
| `config/batchWrite` | `ConfigBatchWriteParams` | — |
| `configRequirements/read` | `undefined` | — |
| `account/read` | `GetAccountParams` | — |
| `getConversationSummary` | `GetConversationSummaryParams` | — |
| `gitDiffToRemote` | `GitDiffToRemoteParams` | — |
| `getAuthStatus` | `GetAuthStatusParams` | — |
| `fuzzyFileSearch` | `FuzzyFileSearchParams` | — |
| `fuzzyFileSearch/sessionStart` | `FuzzyFileSearchSessionStartParams` | E |
| `fuzzyFileSearch/sessionUpdate` | `FuzzyFileSearchSessionUpdateParams` | E |
| `fuzzyFileSearch/sessionStop` | `FuzzyFileSearchSessionStopParams` | E |

## Server-to-client request endpoints (11)

These run in the reverse direction: Codex asks the host to handle them. Reply using the received JSON-RPC ID; they are not client calls into Codex.

- `item/commandExecution/requestApproval`
- `item/fileChange/requestApproval`
- `item/tool/requestUserInput`
- `mcpServer/elicitation/request`
- `item/permissions/requestApproval`
- `item/tool/call`
- `account/chatgptAuthTokens/refresh`
- `attestation/generate`
- `currentTime/read`
- `applyPatchApproval`
- `execCommandApproval`

For a dynamic tool call, `item/tool/call` requests host execution. The matching JSON-RPC response supplies the tool result; it is not `thread/resume`. This is a separate kind of continuation from reopening a thread.

## Client-to-server notifications (1)

Notifications have no request ID and are not callable request endpoints. Included here to make the protocol inventory complete.

- `initialized`

## Server-to-client notifications (83)

Notifications have no request ID and are not callable request endpoints. Included here to make the protocol inventory complete.

- `error`
- `thread/started`
- `thread/status/changed`
- `thread/archived`
- `thread/deleted`
- `thread/unarchived`
- `thread/closed`
- `thread/reverted`
- `skills/changed`
- `thread/name/updated`
- `thread/goal/updated`
- `thread/goal/cleared`
- `thread/queue/changed`
- `project/changed`
- `thread/project/updated`
- `thread/environment/connected`
- `thread/environment/disconnected`
- `thread/settings/updated`
- `thread/tokenUsage/updated`
- `turn/started`
- `hook/started`
- `turn/completed`
- `hook/completed`
- `turn/diff/updated`
- `turn/plan/updated`
- `item/started`
- `item/autoApprovalReview/started`
- `item/autoApprovalReview/completed`
- `autoApprovalReview/strictReviewRequired`
- `item/completed`
- `rawResponseItem/completed`
- `rawResponse/completed`
- `item/agentMessage/delta`
- `item/plan/delta`
- `command/exec/outputDelta`
- `process/outputDelta`
- `process/exited`
- `item/commandExecution/outputDelta`
- `item/commandExecution/terminalInteraction`
- `item/fileChange/outputDelta`
- `item/fileChange/patchUpdated`
- `serverRequest/resolved`
- `item/mcpToolCall/progress`
- `mcpServer/oauthLogin/completed`
- `mcpServer/startupStatus/updated`
- `mcpServer/event/stream/notification`
- `account/updated`
- `account/rateLimits/updated`
- `app/list/updated`
- `remoteControl/status/changed`
- `externalAgentConfig/import/progress`
- `externalAgentConfig/import/completed`
- `fs/changed`
- `item/reasoning/summaryTextDelta`
- `item/reasoning/summaryPartAdded`
- `item/reasoning/textDelta`
- `thread/compacted`
- `model/rerouted`
- `model/verification`
- `modelProvider/authRecoveryStarted`
- `modelProvider/authRecoveryCompleted`
- `turn/moderationMetadata`
- `model/safetyBuffering/updated`
- `warning`
- `guardianWarning`
- `deprecationNotice`
- `configWarning`
- `fuzzyFileSearch/sessionUpdated`
- `fuzzyFileSearch/sessionCompleted`
- `thread/realtime/started`
- `thread/realtime/itemAdded`
- `thread/realtime/item/started`
- `thread/realtime/item/transcript/delta`
- `thread/realtime/item/completed`
- `thread/realtime/transcript/delta`
- `thread/realtime/transcript/done`
- `thread/realtime/outputAudio/delta`
- `thread/realtime/sdp`
- `thread/realtime/error`
- `thread/realtime/closed`
- `windows/worldWritableWarning`
- `windowsSandbox/setupCompleted`
- `account/login/completed`

## Fit with the Surface discussion

A higher-level "start conversation" Surface can compose the `thread/start` and `turn/start` leaf endpoints. A "continue conversation" Surface can select `thread/resume` when the thread must be rejoined, then `turn/start`. This is a proposed Skyra composition, not a new provider endpoint.

Ingress can resolve the stored provider thread identity and prepare the exact endpoint parameters under its profile. Egress can correlate RPC responses and asynchronous events under a separate profile turn. The provider thread/turn/request identities belong in that correlation state; they should not be conflated with SurfaceID, MoveID, or Skyra RequestID.

Endpoint output includes both responses and notifications, plus server-initiated requests. The Surface design must account for those separate directions before treating a provider interaction as a single input/output call.
