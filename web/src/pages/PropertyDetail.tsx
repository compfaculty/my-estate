import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { fetchHistory, fetchProperty } from "../api";
import { PropertyMap } from "../components/PropertyMap";
import { formatPrice, formatWhen } from "../format";

export function PropertyDetail() {
  const { id } = useParams();
  const [photo, setPhoto] = useState(0);
  const detail = useQuery({
    queryKey: ["property", id],
    queryFn: () => fetchProperty(id!),
    enabled: Boolean(id),
  });
  const history = useQuery({
    queryKey: ["history", id],
    queryFn: () => fetchHistory(id!),
    enabled: Boolean(id),
  });

  if (detail.isLoading) {
    return (
      <div className="page">
        <p className="muted">Loading property…</p>
      </div>
    );
  }
  if (detail.isError || !detail.data) {
    return (
      <div className="page">
        <p className="error">Property not found.</p>
      </div>
    );
  }

  const { property, listings } = detail.data;
  const photos = listings.flatMap((l) => l.photos).sort((a, b) => a.sort_order - b.sort_order);
  const primary = listings.find((l) => l.status === "active") ?? listings[0];

  return (
    <div className="page page-wide">
      <h1>{property.address}</h1>
      <p className="lede">
        {property.property_type} · {property.neighborhood}, {property.city}
      </p>

      <div className="detail-hero">
        <div className="gallery">
          {photos[photo] ? (
            <img className="main" src={photos[photo].url} alt="" />
          ) : (
            <div className="placeholder">No photos</div>
          )}
          {photos.slice(0, 6).map((p, i) => (
            <img key={`${p.url}-${i}`} src={p.url} alt="" onClick={() => setPhoto(i)} />
          ))}
        </div>
        <div className="panel section" style={{ margin: 0 }}>
          <h3>Overview</h3>
          <div className="kv">
            <span>Asking</span>
            <strong className="mono">{formatPrice(primary?.price_uah ?? primary?.asking_price, primary?.price_usd)}</strong>
            <span>₴ / m²</span>
            <span className="mono">
              {primary?.price_per_m2_uah ?? primary?.price_per_m2
                ? Math.round(primary.price_per_m2_uah ?? primary.price_per_m2 ?? 0).toLocaleString("uk-UA")
                : "—"}
            </span>
            <span>Area</span>
            <span className="mono">{property.area_m2 ?? "—"} m²</span>
            <span>Rooms</span>
            <span className="mono">{property.rooms ?? "—"}</span>
            <span>Floor</span>
            <span className="mono">
              {property.floor ?? "—"} / {property.total_floors ?? "—"}
            </span>
            <span>Year</span>
            <span className="mono">{property.construction_year ?? "—"}</span>
            <span>Condition</span>
            <span>{property.renovation_state || "—"}</span>
            <span>Elevator</span>
            <span>{boolLabel(property.elevator)}</span>
            <span>Parking</span>
            <span>{boolLabel(property.parking)}</span>
            <span>Safe room</span>
            <span>{boolLabel(property.safe_room)}</span>
          </div>
        </div>
      </div>

      <div className="section">
        <h3>Map</h3>
        <PropertyMap location={property.location} />
      </div>

      <div className="section">
        <h3>Sources</h3>
        <table>
          <thead>
            <tr>
              <th>Source</th>
              <th>Listing ID</th>
              <th>Price</th>
              <th>Seller</th>
              <th>Status</th>
              <th>Link</th>
            </tr>
          </thead>
          <tbody>
            {listings.map((l) => (
              <tr key={l.id}>
                <td>{l.source}</td>
                <td className="mono">{l.source_listing_id}</td>
                <td className="mono">{formatPrice(l.price_uah ?? l.asking_price, l.price_usd)}</td>
                <td>
                  {l.seller_name} <span className="badge">{l.seller_type}</span>
                </td>
                <td>
                  <span className={`badge ${l.status === "active" ? "ok" : "bad"}`}>{l.status}</span>
                </td>
                <td>
                  <a href={l.source_url} target="_blank" rel="noreferrer">
                    Open
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {primary?.description ? <p style={{ marginTop: 12 }}>{primary.description}</p> : null}
      </div>

      <div className="section">
        <h3>Listing history</h3>
        <table>
          <thead>
            <tr>
              <th>When</th>
              <th>Price</th>
              <th>Status</th>
              <th>Seller type</th>
            </tr>
          </thead>
          <tbody>
            {(history.data?.items ?? []).map((h) => (
              <tr key={h.id}>
                <td>{formatWhen(h.recorded_at)}</td>
                <td className="mono">{formatPrice(h.price_uah ?? h.price)}</td>
                <td>{h.status}</td>
                <td>{h.seller_type}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="section">
        <h3>Evaluation</h3>
        <div className="placeholder">Comparable valuation and opportunity score — Phase 3.</div>
      </div>
      <div className="section">
        <h3>Risk signals</h3>
        <div className="placeholder">Scam / risk agent findings — Phase 6. Signals only, not accusations.</div>
      </div>
      <div className="section">
        <h3>Comparables</h3>
        <div className="placeholder">Neighborhood comps — Phase 3.</div>
      </div>
    </div>
  );
}

function boolLabel(v: boolean | null) {
  if (v == null) return "—";
  return v ? "Yes" : "No";
}
