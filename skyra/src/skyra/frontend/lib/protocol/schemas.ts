/**
 * Zod schemas matching lib/protocol/types.ts (v.05 runtime).
 *
 * Every inbound WebSocket message is parsed through ServerMessageSchema before
 * it touches state. Anything that doesn't parse is logged as a drop and
 * thrown away — never silently passed through. This is how we catch backend
 * protocol drift the moment it happens.
 */

import { z } from "zod";

// ---------------------------------------------------------------------------
// Beings
// ---------------------------------------------------------------------------

export const ThoughtSnapshotSchema = z.object({
  peer: z.string(),
  thought: z.string(),
  ts: z.number(),
});

export const ThinkSnapshotSchema = z.object({
  budget: z.number(),
  operators: z.array(z.string()),
  history: z.array(ThoughtSnapshotSchema),
});

export const ActSnapshotSchema = z.object({
  operators: z.array(z.string()),
});

export const LayersSnapshotSchema = z.object({
  think: ThinkSnapshotSchema,
  act: ActSnapshotSchema,
});

export const TaskSnapshotSchema: z.ZodType<{
  name: string;
  state: string;
  description?: string;
  assumptions?: string[];
  commands?: string[];
  validation?: string;
  accepted_by?: string;
  items?: unknown[];
}> = z.lazy(() =>
  z.object({
    name: z.string(),
    description: z.string().optional(),
    assumptions: z.array(z.string()).optional(),
    commands: z.array(z.string()).optional(),
    validation: z.string().optional(),
    accepted_by: z.string().optional(),
    state: z.string(),
    items: z.array(TaskSnapshotSchema).optional(),
  }),
);

export const DeskSnapshotSchema = z.object({
  tasks: z.record(z.array(TaskSnapshotSchema)),
  views: z.record(z.string()),
});

export const LevelSnapshotSchema = z.object({
  xp: z.number(),
  level: z.number(),
  next: z.number(),
});

export const MemoryItemSchema = z.object({
  filename: z.string(),
  content: z.string(),
});

export const SkillItemSchema = z.object({
  name: z.string(),
  content: z.string(),
});

export const MemorySnapshotSchema = z.object({
  items: z.array(MemoryItemSchema),
  skills: z.array(SkillItemSchema),
});

export const BeingSnapshotSchema = z.object({
  name: z.string(),
  type: z.string(),
  identity: z.string(),
  purpose: z.string(),
  peers: z.array(z.string()),
  entrypoints: z.array(z.string()),
  status: z.string(),
  device: z.string(),
  layers: LayersSnapshotSchema.optional().nullable(),
  desk: DeskSnapshotSchema.optional().nullable(),
  memories: MemorySnapshotSchema,
  level: LevelSnapshotSchema.optional().nullable(),
});

// ---------------------------------------------------------------------------
// Threads
// ---------------------------------------------------------------------------

export const EdgeSnapshotSchema = z.object({
  from: z.string(),
  to: z.string(),
});

export const ThreadSnapshotSchema = z.object({
  id: z.string(),
  created_by: z.string(),
  active: z.boolean(),
  members: z.array(z.string()),
  edges: z.array(EdgeSnapshotSchema),
});

// ---------------------------------------------------------------------------
// Exchanges
// ---------------------------------------------------------------------------

export const EntrySnapshotSchema = z.object({
  index: z.number(),
  from: z.string(),
  content: z.string(),
  ts: z.number(),
});

export const ExchangeSnapshotSchema = z.object({
  key: z.string(),
  parties: z.tuple([z.string(), z.string()]),
  active: z.boolean(),
  entries: z.array(EntrySnapshotSchema),
  context: z.record(z.string()).optional(),
});

// ---------------------------------------------------------------------------
// Reality graph (recursive)
// ---------------------------------------------------------------------------

export interface RealityNodeZ {
  id: string;
  type: string;
  children: RealityNodeZ[];
}

export const RealityNodeSchema: z.ZodType<RealityNodeZ> = z.lazy(() =>
  z.object({
    id: z.string(),
    type: z.string(),
    children: z.array(RealityNodeSchema),
  }),
);

// ---------------------------------------------------------------------------
// Universe state
// ---------------------------------------------------------------------------

export const UniverseStateSchema = z.object({
  beings: z.array(BeingSnapshotSchema),
  threads: z.array(ThreadSnapshotSchema),
  exchanges: z.array(ExchangeSnapshotSchema),
  economics: z.record(z.number()),
  reality_graph: RealityNodeSchema,
});

// ---------------------------------------------------------------------------
// Server messages
// ---------------------------------------------------------------------------

export const ServerMessageSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("universe"),
    ts: z.number(),
    payload: UniverseStateSchema,
  }),
  z.object({
    type: z.literal("impulse"),
    ts: z.number(),
    payload: z.object({
      from: z.string(),
      content: z.string(),
    }),
  }),
]);

export type ServerMessageParsed = z.infer<typeof ServerMessageSchema>;
