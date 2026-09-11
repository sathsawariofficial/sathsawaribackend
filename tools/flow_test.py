#!/usr/bin/env python3
"""Drives the whole product end to end against the running server and records the
real request/response of every call, so the Postman collection can carry genuine
examples instead of invented ones.

Every business rule of Pick & Drop and shifts is proved twice: once by the call that
is allowed, and once by the call that is refused."""
import json
import urllib.request
import urllib.error
import sys
import os
import time
from urllib.parse import quote
from datetime import datetime, timedelta, timezone

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
WORK = os.path.join(ROOT, ".sath")
OPEN_TOKEN = open(os.path.join(WORK, "open_token.txt")).read().strip()
BASE = os.environ.get("SATH_BASE", "http://localhost:5000")
RESULTS = []                  # every call, for the collection
ENV = {}

FAILS = []
CHECKS = []


def call(label, method, path, token=None, body=None, expect=200, note=""):
    url = BASE + path
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(req, timeout=25) as r:
            status, raw = r.status, r.read().decode()
    except urllib.error.HTTPError as e:
        status, raw = e.code, e.read().decode()
    except Exception as e:
        status, raw = 0, json.dumps({"error": str(e)})

    try:
        parsed = json.loads(raw)
    except Exception:
        parsed = {"raw": raw[:400]}

    ok = (status == expect)
    RESULTS.append({
        "label": label, "method": method, "path": path,
        "request": body, "status": status, "response": parsed,
        "expected": expect, "ok": ok, "note": note,
    })
    mark = "OK " if ok else "!! "
    print(f"{mark}{status:3} {method:6} {path[:58]:58} {label}")
    if not ok:
        FAILS.append((label, method, path, status, expect, json.dumps(parsed)[:300]))
        print(f"      expected {expect} — {json.dumps(parsed)[:260]}")
    return parsed


def check(label, condition, detail=""):
    """Asserts something about a response that the status code alone cannot prove."""
    CHECKS.append((label, bool(condition)))
    if condition:
        print(f"OK  ✓  {'check':6} {detail[:58]:58} {label}")
    else:
        FAILS.append((label, "CHECK", detail, "-", "-", ""))
        print(f"!!  ✗  {'check':6} {detail[:58]:58} {label}")


def data(r, *keys):
    d = (r or {}).get("data") or {}
    for k in keys:
        if isinstance(d, dict):
            d = d.get(k)
        else:
            return None
    return d


# the server keeps Pick & Drop time in PKT, so the test does too
def pkt_now():
    return datetime.now(timezone.utc) + timedelta(hours=5)


def wait_for_clock(clock):
    """Sleeps until the PKT wall clock reaches HH:MM, so a trip that was due has started."""
    if pkt_now().strftime("%H:%M") < clock:
        print(f"      waiting for {clock} PKT so today's trip has started")
    while pkt_now().strftime("%H:%M") < clock:
        time.sleep(5)


NOW = pkt_now()
TODAY = NOW.date()


def day(offset):
    return (TODAY + timedelta(days=offset)).strftime("%Y-%m-%d")


D0 = day(1)          # the day before the recurring shifts begin
D1 = day(2)          # the first day they run
D_END = day(16)
D_RIDE = day(10)     # a carpool day far enough out for the 2 hour guard
DOW_TODAY = TODAY.isoweekday()
DOW_RIDE = (TODAY + timedelta(days=10)).isoweekday()
ALL_DAYS = [1, 2, 3, 4, 5, 6, 7]
SUF = datetime.now().strftime("%H%M%S")

BAHRIA = ("Bahria Town Phase 4", 33.5121, 73.0951)
DHA = ("DHA Phase 2", 33.5350, 73.1350)
ROOTS = ("Roots School F-8", 33.7101, 73.0441)
MARKAZ = ("F-8 Markaz", 33.7090, 73.0400)
GULBERG = ("Gulberg Greens", 33.6180, 73.1560)
G11 = ("G-11 Markaz", 33.6680, 72.9980)
BLUE = ("Blue Area", 33.7100, 73.0600)


def at(place, time):
    return {"location": place[0], "lat": place[1], "lng": place[2], "time": time}


def section(title):
    print("=" * 100)
    print(title)
    print("=" * 100)


def ids_of(rows, key):
    return [row.get(key) for row in (rows or [])]


section("ADMIN")
r = call("Admin login", "POST", "/api/v1/admin/login", OPEN_TOKEN,
         {"username": "twssawari", "password": "vR7!xK2@pQ9#Lm4$Zw8^Ty1&Nc5*Hs3%Df6!Ba"})
ENV["admin"] = data(r, "sessionId")
adm = ENV.get("admin")
call("Admin overview", "GET", "/api/v1/admin/overview", adm)

section("ACCOUNTS")
OWNER_M = "+92340" + SUF + "1"
DRV2_M = "+92340" + SUF + "2"
PSG1_M = "+92340" + SUF + "3"
PSG2_M = "+92340" + SUF + "4"
PSG3_M = "+92340" + SUF + "5"

r = call("Register owner driver", "POST", "/api/v1/driver/register", OPEN_TOKEN,
         {"deviceId": "flow", "mobile": OWNER_M, "name": "Owner Driver",
          "password": "Golang@12122", "gender": "male"})
