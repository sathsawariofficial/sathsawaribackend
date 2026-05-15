import json
import os


def unify_data():
    unified_data = {}

    # helper to create a unique key based on coordinates to avoid duplicates
    def get_geo_key(lat, lng):
        return f"{round(float(lat), 4)}_{round(float(lng), 4)}"

    def add_to_unified(name, lat, lng, district, category):
        if not name or str(name).lower() == "null":
            return

        key = get_geo_key(lat, lng)
        if key not in unified_data:
            # Structuring for Meilisearch Geosearch
            unified_data[key] = {
                "id": len(unified_data) + 1,
                "name": str(name).replace('\n', ' ').strip(),
                "district": district,
                "lat": float(lat),
                "lng": float(lng),
                "category": category,
                "_geo": {
                    "lat": float(lat),
                    "lng": float(lng)
                }
            }

    # Paths (using your provided absolute paths)
    paths = {
        "hotosm": '/home/raotalha/Code/PersonalCode/rideshare/backend/location/data/hotosm_pak_populated_places_points_geojson/hotosm_pak_populated_places_points_geojson.geojson',
        "overpass": '/home/raotalha/Code/PersonalCode/rideshare/backend/location/data/export.geojson',
        "cities": '/home/raotalha/Code/PersonalCode/rideshare/backend/location/data/pk.json',
        "polling": '/home/raotalha/Code/PersonalCode/rideshare/backend/location/scrapper/pakistan_locations.json'
    }

    # 1. HotOSM
    if os.path.exists(paths["hotosm"]):
        with open(paths["hotosm"], 'r') as f:
            data = json.load(f)
            for feat in data.get('features', []):
                props = feat['properties']
                coords = feat['geometry']['coordinates']
                name = props.get('name:en') or props.get('name')
                add_to_unified(name, coords[1], coords[0], props.get(
                    'is_in', 'Unknown'), 'area')

    # 2. Overpass
    if os.path.exists(paths["overpass"]):
        with open(paths["overpass"], 'r') as f:
            data = json.load(f)
            for feat in data.get('features', []):
                props = feat['properties']
                coords = feat['geometry']['coordinates']
                cat = 'bus_stop' if 'highway' in props else 'landmark'
                add_to_unified(props.get('name'), coords[1], coords[0], props.get(
                    'is_in', 'Unknown'), cat)

    # 3. Cities
    if os.path.exists(paths["cities"]):
        with open(paths["cities"], 'r') as f:
            data = json.load(f)
            for item in data:
                add_to_unified(item.get('city'), item.get('lat'), item.get(
                    'lng'), item.get('admin_name'), 'city')

    # 4. Polling
    if os.path.exists(paths["polling"]):
        with open(paths["polling"], 'r') as f:
            data = json.load(f)
            for item in data:
                add_to_unified(item.get('name'), item.get('lat'), item.get(
                    'lng'), item.get('district'), 'local_stop')

    final_list = list(unified_data.values())
    with open('unified_pakistan_locations.json', 'w', encoding='utf-8') as f:
        json.dump(final_list, f, indent=2, ensure_ascii=False)

    print(f"Combined everything! Total unique locations: {len(final_list)}")


if __name__ == "__main__":
    unify_data()
