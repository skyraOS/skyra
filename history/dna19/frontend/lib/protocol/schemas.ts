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

export const EntryRoleSchema = z.enum(["user", "agent"]);

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
});

export const ServerMessageSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("snapshot"),
    uiIncarnation: z.string(),
    revision: z.number(),
    conversations: z.array(ConversationViewSchema),
    requests: z.array(RequestRecordSchema),
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