call("Verify driver otp", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": OWNER_M, "otp": data(r, "tempOTP"), "operation": "ACTIVATE_DRIVER"})
r = call("Login owner driver", "POST", "/api/v1/driver/login", OPEN_TOKEN,
         {"deviceId": "flow", "mobile": OWNER_M, "password": "Golang@12122"})
ENV["owner"] = data(r, "sessionId")
ENV["owner_id"] = data(r, "driver", "id")
call("Driver profile", "GET", "/api/v1/driver/info", ENV["owner"])
call("Set pin", "POST", "/api/v1/driver/pin?pin=121212", ENV["owner"])

r = call("Register vehicle (7 seats)", "POST", "/api/v1/vehicle/register", ENV["owner"],
         {"vehicleNumber": f"PD-{SUF}A", "vehicleInfo": "Hiace van",
          "numberOfSeats": 7, "hasAC": True, "hasHeating": False, "pin": "121212"})
ENV["vehicle"] = data(r, "vehicleId")
r = call("Register small vehicle (1 seat)", "POST", "/api/v1/vehicle/register", ENV["owner"],
         {"vehicleNumber": f"PD-{SUF}C", "vehicleInfo": "Mehran",
          "numberOfSeats": 1, "hasAC": False, "hasHeating": False, "pin": "121212"})
ENV["vehicle3"] = data(r, "vehicleId")
call("Get vehicles", "GET", "/api/v1/vehicle/", ENV["owner"])

r = call("Register second driver", "POST", "/api/v1/driver/register", OPEN_TOKEN,
         {"deviceId": "flow2", "mobile": DRV2_M, "name": "Second Driver",
          "password": "Golang@12122", "gender": "male"})
call("Verify driver 2 otp", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": DRV2_M, "otp": data(r, "tempOTP"), "operation": "ACTIVATE_DRIVER"})
r = call("Login second driver", "POST", "/api/v1/driver/login", OPEN_TOKEN,
         {"deviceId": "flow2", "mobile": DRV2_M, "password": "Golang@12122"})
ENV["drv2"] = data(r, "sessionId")
ENV["drv2_id"] = data(r, "driver", "id")
call("Second driver pin", "POST", "/api/v1/driver/pin?pin=131313", ENV["drv2"])
r = call("Register second vehicle", "POST", "/api/v1/vehicle/register", ENV["drv2"],
         {"vehicleNumber": f"PD-{SUF}B", "vehicleInfo": "Corolla",
          "numberOfSeats": 4, "hasAC": True, "hasHeating": True, "pin": "131313"})
ENV["vehicle2"] = data(r, "vehicleId")
r = call("Register fourth vehicle", "POST", "/api/v1/vehicle/register", ENV["drv2"],
         {"vehicleNumber": f"PD-{SUF}D", "vehicleInfo": "Cultus",
          "numberOfSeats": 4, "hasAC": True, "hasHeating": False, "pin": "131313"})
ENV["vehicle4"] = data(r, "vehicleId")

for key, mob, name, gender, dev in (("psg1", PSG1_M, "Ayesha", "female", "p1"),
                                    ("psg2", PSG2_M, "Bilal", "male", "p2"),
                                    ("psg3", PSG3_M, "Sana", "female", "p3")):
    r = call(f"Register passenger {name}", "POST", "/api/v1/passenger/register", OPEN_TOKEN,
             {"deviceId": dev, "mobile": mob, "name": name, "gender": gender, "password": "Golang@12122"})
    call(f"Verify passenger {name} otp", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
         {"mobile": mob, "otp": data(r, "tempOTP"), "operation": "ACTIVATE_PASSENGER"})
    r = call(f"Login passenger {name}", "POST", "/api/v1/passenger/login", OPEN_TOKEN,
             {"deviceId": dev, "mobile": mob, "password": "Golang@12122"})
    ENV[key] = data(r, "sessionId")
    ENV[key + "_id"] = data(r, "passenger", "id")

call("Passenger profile", "GET", "/api/v1/passenger/info", ENV["psg1"])
call("Passenger forgot password", "GET", f"/api/v1/passenger/password/forgot?mobile_number={quote(PSG1_M)}", OPEN_TOKEN)

OWNER, DRV2, P1, P2, P3 = ENV["owner"], ENV["drv2"], ENV["psg1"], ENV["psg2"], ENV["psg3"]
OWNER_ID, DRV2_ID = ENV["owner_id"], ENV["drv2_id"]
P1_ID, P2_ID, P3_ID = ENV["psg1_id"], ENV["psg2_id"], ENV["psg3_id"]
V1, V2, V3, V4 = ENV["vehicle"], ENV["vehicle2"], ENV["vehicle3"], ENV["vehicle4"]

section("PICK & DROP SERVICE")
r = call("Enable Pick & Drop", "POST", "/api/v1/pickdrop", OWNER,
         {"name": f"Islamabad School Run {SUF}", "description": "Morning and afternoon school runs",
          "vehicleIds": [V1]})
ENV["service"] = data(r, "serviceId")
S = ENV["service"]
call("Enable a second service (refused)", "POST", "/api/v1/pickdrop", OWNER,
     {"name": "Another service", "description": "", "vehicleIds": []},
     expect=400, note="a driver owns at most one Pick & Drop service")
r = call("My service (owner)", "GET", "/api/v1/pickdrop", OWNER)
check("Advertisement limit is approved vehicles x 2", data(r, "counts", "advertisementLimit") == 2,
      f"limit {data(r, 'counts', 'advertisementLimit')}")
call("Search services (driver)", "GET", "/api/v1/pickdrop/search?page=1&search=School", DRV2)
call("Search services (passenger)", "GET", "/api/v1/passenger/pickdrop/search?page=1&search=School", P1)
call("Passenger token on a driver route (refused)", "GET", "/api/v1/pickdrop", P1,
     expect=401, note="the middleware rejects the wrong token type before the handler is reached")

section("ADVERTISEMENTS")
# run while the owner's one vehicle is the only approved one, so the limit is two
ad_body = {"title": "Bahria to F-8 school run", "description": "Morning pick up in an AC van", "fare": 6000,
           "daysOfWeek": [1, 2, 3, 4, 5], "startTime": "06:30", "endTime": "08:30",
           "locations": [at(BAHRIA, "06:45"), at(DHA, "07:10"), at(ROOTS, "08:00")]}
call("Advertisement with eleven places (refused)", "POST", "/api/v1/pickdrop/advertisement", OWNER,
     dict(ad_body, locations=[at(BAHRIA, f"06:{31 + i}") for i in range(11)]),
     expect=400, note="an advertisement carries up to ten places")
r = call("Create advertisement", "POST", "/api/v1/pickdrop/advertisement", OWNER, ad_body)
ENV["ad1"] = data(r, "advertisementId")
r = call("Create second advertisement", "POST", "/api/v1/pickdrop/advertisement", OWNER,
         dict(ad_body, title="DHA to F-8 school run", locations=[at(DHA, "07:00"), at(ROOTS, "08:00")]))
ENV["ad2"] = data(r, "advertisementId")
call("Advertisement over the limit (refused)", "POST", "/api/v1/pickdrop/advertisement", OWNER,
     dict(ad_body, title="One too many"), expect=400,
     note="one approved vehicle allows two advertisements")
call("My advertisements", "GET", "/api/v1/pickdrop/advertisements?page=1", OWNER)
r = call("Search advertisements (public)", "GET", "/api/v1/pickdrop/advertisements/search?page=1&search=Bahria", OPEN_TOKEN)
check("Public search finds the advertisement", ENV["ad1"] in ids_of(data(r, "advertisements"), "id"))
call("Delete advertisement", "DELETE", f"/api/v1/pickdrop/advertisement?advertisement_id={ENV['ad2']}", OWNER)
call("Add another own vehicle", "POST", "/api/v1/pickdrop/vehicles", OWNER, {"vehicleIds": [V3]})
r = call("My service after adding a vehicle", "GET", "/api/v1/pickdrop", OWNER)
check("Advertisement limit follows the vehicles", data(r, "counts", "advertisementLimit") == 4,
      f"limit {data(r, 'counts', 'advertisementLimit')}")

section("JOIN REQUESTS")
r = call("Driver join request (with a vehicle)", "POST", "/api/v1/pickdrop/request", DRV2,
         {"serviceId": S, "vehicleIds": [V2]})
call("Duplicate join request (refused)", "POST", "/api/v1/pickdrop/request", DRV2,
     {"serviceId": S, "vehicleIds": []}, expect=400, note="a request is already waiting")
call("Enable a service while a request is open (refused)", "POST", "/api/v1/pickdrop", DRV2,
     {"name": "Second Driver Service", "description": "", "vehicleIds": []},
     expect=400, note="a driver belongs to one service, as owner or as member")
call("Owner asks to join a service (refused)", "POST", "/api/v1/pickdrop/request", OWNER,
     {"serviceId": S, "vehicleIds": []}, expect=400, note="an owner cannot also be a member")
call("Availability before approval (refused)", "PUT", "/api/v1/passenger/availability", P1,
     {"days": [{"dayOfWeek": 1, "isRequired": True, "locations": [at(BAHRIA, "07:00")]}]},
     expect=400, note="weekly availability opens once the owner approves the passenger")
for name, token in (("Ayesha", P1), ("Bilal", P2), ("Sana", P3)):
    call(f"Passenger {name} join request", "POST", "/api/v1/passenger/pickdrop/request", token, {"serviceId": S})
call("Passenger duplicate join request (refused)", "POST", "/api/v1/passenger/pickdrop/request", P1,
     {"serviceId": S}, expect=400, note="a passenger belongs to one service")
call("Passenger membership (pending)", "GET", "/api/v1/passenger/pickdrop", P1)

r = call("Pending driver requests", "GET", "/api/v1/pickdrop/requests?type=driver&page=1", OWNER)
driver_reqs = {x["driverId"]: x["requestId"] for x in (data(r, "requests") or [])}
r = call("Pending vehicle requests", "GET", "/api/v1/pickdrop/requests?type=vehicle&page=1", OWNER)
vehicle_reqs = {x["vehicleId"]: x["requestId"] for x in (data(r, "requests") or [])}
r = call("Pending passenger requests", "GET", "/api/v1/pickdrop/requests?type=passenger&page=1", OWNER)
passenger_reqs = {x["passengerId"]: x["requestId"] for x in (data(r, "requests") or [])}
ENV["driver2_request"] = driver_reqs.get(DRV2_ID)
ENV["vehicle2_request"] = vehicle_reqs.get(V2)
call("Join request detail", "GET", f"/api/v1/pickdrop/request?type=driver&request_id={ENV['driver2_request']}", OWNER)

r = call("Approve a vehicle before its driver (skipped)", "PATCH", "/api/v1/pickdrop/requests", OWNER,
         {"decisions": [{"type": "vehicle", "requestId": ENV["vehicle2_request"], "action": "approve"}]},
         note="a vehicle is only approved once the driver who offered it is")
check("Vehicle before driver comes back skipped", len(data(r, "skipped") or []) == 1)

decisions = [
    {"type": "passenger", "requestId": passenger_reqs.get(P1_ID), "action": "approve"},
    {"type": "passenger", "requestId": passenger_reqs.get(P2_ID), "action": "approve"},
    {"type": "passenger", "requestId": passenger_reqs.get(P3_ID), "action": "reject"},
    {"type": "vehicle", "requestId": ENV["vehicle2_request"], "action": "approve"},
    {"type": "driver", "requestId": ENV["driver2_request"], "action": "approve"},
]
r = call("Decide requests in bulk", "PATCH", "/api/v1/pickdrop/requests", OWNER, {"decisions": decisions},
         note="drivers are settled first, then vehicles, then passengers, whatever order they are sent in")
check("Every bulk decision applied", len(data(r, "applied") or []) == 5, f"applied {len(data(r, 'applied') or [])}")
r = call("Remove a rejected passenger (skipped)", "PATCH", "/api/v1/pickdrop/requests", OWNER,
         {"decisions": [{"type": "passenger", "requestId": passenger_reqs.get(P3_ID), "action": "remove"}]},
         note="only an approved member can be removed")
check("Removing a rejected request is skipped", len(data(r, "skipped") or []) == 1)
r = call("Re-decide the same requests (all skipped)", "PATCH", "/api/v1/pickdrop/requests", OWNER,
         {"decisions": decisions}, note="already decided, every line comes back skipped with the reason")
check("Re-deciding skips every line", len(data(r, "skipped") or []) == 5)

call("My service (member driver)", "GET", "/api/v1/pickdrop", DRV2)
call("Passenger membership (approved)", "GET", "/api/v1/passenger/pickdrop", P1)
call("Offer another vehicle", "POST", "/api/v1/pickdrop/vehicles/offer", DRV2, {"vehicleIds": [V4]})
call("Withdraw the offered vehicle", "DELETE", f"/api/v1/pickdrop/vehicle/offer?vehicle_id={V4}", DRV2)
call("Withdraw a vehicle that is not offered (refused)", "DELETE", f"/api/v1/pickdrop/vehicle/offer?vehicle_id={V4}", DRV2,
     expect=400, note="nothing open for that vehicle any more")
call("Member driver adds vehicles as owner (refused)", "POST", "/api/v1/pickdrop/vehicles", DRV2,
     {"vehicleIds": [V4]}, expect=400, note="only the owner adds vehicles straight into the service")

section("PASSENGER AVAILABILITY")
call("Set weekly availability", "PUT", "/api/v1/passenger/availability", P1, {"days": [
    {"dayOfWeek": 1, "isRequired": True, "locations": [at(BAHRIA, "06:55"), at(ROOTS, "13:30")]},
    {"dayOfWeek": 2, "isRequired": True, "locations": [at(BAHRIA, "06:55")]},
    {"dayOfWeek": 3, "isRequired": False, "locations": []},
]}, note="Wednesday is not needed, the other days not sent stay as they were")
call("Get weekly availability", "GET", "/api/v1/passenger/availability", P1)
call("Passenger 2 weekly availability", "PUT", "/api/v1/passenger/availability", P2, {"days": [
    {"dayOfWeek": d, "isRequired": True, "locations": [at(DHA, "07:20")]} for d in (1, 2, 3, 4, 5)
]})
call("Seven places in a day (refused)", "PUT", "/api/v1/passenger/availability", P1, {"days": [
    {"dayOfWeek": 4, "isRequired": True,
     "locations": [at(BAHRIA, f"06:0{i}") for i in range(7)]},
]}, expect=400, note="a day holds up to six places")
call("Places on a day that is not required (refused)", "PUT", "/api/v1/passenger/availability", P1, {"days": [
    {"dayOfWeek": 5, "isRequired": False, "locations": [at(BAHRIA, "07:00")]},
]}, expect=400, note="a day off carries no places")

call("Available drivers", "GET", "/api/v1/pickdrop/drivers/available?page=1", OWNER)
call("Available drivers for a schedule", "GET",
     f"/api/v1/pickdrop/drivers/available?page=1&days_of_week=1,2,3,4,5,6,7&start_date={D1}&start_time=07:00&end_time=08:00",
     OWNER, note="with a schedule each driver says whether they are free for it")
call("Available vehicles", "GET", "/api/v1/pickdrop/vehicles/available?page=1", OWNER)
r = call("Available passengers with weekly demand", "GET", "/api/v1/pickdrop/passengers/available?page=1", OWNER)
demand = {x["passengerId"]: x.get("weeklyDemand") for x in (data(r, "passengers") or [])}
check("Owner sees the passenger's weekly demand", bool(demand.get(P1_ID)))

section("SHIFTS")
shift_a = {
    "driverId": OWNER_ID, "vehicleId": V1,
    "daysOfWeek": ALL_DAYS, "startDate": D1, "endDate": D_END,
    "locations": [at(BAHRIA, "07:00"), at(DHA, "07:25"), at(ROOTS, "08:00")],
    "passengers": [{"passengerId": P1_ID, "locationSequence": 1},
                   {"passengerId": P2_ID, "locationSequence": 2}],
}
r = call("Create shift", "POST", "/api/v1/shift", OWNER, shift_a)
ENV["shift"] = data(r, "shiftId")
A = ENV["shift"]
check("Seats are counted on the shift", data(r, "seats") == {"totalSeats": 7, "occupiedSeats": 2, "remainingSeats": 5},
      json.dumps(data(r, "seats")))


def variant(**changes):
    body = json.loads(json.dumps(shift_a))
    body.update(changes)
    return body


call("Shift in the past (refused)", "POST", "/api/v1/shift", OWNER,
     variant(startDate="2020-01-01", endDate=""), expect=400, note="a shift cannot start before today")
call("Driver on an overlapping shift (refused)", "POST", "/api/v1/shift", OWNER,
     variant(vehicleId=V3, locations=[at(DHA, "07:30"), at(ROOTS, "08:30")], passengers=[]),
     expect=400, note="the same driver, another vehicle, but the trips overlap on the same days")
call("Vehicle on an overlapping shift (refused)", "POST", "/api/v1/shift", OWNER,
     variant(driverId=DRV2_ID, locations=[at(DHA, "07:30"), at(ROOTS, "08:30")], passengers=[]),
     expect=400, note="one vehicle cannot be on two overlapping trips")
call("Passenger on an overlapping shift (refused)", "POST", "/api/v1/shift", OWNER,
     variant(driverId=DRV2_ID, vehicleId=V2, locations=[at(DHA, "07:30"), at(ROOTS, "08:30")],
             passengers=[{"passengerId": P1_ID, "locationSequence": 1}]),
     expect=400, note="another driver and vehicle, but the passenger cannot be on two trips at once")
call("More passengers than seats (refused)", "POST", "/api/v1/shift", OWNER,
     variant(vehicleId=V3, locations=[at(BAHRIA, "10:00"), at(ROOTS, "10:30")]),
     expect=400, note="a one seat vehicle takes one passenger")
call("Passenger outside the service (refused)", "POST", "/api/v1/shift", OWNER,
     variant(vehicleId=V3, locations=[at(BAHRIA, "10:00"), at(ROOTS, "10:30")],
             passengers=[{"passengerId": P3_ID, "locationSequence": 1}]),
     expect=400, note="only approved passengers of the service can be put on its shifts")
call("Member driver builds a shift (refused)", "POST", "/api/v1/shift", DRV2,
     variant(driverId=DRV2_ID, vehicleId=V2, locations=[at(BAHRIA, "10:00"), at(ROOTS, "10:30")],
             passengers=[]),
     expect=400, note="only the owner creates shifts")

r = call("Create touching shift (member driver)", "POST", "/api/v1/shift", OWNER,
         variant(driverId=DRV2_ID, vehicleId=V2,
                 locations=[at(ROOTS, "08:00"), at(MARKAZ, "08:30")],
                 passengers=[{"passengerId": P1_ID, "locationSequence": 1}]),
         note="it starts the minute the morning run ends, touching trips do not overlap")
ENV["shift_c"] = data(r, "shiftId")
C = ENV["shift_c"]
# a trip is never invented from before its shift existed, so today's trip is set a
# couple of minutes ahead and the travel history waits for it to start
T_START_AT = pkt_now() + timedelta(minutes=2)
T_START = T_START_AT.strftime("%H:%M")
T_END = (T_START_AT + timedelta(minutes=1)).strftime("%H:%M")
if T_START_AT.date() != TODAY or T_END <= T_START:
    print("!! too close to midnight PKT for today's trip, rerun after 00:00")
r = call("Create today's shift", "POST", "/api/v1/shift", OWNER,
         variant(driverId=DRV2_ID, vehicleId=V2, daysOfWeek=[DOW_TODAY],
                 startDate=day(0), endDate=day(0),
                 locations=[at(GULBERG, T_START), at(ROOTS, T_END)],
                 passengers=[{"passengerId": P1_ID, "locationSequence": 1}]),
         note="runs once, today, so the driver update and travel history have a trip to work on")
ENV["shift_today"] = data(r, "shiftId")
T = ENV["shift_today"]

r = call("Shift detail (owner)", "GET", f"/api/v1/shift/detail?shift_id={A}", OWNER)
check("A shift is named after its vehicle number", data(r, "shift", "name") == f"PD-{SUF}A", str(data(r, "shift", "name")))
call("Service shifts", "GET", "/api/v1/shift?page=1", OWNER)
r = call("Service shifts for a day", "GET", "/api/v1/shift?page=1&day_of_week=1&search=Bahria", OWNER,
         note="search matches the vehicle number or any place on the route")
check("Searching a place finds the shifts passing it", A in ids_of(data(r, "shifts"), "id"))
r = call("Service shifts in a time window", "GET", "/api/v1/shift?page=1&start_time=06:30&end_time=08:05", OWNER,
         note="like the ride search: trips starting at or after start_time and over by end_time")
found = ids_of(data(r, "shifts"), "id")
check("The time window keeps the morning run and leaves out the later trip", A in found and C not in found)
r = call("Service shifts by vehicle number", "GET", f"/api/v1/shift?page=1&search=PD-{SUF}A", OWNER)
found = ids_of(data(r, "shifts"), "id")
check("Searching a vehicle number finds its shifts only", A in found and C not in found)
call("My shifts (owner driving)", "GET", "/api/v1/shift/mine?page=1", OWNER)
call("My shifts (member driver)", "GET", "/api/v1/shift/mine?page=1", DRV2)
call("Driver reads the shift they drive", "GET", f"/api/v1/shift/detail?shift_id={C}", DRV2)
call("Driver reads a shift they do not drive (refused)", "GET", f"/api/v1/shift/detail?shift_id={A}", DRV2,
     expect=400, note="a member driver sees only their own shifts")
call("My shifts (passenger)", "GET", "/api/v1/passenger/shifts?page=1", P1)
r = call("Passenger reads shift detail", "GET", f"/api/v1/passenger/shift/detail?shift_id={A}", P1)
check("Passenger sees their stop and no mobile numbers",
      data(r, "myStop") is not None and all(not p.get("passengerMobile") for p in (data(r, "passengers") or [])))
call("Passenger reads a shift they are not on (refused)", "GET", f"/api/v1/passenger/shift/detail?shift_id={A}", P3,
     expect=400, note="only the passengers of a shift can read it")

section("EDITING SHIFTS")
call("Edit the route", "PATCH", "/api/v1/shift", OWNER,
     {"shiftId": A,
      "locations": [at(BAHRIA, "06:55"), at(DHA, "07:20"), at(ROOTS, "07:55")]},
     note="every edit runs the clash checks again")
call("Move a shift onto another (refused)", "PATCH", "/api/v1/shift", OWNER,
     {"shiftId": C, "locations": [at(ROOTS, "07:30"), at(MARKAZ, "08:30")]},
     expect=400, note="its passenger would be on the morning run at the same time")
call("Member driver edits a shift (refused)", "PATCH", "/api/v1/shift", DRV2,
     {"shiftId": A, "endDate": D_END}, expect=400, note="only the owner edits shifts")
r = call("Remove a passenger", "PUT", "/api/v1/shift/passengers", OWNER, {"shiftId": A, "remove": [P2_ID]})
check("Removing frees the seat", data(r, "seats", "occupiedSeats") == 1)
r = call("Add and move passengers", "PUT", "/api/v1/shift/passengers", OWNER,
         {"shiftId": A, "add": [{"passengerId": P2_ID, "locationSequence": 2}],
          "move": [{"passengerId": P1_ID, "locationSequence": 3}]})
check("Adding takes the seat again", data(r, "seats", "occupiedSeats") == 2)
call("Add a passenger already on the shift (refused)", "PUT", "/api/v1/shift/passengers", OWNER,
     {"shiftId": A, "add": [{"passengerId": P1_ID, "locationSequence": 1}]},
     expect=400, note="a passenger sits on a shift once")
call("Add a passenger outside the service (refused)", "PUT", "/api/v1/shift/passengers", OWNER,
     {"shiftId": A, "add": [{"passengerId": P3_ID, "locationSequence": 1}]},
     expect=400, note="only approved passengers of the service")
call("Shrink the vehicle under a seated shift (refused)", "PATCH", "/api/v1/vehicle/update", OWNER,
     {"vehicleId": V1, "numberOfSeats": 1, "hasAC": True, "hasHeating": False, "pin": "121212"},
     expect=400, note="two passengers already ride in it on an active shift")
call("Grow the vehicle", "PATCH", "/api/v1/vehicle/update", OWNER,
     {"vehicleId": V1, "numberOfSeats": 8, "hasAC": True, "hasHeating": False, "pin": "121212"})
r = call("Shift detail after the vehicle grew", "GET", f"/api/v1/shift/detail?shift_id={A}", OWNER)
check("Active shifts take the new seat count", data(r, "shift", "seats", "totalSeats") == 8,
      json.dumps(data(r, "shift", "seats")))
call("Rename the vehicle", "PATCH", "/api/v1/vehicle/update", OWNER,
     {"vehicleId": V1, "vehicleNumber": f"PD-{SUF}R", "numberOfSeats": 8, "pin": "121212"},
     note="its active shifts and their upcoming trips take the new number as their name")
r = call("Shift detail after the vehicle was renamed", "GET", f"/api/v1/shift/detail?shift_id={A}", OWNER)
check("The shift follows its vehicle's new number", data(r, "shift", "name") == f"PD-{SUF}R", str(data(r, "shift", "name")))

section("ATTENDANCE")
r = call("Mark absent", "PATCH", "/api/v1/passenger/shift/attendance", P1,
         {"shiftId": A, "date": D1, "status": "absent"},
         note="the driver and the owner are told, the other passengers can see it")
check("Absence is recorded", data(r, "changed") is True)
r = call("Mark absent again", "PATCH", "/api/v1/passenger/shift/attendance", P1,
         {"shiftId": A, "date": D1, "status": "absent"}, note="nothing changes and nobody is told twice")
check("Marking the same status again changes nothing", data(r, "changed") is False)
call("Mark absent on a day the shift does not run (refused)", "PATCH", "/api/v1/passenger/shift/attendance", P1,
     {"shiftId": A, "date": D0, "status": "absent"}, expect=400, note="the shift starts the day after")
call("Mark absent on a shift you are not on (refused)", "PATCH", "/api/v1/passenger/shift/attendance", P3,
     {"shiftId": A, "date": D1, "status": "absent"}, expect=400)
r = call("Attendance of a trip (owner)", "GET", f"/api/v1/shift/attendance?shift_id={A}&date={D1}", OWNER)
marks = {x["passengerId"]: x["status"] for x in (data(r, "attendance") or [])}
check("Owner sees the absence", marks.get(P1_ID) == "absent" and marks.get(P2_ID) == "present", json.dumps(marks))
r = call("Attendance of a trip (other passenger)", "GET", f"/api/v1/passenger/shift/attendance?shift_id={A}&date={D1}", P2)
check("Other passengers see who is absent", any(x["passengerId"] == P1_ID and x["status"] == "absent"
                                                for x in (data(r, "attendance") or [])))
call("Attendance of a trip you do not run (refused)", "GET", f"/api/v1/shift/attendance?shift_id={A}&date={D1}", DRV2,
     expect=400, note="only the owner and the driver of the shift")
r = call("Trips of a shift", "GET", f"/api/v1/shift/occurrences?shift_id={A}&page=1", OWNER)
check("The trip counts the absence", any(x["date"] == D1 and x["absentCount"] == 1 for x in (data(r, "occurrences") or [])))

section("VEHICLES ONLY MEMBERS")
DRV3_M = "+92340" + SUF + "6"
r = call("Register vehicle owner", "POST", "/api/v1/driver/register", OPEN_TOKEN,
         {"deviceId": "flow3", "mobile": DRV3_M, "name": "Vehicle Owner", "password": "Golang@12122", "gender": "male"})
call("Verify vehicle owner otp", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": DRV3_M, "otp": data(r, "tempOTP"), "operation": "ACTIVATE_DRIVER"})
r = call("Login vehicle owner", "POST", "/api/v1/driver/login", OPEN_TOKEN,
         {"deviceId": "flow3", "mobile": DRV3_M, "password": "Golang@12122"})
ENV["drv3"] = data(r, "sessionId")
ENV["drv3_id"] = data(r, "driver", "id")
DRV3, DRV3_ID = ENV["drv3"], ENV["drv3_id"]
call("Vehicle owner pin", "POST", "/api/v1/driver/pin?pin=141414", DRV3)
r = call("Vehicle owner registers a van", "POST", "/api/v1/vehicle/register", DRV3,
         {"vehicleNumber": f"PD-{SUF}E", "vehicleInfo": "Hiace van", "numberOfSeats": 10,
          "hasAC": True, "hasHeating": False, "pin": "141414"})
V5 = ENV["vehicle5"] = data(r, "vehicleId")
r = call("Vehicle owner registers a car", "POST", "/api/v1/vehicle/register", DRV3,
         {"vehicleNumber": f"PD-{SUF}F", "vehicleInfo": "Corolla", "numberOfSeats": 4,
          "hasAC": True, "hasHeating": True, "pin": "141414"})
V6 = ENV["vehicle6"] = data(r, "vehicleId")

call("Join with vehicles only and no vehicle (refused)", "POST", "/api/v1/pickdrop/request", DRV3,
     {"serviceId": S, "joinType": "vehicles", "vehicleIds": []},
     expect=400, note="somebody who joins only through their vehicles has to bring at least one")
r = call("Join with vehicles only", "POST", "/api/v1/pickdrop/request", DRV3,
         {"serviceId": S, "joinType": "vehicles", "vehicleIds": [V5, V6]},
         note="they belong to the service through their vehicles and never drive its shifts")
ENV["driver3_request"] = data(r, "requestId")
r = call("Pending vehicle requests (vehicles only member)", "GET", "/api/v1/pickdrop/requests?type=vehicle&page=1", OWNER)
offers = {x["vehicleId"]: x for x in (data(r, "requests") or [])}
check("Offered vehicles say their owner joined with vehicles only",
      offers.get(V5, {}).get("ownerJoinType") == "vehicles" and offers.get(V6, {}).get("ownerJoinType") == "vehicles")
r = call("Approve a vehicles only member and their vehicles", "PATCH", "/api/v1/pickdrop/requests", OWNER, {"decisions": [
    {"type": "driver", "requestId": ENV["driver3_request"], "action": "approve"},
    {"type": "vehicle", "requestId": offers.get(V5, {}).get("requestId"), "action": "approve"},
    {"type": "vehicle", "requestId": offers.get(V6, {}).get("requestId"), "action": "approve"},
]})
check("The member and both vehicles are approved", len(data(r, "applied") or []) == 3)
r = call("My service (vehicles only member)", "GET", "/api/v1/pickdrop", DRV3)
check("The membership says vehicles only", data(r, "membership", "joinType") == "vehicles")
r = call("Available drivers leave out vehicles only members", "GET", "/api/v1/pickdrop/drivers/available?page=1", OWNER)
check("A vehicles only member is never offered as a driver", DRV3_ID not in ids_of(data(r, "drivers"), "driverId"))
call("Vehicles only member put behind the wheel (refused)", "POST", "/api/v1/shift", OWNER,
     variant(driverId=DRV3_ID, vehicleId=V5, locations=[at(BAHRIA, "10:00"), at(ROOTS, "10:30")], passengers=[]),
     expect=400, note="they joined with vehicles only, somebody else drives their vehicles")
r = call("Shift on the member's van, owner driving", "POST", "/api/v1/shift", OWNER,
         variant(vehicleId=V5, locations=[at(BAHRIA, "10:00"), at(ROOTS, "10:30")], passengers=[]),
         note="the vehicle's owner is told their vehicle is on a shift")
ENV["shift_v5"] = data(r, "shiftId")
r = call("Shift on the member's car at the same time, second driver", "POST", "/api/v1/shift", OWNER,
         variant(driverId=DRV2_ID, vehicleId=V6, locations=[at(DHA, "10:00"), at(ROOTS, "10:30")], passengers=[]),
         note="two vehicles of one owner on two shifts at once, each with its own driver")
ENV["shift_v6"] = data(r, "shiftId")
r = call("My shifts (vehicles only member)", "GET", "/api/v1/shift/mine?page=1", DRV3,
         note="the shifts their vehicles are on")
mine = ids_of(data(r, "shifts"), "id")
check("A vehicle owner sees every shift their vehicles are on", ENV["shift_v5"] in mine and ENV["shift_v6"] in mine)
r = call("Vehicle owner reads a shift their vehicle is on", "GET", f"/api/v1/shift/detail?shift_id={ENV['shift_v5']}", DRV3)
check("The shift names its vehicle's owner", data(r, "shift", "vehicleOwnerId") == DRV3_ID)
call("Withdraw a vehicle that is on a shift (refused)", "DELETE", f"/api/v1/pickdrop/vehicle/offer?vehicle_id={V6}", DRV3,
     expect=400, note="an active shift runs on it")
call("Vehicles only member enables a service (refused)", "POST", "/api/v1/pickdrop", DRV3,
     {"name": "Own service", "description": "", "vehicleIds": []}, expect=400, note="they already belong to a service")
call("Vehicle owner's ride on top of their vehicle's shift (refused)", "POST", "/api/v1/ride/create", DRV3,
     {"startDatetime": f"{D1} 10:05:00", "estimatedEndDatetime": f"{D1} 10:20:00", "numberOfSeats": 2,
      "startLocation": "Location A", "endLocation": "Location B", "routePoints": ["LocationA1"], "fare": 20.5,
      "routeDetails": "Clashes with the shift", "vehicleId": V5, "makeTemplate": False, "isRecurring": False,
      "frequency": 1, "period": 1, "daysOfWeek": [1]},
     expect=400, note="the vehicle is on a shift trip at that hour")

section("DRIVER UPDATES AND TRAVEL HISTORY")
r = call("Driver sends a location update", "POST", "/api/v1/shift/location", DRV2,
         {"shiftId": T, "message": "Running five minutes late", "location": "Gulberg Greens main gate",
          "lat": 33.6180, "lng": 73.1560}, note="only the passengers travelling on today's trip receive it")
check("Only the shift's passengers are told", data(r, "recipients") == 1, f"recipients {data(r, 'recipients')}")
call("Update from a driver who is not driving (refused)", "POST", "/api/v1/shift/location", OWNER,
     {"shiftId": T, "message": "Hello"}, expect=400, note="only the driver of the shift")
call("Update on a shift with no trip today (refused)", "POST", "/api/v1/shift/location", DRV2,
     {"shiftId": C, "message": "Hello"}, expect=400, note="that shift does not run today")
wait_for_clock(T_START)
call("Mark absent after the trip started (refused)", "PATCH", "/api/v1/passenger/shift/attendance", P1,
     {"shiftId": T, "date": day(0), "status": "absent"}, expect=400, note="a trip that has started cannot change")
r = call("Travel history", "GET", "/api/v1/passenger/shift/history?page=1", P1)
check("Today's trip is in the travel history", T in ids_of(data(r, "trips"), "shiftId"))

section("SHIFT REQUESTS")
r = call("Create shift request", "POST", "/api/v1/passenger/shift/request", P3,
         {"daysOfWeek": [1, 2, 3, 4, 5], "contactNumber": "+923001234567",
          "note": "A seat for my daughter", "locations": [at(G11, "07:30"), at(BLUE, "08:15")]})
ENV["shift_request"] = data(r, "requestId")
r = call("Create shift request for one service", "POST", "/api/v1/passenger/shift/request", P1,
         {"serviceId": S, "daysOfWeek": [6], "contactNumber": "+923001234568",
          "note": "Saturday classes", "locations": [at(BAHRIA, "09:00"), at(ROOTS, "10:00")]})
ENV["shift_request_addressed"] = data(r, "requestId")
call("Shift request with one place (refused)", "POST", "/api/v1/passenger/shift/request", P3,
     {"daysOfWeek": [1], "contactNumber": "+923001234567", "note": "", "locations": [at(G11, "07:30")]},
     expect=400, note="a requirement needs where from and where to")
call("My shift requests", "GET", "/api/v1/passenger/shift/requests?page=1", P3)
r = call("Search shift requests (owner)", "GET", "/api/v1/shift/requests?page=1", OWNER)
found = ids_of(data(r, "requests"), "id")
check("Owner finds open and addressed requests",
      ENV["shift_request"] in found and ENV["shift_request_addressed"] in found)
call("Search shift requests by place", "GET", "/api/v1/shift/requests?page=1&search=Blue", OWNER)
r = call("Search shift requests in a time window", "GET", "/api/v1/shift/requests?page=1&start_time=07:00&end_time=09:00", OWNER,
         note="like the ride search: requests starting at or after start_time and over by end_time")
found = ids_of(data(r, "requests"), "id")
check("The time window keeps the morning request and leaves out the later one",
      ENV["shift_request"] in found and ENV["shift_request_addressed"] not in found)
r = call("My shift requests in a time window", "GET", "/api/v1/passenger/shift/requests?page=1&start_time=07:00&end_time=09:00", P3)
check("A passenger's own list filters by time too", ENV["shift_request"] in ids_of(data(r, "requests"), "id"))
call("Delete shift request", "DELETE", f"/api/v1/passenger/shift/request?request_id={ENV['shift_request']}", P3)

section("RIDE SHARE")
ride_body = {
    "startDatetime": f"{D_RIDE} 12:00:00", "estimatedEndDatetime": f"{D_RIDE} 13:00:00",
    "numberOfSeats": 3, "startLocation": "Location A", "endLocation": "Location B",
    "routePoints": ["LocationA1", "LocationA2"], "fare": 20.5,
    "routeDetails": "Via Highway 1", "vehicleId": V1,
    "makeTemplate": True, "isRecurring": False, "frequency": 1, "period": 1, "daysOfWeek": [1]}
r = call("Create carpool ride", "POST", "/api/v1/ride/create", OWNER, ride_body)
ENV["ride"] = data(r, "id")
call("Ride with more seats than the vehicle (refused)", "POST", "/api/v1/ride/create", OWNER,
     dict(ride_body, numberOfSeats=12, makeTemplate=False), expect=400, note="the vehicle has 8 seats")
call("Ride on another driver's vehicle (refused)", "POST", "/api/v1/ride/create", OWNER,
     dict(ride_body, vehicleId=V2, makeTemplate=False), expect=400, note="a ride goes on your own vehicle")
call("Ride on top of a shift trip (refused)", "POST", "/api/v1/ride/create", OWNER,
     dict(ride_body, startDatetime=f"{D1} 07:10:00", estimatedEndDatetime=f"{D1} 07:40:00", makeTemplate=False),
     expect=400, note="the driver and the vehicle are on the morning run")
call("Shift on top of a ride (refused)", "POST", "/api/v1/shift", OWNER,
     variant(vehicleId=V3, daysOfWeek=[DOW_RIDE], startDate=D_RIDE, endDate=D_RIDE,
             locations=[at(BAHRIA, "12:10"), at(ROOTS, "12:40")], passengers=[]),
     expect=400, note="the owner drives a carpool ride at that hour")
one_seat = dict(ride_body, vehicleId=V3, numberOfSeats=1, makeTemplate=False)
call("Overlapping ride on the same vehicle (refused)", "POST", "/api/v1/ride/create", OWNER,
     dict(ride_body, startDatetime=f"{D_RIDE} 12:30:00", estimatedEndDatetime=f"{D_RIDE} 13:30:00", makeTemplate=False),
     expect=400, note="the vehicle already carries the 12:00 ride")
r = call("Same driver on another vehicle at the same time (refused)", "POST", "/api/v1/ride/create", OWNER,
         dict(one_seat, startDatetime=f"{D_RIDE} 12:30:00", estimatedEndDatetime=f"{D_RIDE} 13:30:00"),
         expect=400, note="a driver cannot drive two vehicles at once")
check("The refusal names the driver's other ride", "The driver already has a ride" in str(r.get("message")),
      str(r.get("message")))
r = call("Ride right after another ride", "POST", "/api/v1/ride/create", OWNER,
         dict(one_seat, startDatetime=f"{D_RIDE} 13:00:00", estimatedEndDatetime=f"{D_RIDE} 14:00:00"),
         note="it starts the minute the other ends, touching rides do not overlap")
ENV["ride_after"] = data(r, "id")
r = call("Recurring series landing on a ride (refused)", "POST", "/api/v1/ride/create", OWNER,
         dict(one_seat, startDatetime=f"{day(8)} 12:30:00", estimatedEndDatetime=f"{day(8)} 13:30:00",
              isRecurring=True, frequency=3, period=1),
         expect=400, note="its third ride lands on the 12:00 ride, so none of the series is created")
check("The series refusal names the clashing date", D_RIDE in str(r.get("message")), str(r.get("message")))
r = call("The refused series left nothing behind", "POST", "/api/v1/ride/create", OWNER,
         dict(one_seat, startDatetime=f"{day(8)} 12:30:00", estimatedEndDatetime=f"{day(8)} 13:30:00"),
         note="the first date of the refused series is still free")
ENV["ride_free"] = data(r, "id")
r = call("Recurring series landing on a shift trip (refused)", "POST", "/api/v1/ride/create", OWNER,
         dict(one_seat, startDatetime=f"{D0} 07:10:00", estimatedEndDatetime=f"{D0} 07:40:00",
              isRecurring=True, frequency=2, period=1),
         expect=400, note="the first day is free, the next is the morning run, and the whole series is refused")
check("The series refusal names the shift and its date",
      "shift" in str(r.get("message")) and D1 in str(r.get("message")), str(r.get("message")))
call("Recurring series that runs into itself (refused)", "POST", "/api/v1/ride/create", OWNER,
     dict(one_seat, startDatetime=f"{day(11)} 20:00:00", estimatedEndDatetime=f"{day(12)} 21:00:00",
          isRecurring=True, frequency=2, period=1),
     expect=400, note="a 25 hour ride cannot repeat every day")
r = call("Create a recurring ride series", "POST", "/api/v1/ride/create", OWNER,
         dict(one_seat, startDatetime=f"{day(11)} 18:00:00", estimatedEndDatetime=f"{day(11)} 19:00:00",
              isRecurring=True, frequency=2, period=1),
         note="every date is checked before anything is written")
ENV["series"] = data(r, "id")
r = call("Get a ride series", "GET", f"/api/v1/ride?ride_id={ENV['series']}", OPEN_TOKEN)
check("A daily series with two repeats writes both repeats", len(data(r, "childRides") or []) == 2,
      f"children {len(data(r, 'childRides') or [])}")
r = call("Cancel the ride series", "DELETE", f"/api/v1/ride/series?ride_id={ENV['series']}", OWNER)
check("Cancelling a series removes all three rides", len(data(r, "deletedRides") or []) == 3,
      f"deleted {len(data(r, 'deletedRides') or [])}")
call("Empty status leaves the ride alone", "PATCH", f"/api/v1/driver/ride/update?ride_id={ENV['ride_after']}", OWNER,
     {"status": ""}, note="only status inactive switches a ride off, through the cancellation guards and the archive")
call("The ride is still active", "GET", f"/api/v1/ride?ride_id={ENV['ride_after']}", OPEN_TOKEN)
call("Cancel a ride", "PATCH", f"/api/v1/driver/ride/update?ride_id={ENV['ride_free']}", OWNER,
     {"status": "inactive"}, note="archived, then deleted")
call("Update ride seats", "PATCH", f"/api/v1/driver/ride/update?ride_id={ENV['ride']}", OWNER, {"numberOfSeats": 2})
call("Update ride above the vehicle (refused)", "PATCH", f"/api/v1/driver/ride/update?ride_id={ENV['ride']}", OWNER,
     {"numberOfSeats": 20}, expect=400, note="the same capacity rule as creating a ride")
call("Update somebody else's ride (refused)", "PATCH", f"/api/v1/driver/ride/update?ride_id={ENV['ride']}", DRV2,
     {"numberOfSeats": 2}, expect=400, note="only the driver of a ride can change it")
call("Driver rides", "GET", "/api/v1/driver/rides?page=1&status=all", OWNER)
call("Get one ride", "GET", f"/api/v1/ride?ride_id={ENV['ride']}", OPEN_TOKEN)
r = call("Filtered rides", "GET",
         f"/api/v1/ride/filtered?page=1&search=LocationA1&start_time={quote(D_RIDE + ' 00:00:00')}"
         f"&estimated_end_time={quote(D_RIDE + ' 23:59:59')}", OPEN_TOKEN,
         note="without dates the search only looks at the next seven days, the ride is further out")
check("The search finds the carpool ride", ENV["ride"] in ids_of(data(r, "rides"), "id"))
r = call("Ride templates", "GET", "/api/v1/ride/templates", OWNER)
templates = data(r) or []
ENV["ride_template"] = next((t["id"] for t in templates if t.get("rideId") == ENV["ride"]), None)
call("Delete ride template", "DELETE", f"/api/v1/ride/template?ride_template_id={ENV['ride_template']}", OWNER)
r = call("Passenger ride request", "POST", "/api/v1/passenger/ride/request", OPEN_TOKEN, {
    "startDatetime": f"{D_RIDE} 15:30:00", "estimatedEndDatetime": f"{D_RIDE} 17:00:00",
    "numberOfSeats": 2, "startLocation": "Location A", "endLocation": "Location B",
    "routeDetails": "via gt road", "contactNumber": "+923301221121"})
ENV["ride_request"] = data(r, "requestId")
call("Get a posted ride request", "GET", f"/api/v1/ride/request?request_id={ENV['ride_request']}", OPEN_TOKEN)
call("Get ride requests", "GET", "/api/v1/driver/ride/requests?page=1", OWNER)
call("Get announcements", "GET", "/api/v1/announcements", OPEN_TOKEN)
call("Driver notifications", "GET", "/api/v1/user/notifications", OWNER)
call("Passenger notifications", "GET", "/api/v1/passenger/notifications", P1)

section("LEGACY SEAT BOOKING")
r = call("Get one ride", "GET", f"/api/v1/ride?ride_id={ENV['ride']}", OPEN_TOKEN)
ENV["ride_code"] = data(r, "ride", "code")
r = call("Driver books a seat for a phone caller", "POST", "/api/v1/driver/seat/book", OWNER, {
    "rideId": ENV["ride"], "name": "Phone Caller", "mobileNumber": "+923001112233",
    "code": ENV["ride_code"], "seats": 1, "isBook": True},
    note="the older way to fill a ride board post: the driver takes the call and enters the booking themselves")
ENV["driver_booking"] = data(r, "bookingId")
call("Driver's bookings on the ride", "GET", f"/api/v1/driver/bookings?ride_id={ENV['ride']}", OWNER)
call("Reserve (confirm) a booking", "GET", f"/api/v1/driver/booking/reserve?booking_id={ENV['driver_booking']}", OWNER)
call("Driver unbooks the seat", "POST", "/api/v1/driver/seat/book", OWNER, {
    "rideId": ENV["ride"], "mobileNumber": "+923001112233", "code": ENV["ride_code"], "isBook": False},
    note="frees it back up for someone else")
call("Passenger books a seat directly", "POST", "/api/v1/passenger/seat/book", OPEN_TOKEN, {
    "rideId": ENV["ride"], "name": "Direct Rider", "mobileNumber": "+923001112244",
    "seats": 1, "code": ENV["ride_code"]},
    note="the open-board equivalent: a rider with the ride's code books straight from the app, no driver call needed")
call("Rate the driver", "POST", "/api/v1/driver/rate", OPEN_TOKEN,
     {"driverId": OWNER_ID, "rideId": ENV["ride"], "mobileNumber": "+923001112244", "rating": 5},
     note="left after a ride, public because the rider may not hold a driver session")

section("ADMIN OVERSIGHT")
r = call("Admin overview (populated)", "GET", "/api/v1/admin/overview", adm,
         note="now counts Pick & Drop services, shifts, advertisements and open demand too")
check("The overview counts the live shift", data(r, "shifts", "live") >= 1, str(data(r, "shifts")))
r = call("Admin lists every Pick & Drop service", "GET", "/api/v1/admin/pickdrop/services?page=1", adm)
check("The service shows up with its roster", S in ids_of(data(r, "services"), "id"))
r = call("Admin searches services by name", "GET", "/api/v1/admin/pickdrop/services?page=1&search=School", adm)
check("The search finds it by name", S in ids_of(data(r, "services"), "id"))
r = call("Admin reads one service", "GET", f"/api/v1/admin/pickdrop/service?service_id={S}", adm,
         note="the same roster counts the owner sees, plus every active shift it runs")
check("Its shifts are listed", A in ids_of(data(r, "shifts"), "id"))
check("Its roster counts are populated", data(r, "counts", "approvedDrivers") >= 1, json.dumps(data(r, "counts")))
call("Read a service that does not exist (refused)", "GET", "/api/v1/admin/pickdrop/service?service_id=00000000-0000-0000-0000-000000000000", adm,
     expect=400)
r = call("Admin lists every shift on the platform", "GET", "/api/v1/admin/shifts?page=1", adm)
check("The morning run is in it", A in ids_of(data(r, "shifts"), "id"))
r = call("Admin searches shifts by service name", "GET", "/api/v1/admin/shifts?page=1&search=School", adm)
check("The search finds the service's shifts", A in ids_of(data(r, "shifts"), "id"))
r = call("Admin lists open shift requests", "GET", "/api/v1/admin/shift/requests?page=1", adm)
check("A request addressed to the service is visible platform wide",
      ENV["shift_request_addressed"] in ids_of(data(r, "requests"), "id"))
r = call("Admin lists advertisements", "GET", "/api/v1/admin/pickdrop/advertisements?page=1", adm)
check("The remaining advertisement is visible", ENV["ad1"] in ids_of(data(r, "advertisements"), "id"))
call("Admin list passengers", "GET", "/api/v1/admin/passengers?page=1", adm)
call("Admin passenger profile", "GET", f"/api/v1/admin/passenger?passenger_id={P1_ID}", adm)
call("Admin list vehicles", "GET", "/api/v1/admin/vehicles?page=1", adm)
call("Admin list drivers", "GET", "/api/v1/admin/drivers?page=1", adm)
call("Admin list rides", "GET", "/api/v1/admin/rides?page=1", adm)
call("Admin suspend passenger", "PATCH", f"/api/v1/admin/passenger/status?passenger_id={P3_ID}", adm,
     {"status": "inactive"})
call("Suspended passenger's session is refused", "GET", "/api/v1/passenger/info", P3,
     expect=401, note="suspending an account ends the sessions it already holds")
call("Suspended passenger cannot log in", "POST", "/api/v1/passenger/login", OPEN_TOKEN,
     {"deviceId": "p3", "mobile": PSG3_M, "password": "Golang@12122"},
     expect=400, note="a suspended account is refused a token")
call("Admin restore passenger", "PATCH", f"/api/v1/admin/passenger/status?passenger_id={P3_ID}", adm,
     {"status": "active"})
call("Admin suspend driver", "PATCH", f"/api/v1/admin/driver/status?driver_id={DRV2_ID}", adm,
     {"status": "inactive"})
call("Admin restore driver", "PATCH", f"/api/v1/admin/driver/status?driver_id={DRV2_ID}", adm,
     {"status": "active"})
call("Delete a driver who owns a service (refused)", "DELETE", f"/api/v1/admin/driver?user_id={OWNER_ID}", adm,
     expect=400, note="the service and its shifts would be left without an owner")
call("Admin broadcasts a notification", "POST", "/api/v1/admin/broadcast", adm,
     {"userType": 1, "title": "Fare update", "message": "Fares are unchanged this month", "notificationType": "information"})
r = call("Admin posts an announcement", "POST", "/api/v1/admin/announcement", adm,
         {"title": "Eid holiday hours", "message": "Reduced service Eid week", "type": "general", "link": ""})
call("Get announcements", "GET", "/api/v1/announcements", OPEN_TOKEN,
     note="what the announcement above looks like once it's live")
call("Get-in-touch requests", "GET", "/api/v1/admin/approch?type=contact&page=1", adm,
     note="submissions from the public get-in-touch form below")

section("CONTACT")
call("Get in touch", "POST", "/api/v1/approach", OPEN_TOKEN,
     {"name": "Interested Fleet Owner", "number": "+923001234599", "email": "owner@example.com",
      "message": "We run 12 vans in Lahore, interested in Pick & Drop", "type": "contact"},
     note="public contact form — type is contact or complain")

section("ACCOUNT LIFECYCLE")
LC_M, LC_P_M = "+92340" + SUF + "8", "+92340" + SUF + "9"
call("Lifecycle driver · Register", "POST", "/api/v1/driver/register", OPEN_TOKEN,
     {"deviceId": "lc", "mobile": LC_M, "name": "Lifecycle Driver", "password": "Golang@12122", "gender": "male"})
r = call("Resend OTP", "POST", f"/api/v1/otp/resend?mobile_number={quote(LC_M)}&otp_operation=ACTIVATE_DRIVER", OPEN_TOKEN,
         note="issues a fresh code for the same operation; the one from registration no longer verifies")
r = call("Lifecycle driver · Verify OTP", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
         {"mobile": LC_M, "otp": data(r, "tempOTP"), "operation": "ACTIVATE_DRIVER"},
         note="verified with the resent code, not the original one")
r = call("Lifecycle driver · Login", "POST", "/api/v1/driver/login", OPEN_TOKEN,
         {"deviceId": "lc", "mobile": LC_M, "password": "Golang@12122"})
LC = data(r, "sessionId")
call("Lifecycle driver · Set pin", "POST", "/api/v1/driver/pin?pin=151515", LC)

# password and pin, forgotten or just changed, are both two-step: the first call only
# sends a confirmation OTP, nothing changes until it is verified back
r = call("Driver · Forgot password", "GET", f"/api/v1/driver/password/forgot?mobile_number={quote(LC_M)}", OPEN_TOKEN,
         note="always 200 whether or not the number is registered, so a caller can't probe for accounts by it")
call("Driver · Confirm the new password", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": LC_M, "otp": data(r, "tempOTP"), "operation": "FORGOT_PASSWORD", "password": "Golang@12123"},
     note="the new password rides along with this call, nothing was cached by the step above")

call("Driver · Change password, wrong current one (refused)", "POST", "/api/v1/driver/password/reset", LC,
     {"oldPassword": "not the real password", "newPassword": "Golang@12124"}, expect=400)
r = call("Driver · Change password", "POST", "/api/v1/driver/password/reset", LC,
         {"oldPassword": "Golang@12123", "newPassword": "Golang@12124"},
         note="only sends a confirmation OTP; the password is still the one above until it's verified")
call("Driver · Confirm the password change", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": LC_M, "otp": data(r, "tempOTP"), "operation": "UPDATE_PASSWORD"},
     note="this time no password in the body, the new one was already cached by the step above")

r = call("Driver · Forgot pin", "GET", f"/api/v1/driver/pin/forgot?mobile_number={quote(LC_M)}", OPEN_TOKEN)
call("Driver · Confirm the new pin", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": LC_M, "otp": data(r, "tempOTP"), "operation": "FORGOT_PIN", "pin": "161616"})
r = call("Driver · Change pin", "POST", "/api/v1/driver/pin/reset", LC, {"oldPin": "161616", "newPin": "171717"})
call("Driver · Confirm the pin change", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": LC_M, "otp": data(r, "tempOTP"), "operation": "UPDATE_PIN"})

call("Driver · Deactivate own profile", "PATCH", "/api/v1/driver/status?status=inactive&pin=171717", LC,
     note="self-service, distinct from an admin suspension")
call("Driver · Reactivate own profile", "PATCH", "/api/v1/driver/status?status=active&pin=171717", LC)

r = call("Lifecycle passenger · Register", "POST", "/api/v1/passenger/register", OPEN_TOKEN,
         {"deviceId": "lcp", "mobile": LC_P_M, "name": "Lifecycle Passenger", "gender": "female", "password": "Golang@12122"})
call("Lifecycle passenger · Verify OTP", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": LC_P_M, "otp": data(r, "tempOTP"), "operation": "ACTIVATE_PASSENGER"})
r = call("Lifecycle passenger · Login", "POST", "/api/v1/passenger/login", OPEN_TOKEN,
         {"deviceId": "lcp", "mobile": LC_P_M, "password": "Golang@12122"})
LCP = data(r, "sessionId")
r = call("Passenger · Change password", "POST", "/api/v1/passenger/password/reset", LCP,
         {"oldPassword": "Golang@12122", "newPassword": "Golang@12123"},
         note="two-step like the driver's: a confirmation OTP first, nothing changes until it's verified")
call("Passenger · Confirm the password change", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": LC_P_M, "otp": data(r, "tempOTP"), "operation": "PASSENGER_UPDATE_PASSWORD"})
call("Passenger · Delete own profile", "DELETE", "/api/v1/passenger/delete", LCP)
call("Driver · Delete own profile", "DELETE", "/api/v1/driver/delete?pin=171717", LC)

section("LEAVING, DELETING AND HISTORY")
call("Owner leaves their own service (refused)", "DELETE", "/api/v1/pickdrop/leave", OWNER,
     expect=400, note="an owner disables the service instead")
call("Driver leaves while on a shift (refused)", "DELETE", "/api/v1/pickdrop/leave", DRV2,
     expect=400, note="they still drive active shifts")
call("Disable the service with active shifts (refused)", "DELETE", "/api/v1/pickdrop", OWNER,
     expect=400, note="delete the shifts first")
call("Passenger leaves the service", "DELETE", "/api/v1/passenger/pickdrop/leave", P2,
     note="they are taken off every shift of the service on the way out")
r = call("Shift after a passenger left", "GET", f"/api/v1/shift/detail?shift_id={A}", OWNER)
check("The leaving passenger is off the shift", P2_ID not in ids_of(data(r, "passengers"), "passengerId"))
r = call("Passenger history (owner)", "GET", f"/api/v1/shift/passengers/history?page=1&passenger_id={P2_ID}", OWNER)
check("History keeps the stints of a passenger who left",
      any(x.get("removedAt") for x in (data(r, "history") or [])))
call("Delete today's shift", "DELETE", f"/api/v1/shift?shift_id={T}", OWNER,
     note="archived, deleted, and its driver and passengers are told")
r = call("Travel history after the shift was deleted", "GET", "/api/v1/passenger/shift/history?page=1", P1)
check("Travel history outlives the shift", T in ids_of(data(r, "trips"), "shiftId"))
call("Read a deleted shift (refused)", "GET", f"/api/v1/shift/detail?shift_id={T}", OWNER,
     expect=400, note="the shift is gone, its trips remain as history")
call("Delete the touching shift", "DELETE", f"/api/v1/shift?shift_id={C}", OWNER)
call("Delete the shift on the member's van", "DELETE", f"/api/v1/shift?shift_id={ENV['shift_v5']}", OWNER)
call("Delete the shift on the member's car", "DELETE", f"/api/v1/shift?shift_id={ENV['shift_v6']}", OWNER)
call("Vehicles only member takes a vehicle out", "DELETE", f"/api/v1/pickdrop/vehicle/offer?vehicle_id={V5}", DRV3)
call("Take the last vehicle of a vehicles only member out (refused)", "DELETE",
     f"/api/v1/pickdrop/vehicle/offer?vehicle_id={V6}", DRV3,
     expect=400, note="they belong through their vehicles, they leave the service instead")
call("Vehicles only member leaves the service", "DELETE", "/api/v1/pickdrop/leave", DRV3)
call("Driver leaves the service", "DELETE", "/api/v1/pickdrop/leave", DRV2)
call("Delete shift", "DELETE", f"/api/v1/shift?shift_id={A}", OWNER)
r = call("Shift history", "GET", "/api/v1/shift/history?page=1", OWNER)
check("Deleted shifts stay in the history", any(x["id"] == A and x.get("deletedAt") for x in (data(r, "shifts") or [])))
call("Trips of the service", "GET", "/api/v1/shift/occurrences?page=1", OWNER)
call("Owner takes a vehicle out", "DELETE", f"/api/v1/pickdrop/vehicle?vehicle_id={V3}", OWNER)
call("Delete the last advertisement", "DELETE", f"/api/v1/pickdrop/advertisement?advertisement_id={ENV['ad1']}", OWNER)
call("Disable the service", "DELETE", "/api/v1/pickdrop", OWNER)
call("My service after disabling", "GET", "/api/v1/pickdrop", OWNER)
call("Admin deletes a passenger", "DELETE", f"/api/v1/admin/passenger?passenger_id={P3_ID}", adm)
call("Passenger logout", "GET", "/api/v1/passenger/logout", P1)
call("Driver logout", "GET", "/api/v1/driver/logout", OWNER)

print()
print("=" * 100)
ok = sum(1 for r in RESULTS if r["ok"])
checks_ok = sum(1 for c in CHECKS if c[1])
print(f"RESULT: {ok}/{len(RESULTS)} calls behaved as expected, {checks_ok}/{len(CHECKS)} checks held")
if FAILS:
    print(f"\n{len(FAILS)} UNEXPECTED:")
    for f in FAILS:
        print(f"  - {f[0]}: {f[1]} {f[2]} → got {f[3]}, expected {f[4]}")
        if f[5]:
            print(f"      {f[5]}")

out = os.path.join(WORK, "responses.json")
with open(out, "w") as f:
    json.dump({"results": RESULTS, "env": ENV}, f, indent=2)
print(f"\nrecorded {len(RESULTS)} calls → {out}")
sys.exit(1 if FAILS else 0)
