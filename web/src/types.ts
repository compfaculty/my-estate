export type Geo = { lat: number; lng: number };

export type Photo = { url: string; sort_order: number };

export type PropertyCard = {
  id: string;
  address: string;
  city: string;
  neighborhood: string;
  street: string;
  house_number: string;
  property_type: string;
  rooms: number | null;
  area_m2: number | null;
  floor: number | null;
  asking_price: number | null;
  currency: string;
  price_per_m2: number | null;
  price_uah: number | null;
  price_usd: number | null;
  source: string;
  source_url: string;
  first_seen_at: string;
  photo_url: string;
  listing_count: number;
  location?: Geo;
  elevator?: boolean | null;
  parking?: boolean | null;
  balcony?: boolean | null;
  safe_room?: boolean | null;
};

export type PageResponse = {
  items: PropertyCard[];
  total: number;
  limit: number;
  offset: number;
};

export type PropertyJSON = {
  id: string;
  country: string;
  city: string;
  district: string;
  neighborhood: string;
  street: string;
  house_number: string;
  address: string;
  location?: Geo;
  property_type: string;
  rooms: number | null;
  bedrooms: number | null;
  bathrooms: number | null;
  floor: number | null;
  total_floors: number | null;
  area_m2: number | null;
  balcony_area_m2: number | null;
  parking_spaces: number | null;
  construction_year: number | null;
  renovation_state: string;
  elevator: boolean | null;
  parking: boolean | null;
  storage: boolean | null;
  balcony: boolean | null;
  air_conditioning: boolean | null;
  safe_room: boolean | null;
  accessibility: boolean | null;
  field_provenance: { field: string; source: string; confidence: number }[];
  created_at: string;
  updated_at: string;
};

export type ListingJSON = {
  id: string;
  property_id: string;
  source: string;
  source_listing_id: string;
  source_url: string;
  listing_type: string;
  asking_price: number | null;
  currency: string;
  price_per_m2: number | null;
  price_uah: number | null;
  price_usd: number | null;
  price_per_m2_uah: number | null;
  fx_source: string;
  description: string;
  seller_name: string;
  seller_type: string;
  first_seen_at: string;
  last_seen_at: string;
  removed_at: string | null;
  status: string;
  photos: Photo[];
  match_confidence: number | null;
};

export type PropertyDetail = {
  property: PropertyJSON;
  listings: ListingJSON[];
};

export type HistoryItem = {
  id: string;
  listing_id: string;
  recorded_at: string;
  price: number | null;
  currency: string;
  price_uah: number | null;
  price_usd: number | null;
  description_hash: string;
  status: string;
  seller_type: string;
  photos_hash: string;
};

export type SourceItem = {
  id: string;
  name: string;
  enabled: boolean;
  fetch_mode: string;
  last_success_at: string | null;
  last_error: string | null;
  last_crawl_at: string | null;
  last_run_status: string;
  listings_discovered: number;
  listings_updated: number;
};

export type SearchParams = {
  city?: string;
  neighborhood?: string;
  q?: string;
  min_price?: string;
  max_price?: string;
  min_area?: string;
  max_area?: string;
  min_rooms?: string;
  max_rooms?: string;
  source?: string;
  elevator?: string;
  parking?: string;
  balcony?: string;
  limit?: number;
  offset?: number;
};
