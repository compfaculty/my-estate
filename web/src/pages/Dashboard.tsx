import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { fetchProperties, fetchSources, runCollect } from "../api";
import { formatPrice, formatWhen } from "../format";

export function Dashboard() {
  const qc = useQueryClient();
  const sources = useQuery({ queryKey: ["sources"], queryFn: fetchSources });
  const recent = useQuery({
    queryKey: ["recent"],
    queryFn: () => fetchProperties({ limit: 8, offset: 0 }),
  });
  const collect = useMutation({
    mutationFn: runCollect,
    onSuccess: () => {
      void qc.invalidateQueries();
    },
  });

  return (
    <div className="page">
      <h1>Market feed</h1>
      <p className="lede">
        Local PostgreSQL is the source of truth. Collectors write listings; this terminal only reads
        the warehouse.
      </p>

      <div className="grid-cards" style={{ marginBottom: 22 }}>
        <div className="stat">
          <div className="k">Indexed properties</div>
          <div className="v mono">{recent.data?.total ?? "—"}</div>
        </div>
        <div className="stat">
          <div className="k">Sources</div>
          <div className="v mono">{sources.data?.items.length ?? "—"}</div>
        </div>
        <div className="stat">
          <div className="k">Last crawl</div>
          <div className="v" style={{ fontSize: 16 }}>
            {formatWhen(sources.data?.items[0]?.last_crawl_at)}
          </div>
        </div>
        <div className="stat">
          <div className="k">Ingest</div>
          <div className="v">
            <button className="btn" disabled={collect.isPending} onClick={() => collect.mutate()}>
              {collect.isPending ? "Running…" : "Run mock collector"}
            </button>
          </div>
        </div>
      </div>

      <div className="section">
        <h3>Source health</h3>
        {sources.isError && <p className="error">API unavailable. Start the Go server on :8080.</p>}
        <table>
          <thead>
            <tr>
              <th>Source</th>
              <th>Mode</th>
              <th>Status</th>
              <th>Last error</th>
              <th>Last success</th>
              <th>New</th>
              <th>Updated</th>
            </tr>
          </thead>
          <tbody>
            {(sources.data?.items ?? []).map((s) => (
              <tr key={s.id}>
                <td>{s.name}</td>
                <td className="mono">{s.fetch_mode}</td>
                <td>
                  <span className={`badge ${s.last_run_status === "success" ? "ok" : ""}`}>
                    {s.last_run_status || (s.enabled ? "idle" : "disabled")}
                  </span>
                </td>
                <td className="mono" title={s.last_error ?? ""}>
                  {s.last_error ? s.last_error.slice(0, 48) : "—"}
                </td>
                <td>{formatWhen(s.last_success_at)}</td>
                <td className="mono">{s.listings_discovered}</td>
                <td className="mono">{s.listings_updated}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="section">
        <h3>Recently added</h3>
        <table>
          <thead>
            <tr>
              <th></th>
              <th>Address</th>
              <th>Price</th>
              <th>₴/m²</th>
              <th>Rooms</th>
              <th>Source</th>
              <th>First seen</th>
            </tr>
          </thead>
          <tbody>
            {(recent.data?.items ?? []).map((p) => (
              <tr key={p.id} className="clickable" onClick={() => (window.location.href = `/property/${p.id}`)}>
                <td>{p.photo_url ? <img className="thumb" src={p.photo_url} alt="" /> : null}</td>
                <td>
                  <Link to={`/property/${p.id}`}>{p.address}</Link>
                </td>
                <td className="mono">{formatPrice(p.price_uah ?? p.asking_price, p.price_usd)}</td>
                <td className="mono">{p.price_per_m2 ? Math.round(p.price_per_m2).toLocaleString("uk-UA") : "—"}</td>
                <td className="mono">{p.rooms ?? "—"}</td>
                <td>{p.source}</td>
                <td>{formatWhen(p.first_seen_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
