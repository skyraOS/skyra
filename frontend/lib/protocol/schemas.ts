/**
 * Zod schemas for the DNA-19 user-executable UI protocol (types.ts).
 *
 * Every inbound message is parsed through ServerMessageSchema before it
 * touches state. Anything that doesn't parse is logged as a drop and thrown
 * away — never silently passed through. This is how backend protocol drift
 * is caught the moment it happens.
 */

import { z } from "zod";

export const ConversationPhaseSchema = z.enum([
  "ready",
  "pending",
  "failed",
  "closed",
]);

export const EntryRoleSchema = z.enum(["user", "agent", "call"]);

export const CallSchema = z.object({
  to: z.string(),
  arg: z.string(),
  state: z.enum(["running", "returned", "fault"]),
  result: z.string().optional(),
  image: z.string().max(3 << 20).regex(/^data:image\/(png|jpeg);base64,[A-Za-z0-9+/=]+$/).optional(),
  fault: z.string().optional(),
  cause: z.string().optional(),
  stamp: z.string(),
  ms: z.number().optional(),
});

export const RequestStatusSchema = z.enum([
  "accepted",
  "returned",
  "failed",
  "rejected",
]);

export const ConversationEntrySchema = z.object({
  role: EntryRoleSchema,
  text: z.string(),
  ts: z.number(),
  requestId: z.string().nullable(),
  call: CallSchema.optional(),
});

export const RequestRecordSchema = z.object({
  requestId: z.string(),
  status: RequestStatusSchema,
  code: z.string().optional(),
  message: z.string().optional(),
  ts: z.number(),
});

export const ConversationViewSchema = z.object({
  conversation: z.string(),
  phase: ConversationPhaseSchema,
  entries: z.array(ConversationEntrySchema),
  pendingRequestId: z.string().nullable(),
  procedure: z.string(),
});

export const OutlineNodeSchema = z.object({
  name: z.string(),
  calls: z.string(),
  after: z.array(z.string()),
  narrows: z.object({ actor: z.string(), reach: z.array(z.string()) }).optional(),
});

export const DestinationSchema = z.object({
  actor: z.string(),
  title: z.string(),
  schema: z.record(z.string(), z.unknown()),
  outline: z.object({ nodes: z.array(OutlineNodeSchema) }),
});

const TreeNodeSchema = z.object({
  id: z.string(),
  parent: z.string().optional(),
  kind: z.string(),
  phase: z.string(),
  depth: z.number(),
  turns: z.number(),
  holds: z.array(z.string()).nullable(),
  names: z.number(),
  returned: z.number(),
  why: z.string().optional(),
  lost: z.array(z.string()).optional(),
});

const TreeReplaySchema = z.object({
  frame: z.string(),
  goal: z.string(),
  kind: z.string(),
  reach: z.array(z.string()),
  frames: z.array(z.object({
    event: z.string(),
    agent: z.string().optional(),
    detail: z.string().optional(),
    nodes: z.array(TreeNodeSchema),
  })),
});

export const TreeRunSchema = z.object({
  id: z.string(),
  goal: z.string(),
  kind: z.string(),
  reach: z.array(z.string()),
  phase: z.enum(["running", "done", "failed"]),
  answer: z.string().optional(),
  error: z.string().optional(),
  started: z.number(),
  view: z.array(TreeReplaySchema).nullable().optional(),
});

export const ServerMessageSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("snapshot"),
    uiIncarnation: z.string(),
    revision: z.number(),
    conversations: z.array(ConversationViewSchema),
    requests: z.array(RequestRecordSchema),
    destinations: z.array(DestinationSchema),
  }),
  z.object({
    type: z.literal("accepted"),
    requestId: z.string(),
  }),
  z.object({
    type: z.literal("rejected"),
    requestId: z.string(),
    code: z.string(),
    message: z.string(),
  }),
  z.object({
    type: z.literal("trees"),
    targets: z.array(z.string()),
    trees: z.array(TreeRunSchema),
  }),
]);

export type ServerMessageParsed = z.infer<typeof ServerMessageSchema>;
