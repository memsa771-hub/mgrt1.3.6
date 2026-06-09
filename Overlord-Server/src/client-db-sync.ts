import { upsertClientRows, type ClientDbRow } from "./db";
import { metrics } from "./metrics";

export const CLIENT_DB_SYNC_INTERVAL_MS = Number(process.env.OVERLORD_CLIENT_DB_SYNC_MS || 5000);
export const CLIENT_DB_SYNC_BATCH_SIZE = Math.max(
  1,
  Number(process.env.OVERLORD_CLIENT_DB_SYNC_BATCH_SIZE || 100),
);

const lastClientDbSync = new Map<string, number>();
const pendingClientDbUpdates = new Map<string, ClientDbRow>();
let flushSoonTimer: ReturnType<typeof setTimeout> | null = null;

export function queueClientDbUpdate(partial: ClientDbRow): void {
  const existing = pendingClientDbUpdates.get(partial.id);
  if (!existing) {
    pendingClientDbUpdates.set(partial.id, { ...partial });
    return;
  }
  pendingClientDbUpdates.set(partial.id, {
    ...existing,
    ...partial,
    id: partial.id,
    lastSeen: partial.lastSeen ?? existing.lastSeen,
    online: partial.online ?? existing.online,
    pingMs: partial.pingMs ?? existing.pingMs,
  });
}

export function flushQueuedClientDbUpdates(): void {
  if (pendingClientDbUpdates.size === 0) return;

  const startedAt = Date.now();
  let processed = 0;
  const updates: ClientDbRow[] = [];
  for (const [clientId, update] of pendingClientDbUpdates.entries()) {
    updates.push(update);
    pendingClientDbUpdates.delete(clientId);
    processed += 1;
    if (processed >= CLIENT_DB_SYNC_BATCH_SIZE) {
      break;
    }
  }
  upsertClientRows(updates);

  if (pendingClientDbUpdates.size > 0 && !flushSoonTimer) {
    flushSoonTimer = setTimeout(() => {
      flushSoonTimer = null;
      flushQueuedClientDbUpdates();
    }, 25);
  }
  metrics.recordInternalTask("client-db-flush", Date.now() - startedAt);
}

setInterval(flushQueuedClientDbUpdates, CLIENT_DB_SYNC_INTERVAL_MS);

export function shouldSyncClientToDb(clientId: string, now: number): boolean {
  const last = lastClientDbSync.get(clientId) || 0;
  if (now - last < CLIENT_DB_SYNC_INTERVAL_MS) return false;
  lastClientDbSync.set(clientId, now);
  return true;
}

export function markClientDbSynced(clientId: string, now: number): void {
  lastClientDbSync.set(clientId, now);
}

export function clearClientSyncState(clientId: string): void {
  lastClientDbSync.delete(clientId);
  pendingClientDbUpdates.delete(clientId);
}

export function getClientDbSyncStats(): {
  trackedClients: number;
  pendingUpdates: number;
  flushScheduled: boolean;
} {
  return {
    trackedClients: lastClientDbSync.size,
    pendingUpdates: pendingClientDbUpdates.size,
    flushScheduled: flushSoonTimer !== null,
  };
}
