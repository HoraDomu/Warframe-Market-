import requests
import json

item = input("What Item are Searching?: ").strip().lower().replace(" ", "_")
url = f"https://api.warframe.market/v1/items/{item}/orders"

resp = requests.get(url, headers={"User-Agent": "simple-script"})

try:
    resp_json = resp.json()
except json.JSONDecodeError:
    print("Error: Could not decode JSON from the API response.")
    exit()

# Check if 'payload' exists
if "payload" not in resp_json:
    print("Error: Item not found or API returned an unexpected response.")
    print(json.dumps(resp_json, indent=4))  # Show API response for debugging
    exit()

data = resp_json["payload"]["orders"]

lowest_sell = min(
    [o["platinum"] for o in data if o["order_type"] == "sell" and o["visible"]],
    default=None,
)

highest_buy = max(
    [o["platinum"] for o in data if o["order_type"] == "buy" and o["visible"]],
    default=None,
)

if highest_buy is not None:
    print(f"The highest_buy is: {highest_buy}")
else:
    print("No highest_buy")

if lowest_sell is not None:
    print(f"The lowest_sell is: {lowest_sell}")
else:
    print("No lowest_sell")
