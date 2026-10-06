import type { RequestRecord } from "./protocol/types";

export interface PendingSubmission {
  requestId: string;
  text: string;
  revision: number;
  incarnation: string;
}

/** Admission clears the matching draft; absence and unknown outcomes retain it. */
export function deliveryState(pending: PendingSubmission, record: RequestRecord | undefined, revision: number, incarnation: string | null): "waiting" | "accepted" | "retained" {
  if (!record) return revision > pending.revision || incarnation !== pending.incarnation ? "retained" : "waiting";
  return record.status === "rejected" || record.code === "delivery-unknown" ? "retained" : "accepted";
}
