import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { fetchProperties } from "../api";
import { formatPrice, formatWhen } from "../format";
import type { SearchParams } from "../types";

const empty: SearchParams = {
  city: "Київ",
  neighborhood: "",
  q: "",
  min_price: "",
  max_price: "",
  min_area: "",
  max_area: "",
  min_rooms: "",
  max_rooms: "",
  source: "",
};

export function Search() {
  const [draft, setDraft] = useState<SearchParams>(empty);
  const [applied, setApplied] = useState<SearchParams>({ ...empty, limit: 20, offset: 0 });
  const [offset, setOffset] = useState(0);

  const query = useQuery({
    queryKey: ["search", applied, offset],
    queryFn: () => fetchProperties({ ...applied, limit: 20, offset }),
  });

  const set = (k: keyof SearchParams, v: string) => setDraft((d) => ({ ...d, [k]: v }));

  return (
    <div className="page page-wide">
      <h1>Property search</h1>
      <p className="lede">Deterministic filters against PostgreSQL. No ranking agents in this phase.</p>
      <div className="search-layout">
        <form
          className="filters panel"
          onSubmit={(e) => {
            e.preventDefault();
            setOffset(0);
            setApplied({ ...draft });
          }}
        >
          <h2>Filters</h2>
          <div className="field">
            <label>City</label>
            <input value={draft.city ?? ""} onChange={(e) => set("city", e.target.value)} />
          </div>
          <div className="field">
            <label>Neighborhood</label>
            <input value={draft.neighborhood ?? ""} onChange={(e) => set("neighborhood", e.target.value)} />
          </div>
          <div className="field">
            <label>Free text</label>
            <input value={draft.q ?? ""} onChange={(e) => set("q", e.target.value)} />
          </div>
          <div className="field">
            <label>Min price (UAH)</label>
            <input value={draft.min_price ?? ""} onChange={(e) => set("min_price", e.target.value)} />
          </div>
          <div className="field">
            <label>Max price (UAH)</label>
            <input value={draft.max_price ?? ""} onChange={(e) => set("max_price", e.target.value)} />
          </div>
          <div className="field">
            <label>Min area m²</label>
            <input value={draft.min_area ?? ""} onChange={(e) => set("min_area", e.target.value)} />
          </div>
          <div className="field">
            <label>Max area m²</label>
            <input value={draft.max_area ?? ""} onChange={(e) => set("max_area", e.target.value)} />
          </div>
          <div className="field">
            <label>Min rooms</label>
            <input value={draft.min_rooms ?? ""} onChange={(e) => set("min_rooms", e.target.value)} />
          </div>
          <div className="field">
            <label>Max rooms</label>
            <input value={draft.max_rooms ?? ""} onChange={(e) => set("max_rooms", e.target.value)} />
          </div>
          <div className="field">
            <label>Source</label>
            <input value={draft.source ?? ""} onChange={(e) => set("source", e.target.value)} placeholder="mock" />
          </div>
          <label className="check">
            <input
              type="checkbox"
              checked={draft.elevator === "true"}
              onChange={(e) => set("elevator", e.target.checked ? "true" : "")}
            />
            Elevator
          </label>
          <label className="check">
            <input
              type="checkbox"
              checked={draft.parking === "true"}
              onChange={(e) => set("parking", e.target.checked ? "true" : "")}
            />
            Parking
          </label>
          <label className="check">
            <input
              type="checkbox"
              checked={draft.balcony === "true"}
              onChange={(e) => set("balcony", e.target.checked ? "true" : "")}
            />
            Balcony
          </label>
          <button className="btn" type="submit">
            Apply
          </button>
        </form>

        <div>
          {query.isError && <p className="error">Search failed. Is the API running?</p>}
          <div className="muted" style={{ marginBottom: 10 }}>
            {query.data ? `${query.data.total} properties` : "Loading…"}
          </div>
          <div className="panel table-wrap">
            <table>
              <thead>
                <tr>
                  <th></th>
                  <th>Address</th>
                  <th>Price</th>
                  <th>UAH/m²</th>
                  <th>Area</th>
                  <th>Rooms</th>
                  <th>Source</th>
                  <th>First seen</th>
                </tr>
              </thead>
              <tbody>
                {(query.data?.items ?? []).map((p) => (
                  <tr key={p.id} className="clickable">
                    <td>{p.photo_url ? <img className="thumb" src={p.photo_url} alt="" /> : null}</td>
                    <td>
                      <Link to={`/property/${p.id}`}>{p.address}</Link>
                      <div className="muted">{p.neighborhood}</div>
                    </td>
                    <td className="mono">{formatPrice(p.price_uah ?? p.asking_price, p.price_usd)}</td>
                    <td className="mono">{p.price_per_m2 ? Math.round(p.price_per_m2).toLocaleString("uk-UA") : "—"}</td>
                    <td className="mono">{p.area_m2 ?? "—"}</td>
                    <td className="mono">{p.rooms ?? "—"}</td>
                    <td>
                      {p.source}
                      {p.listing_count > 1 ? ` · ${p.listing_count}` : ""}
                    </td>
                    <td>{formatWhen(p.first_seen_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="pager">
            <button
              className="btn secondary"
              disabled={offset === 0}
              onClick={() => setOffset((o) => Math.max(0, o - 20))}
            >
              Previous
            </button>
            <button
              className="btn secondary"
              disabled={!query.data || offset + 20 >= query.data.total}
              onClick={() => setOffset((o) => o + 20)}
            >
              Next
            </button>
            <span>offset {offset}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
