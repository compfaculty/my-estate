import type {
  HistoryItem,
  PageResponse,
  PropertyDetail,
  SearchParams,
  SourceItem,
} from "./types";

function qs(params: SearchParams): string {
  const u = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === "") return;
    u.set(k, String(v));
  });
  const s = u.toString();
  return s ? `?${s}` : "";
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(path);
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || res.statusText);
  }
  return res.json() as Promise<T>;
}

export function fetchProperties(params: SearchParams) {
  return get<PageResponse>(`/api/properties${qs(params)}`);
}

export function fetchProperty(id: string) {
  return get<PropertyDetail>(`/api/properties/${id}`);
}

export function fetchHistory(id: string) {
  return get<{ items: HistoryItem[] }>(`/api/properties/${id}/history`);
}

export function fetchSources() {
  return get<{ items: SourceItem[] }>("/api/sources");
}

export async function runCollect() {
  const res = await fetch("/api/dev/collect", { method: "POST" });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export function fetchHealth() {
  return get<{ status: string; time: string }>("/api/health");
}
