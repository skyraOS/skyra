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
]);

export type ServerMessageParsed = z.infer<typeof ServerMessageSchema>;
