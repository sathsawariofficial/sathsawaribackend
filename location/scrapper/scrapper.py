import geopandas as gpd
import json
import os

# 1. Load the shapefile
print("Loading Shapefile...")
input_path = '/home/raotalha/Code/PersonalCode/rideshare/backend/location/data/polling_station_pakistan/Polling_Station_Pakistan.shp'
gdf = gpd.read_file(input_path)

# 2. Identify the correct columns automatically
dist_col = next((c for c in gdf.columns if 'dist' in c.lower()), None)
name_col = next((c for c in gdf.columns if 'build' in c.lower() or 'name' in c.lower() or 'station' in c.lower()), None)

# 3. Comprehensive list of Ridesharing Hubs
districts_to_include = [
    # Tier 1
    'islamabad', 'rawalpindi', 'lahore', 'karachi east', 'karachi west', 
    'karachi south', 'karachi central', 'malir', 'korangi', 'keamari', 
    
    # Tier 2
    'faisalabad', 'multan', 'peshawar', 'gujranwala', 'sialkot', 'hyderabad',
    'wah cantt', 'taxila', 'murree', 'kahuta'
    
    # Tier 3 & Growth Cities
    'bahawalpur', 'sargodha', 'sukkur', 'abbottabad', 'sahiwal', 'gujrat', 
    'mardan', 'rahim yar khan', 'sheikhupura', 'kasur', 'okara', 'jhang',
    'dera ghazi khan', 'muzaffarabad', 'mirpur'
]

# 4. Filter and Convert CRS to WGS84 (Lat/Lon)
print("Filtering and processing coordinates...")
filtered_gdf = gdf[gdf[dist_col].str.lower().str.strip().isin(districts_to_include)].copy()

if filtered_gdf.crs and filtered_gdf.crs.to_epsg() != 4326:
    filtered_gdf = filtered_gdf.to_crs(epsg=4326)

# 5. Build the JSON List
locations_json = []

for _, row in filtered_gdf.iterrows():
    # Construct a clean entry for each location
    entry = {
        "name": str(row[name_col]).strip(),
        "district": str(row[dist_col]).strip(),
        "lat": float(row.geometry.y),
        "lng": float(row.geometry.x),
        "category": "transit_point" 
    }
    locations_json.append(entry)

# 6. Save to JSON file
output_file = 'pakistan_locations.json'
with open(output_file, 'w', encoding='utf-8') as f:
    json.dump(locations_json, f, ensure_ascii=False, indent=2)

print(f"--- SUCCESS ---")
print(f"Total places saved: {len(locations_json)}")
print(f"File saved to: {os.path.abspath(output_file)}")