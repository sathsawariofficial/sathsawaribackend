import requests
from bs4 import BeautifulSoup
import json


def fetch_place_names(base_url, max_start_row, step):
    place_names = []
    for start_row in range(0, max_start_row + step, step):
        # Construct the URL with the pagination parameter
        url = f"{base_url}&startRow={start_row}"
        print(f"Fetching: {url}")

        # Fetch the page
        response = requests.get(url)
        if response.status_code != 200:
            print(
                f"Failed to fetch {url} (status code: {response.status_code})")
            continue

        # Parse the HTML
        soup = BeautifulSoup(response.content, 'html.parser')

        # Find all anchor tags that match the desired structure
        for anchor in soup.find_all('a', href=True):
            # Example check to filter relevant links
            import re

            if re.search(r'/\d+/[a-zA-Z0-9_-]+\.html$', anchor['href']):
                # Do something
                place_names.append(anchor.text.strip())

    return place_names


def main():
    base_url = "http://www.geonames.org/search.html?q=rawalpindi&country=PK"
    max_start_row = 2800
    step = 50

    # Fetch place names
    place_names = fetch_place_names(base_url, max_start_row, step)

    # Output to JSON
    output_file = "places_rawalpindi.json"
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(place_names, f, ensure_ascii=False, indent=4)

    print(f"Scraped {len(place_names)} place names.")
    print(f"Output written to {output_file}")


if __name__ == "__main__":
    main()
