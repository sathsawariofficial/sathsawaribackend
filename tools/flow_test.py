#!/usr/bin/env python3
"""Drives the whole product end to end against the running server and records the
real request/response of every call, so the Postman collection can carry genuine
examples instead of invented ones."""
import json
import urllib.request
import urllib.error
import sys
from urllib.parse import quote
from datetime import datetime, timedelta

BASE = "http://localhost:5000"
OPEN_TOKEN = open("/tmp/claude-1000/-home-raotalha-Code-PersonalCode-sathsawaribackend/461379a6-41df-48ee-a93f-36ece93a6803/scratchpad/open_token.txt").read().strip()
RESULTS = []                  # every call, for the collection
ENV = {}

FAILS = []


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


def data(r, *keys):
    d = (r or {}).get("data") or {}
    for k in keys:
        if isinstance(d, dict):
            d = d.get(k)
        else:
            return None
    return d


# dates well into the future so nothing trips the 2 hour guard
D1 = (datetime.now() + timedelta(days=7)).strftime("%Y-%m-%d")
D2 = (datetime.now() + timedelta(days=8)).strftime("%Y-%m-%d")
SUF = datetime.now().strftime("%H%M%S")

print("=" * 100)
print("ADMIN")
print("=" * 100)
r = call("Admin login", "POST", "/api/v1/admin/login", OPEN_TOKEN,
         {"username": "twssawari", "password": "vR7!xK2@pQ9#Lm4$Zw8^Ty1&Nc5*Hs3%Df6!Ba"})
ENV["admin"] = data(r, "sessionId")
if not ENV["admin"]:
    print("!! admin login failed, cannot continue admin checks")

adm = ENV.get("admin")
call("Admin overview", "GET", "/api/v1/admin/overview", adm)
r = call("List roles", "GET", "/api/v1/admin/roles?page=1", adm)
roles = data(r, "roles") or []
ENV["owner_role"] = next((x["id"] for x in roles if x["name"] == "group_owner"), None)
r = call("List permissions", "GET", "/api/v1/admin/permissions?page=1", adm)
perms = {p["code"]: p["id"] for p in (data(r, "permissions") or [])}
r = call("Create role", "POST", "/api/v1/admin/role", adm,
         {"name": f"dispatcher_{SUF}", "description": "Builds shifts only"})
ENV["role"] = data(r, "id")
call("Set role permissions", "PUT", "/api/v1/admin/role/permissions", adm,
     {"roleId": ENV["role"], "permissionIds": [perms.get("shift.create"), perms.get("shift.assign_seats")]})
call("Update role", "PATCH", f"/api/v1/admin/role?role_id={ENV['role']}", adm,
     {"description": "Builds and seats shifts"})
call("Rename a system role (refused)", "PATCH", f"/api/v1/admin/role?role_id={ENV['owner_role']}", adm,
     {"name": "renamed_owner"}, expect=400, note="system roles keep their name, the group checks look them up by it")
call("Delete a system role (refused)", "DELETE", f"/api/v1/admin/role?role_id={ENV['owner_role']}", adm,
     expect=400, note="system roles cannot be deleted")

print("=" * 100)
print("ACCOUNTS")
print("=" * 100)
OWNER_M = "+92340" + SUF + "1"
DRV2_M = "+92340" + SUF + "2"
PSG1_M = "+92340" + SUF + "3"
PSG2_M = "+92340" + SUF + "4"

r = call("Register owner driver", "POST", "/api/v1/driver/register", OPEN_TOKEN,
         {"deviceId": "flow", "mobile": OWNER_M, "name": "Owner Driver",
          "password": "Golang@12122", "gender": "male"})
otp = data(r, "tempOTP")
call("Verify driver otp", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
     {"mobile": OWNER_M, "otp": otp, "operation": "ACTIVATE_DRIVER"})
r = call("Login owner driver", "POST", "/api/v1/driver/login", OPEN_TOKEN,
         {"deviceId": "flow", "mobile": OWNER_M, "password": "Golang@12122"})
ENV["owner"] = data(r, "sessionId")
ENV["owner_id"] = data(r, "driver", "id")
call("Driver profile", "GET", "/api/v1/driver/info", ENV["owner"])
call("Set pin", "POST", "/api/v1/driver/pin?pin=121212", ENV["owner"])

r = call("Register vehicle (7 seats)", "POST", "/api/v1/vehicle/register", ENV["owner"],
         {"vehicleNumber": f"FLEET-{SUF}A", "vehicleInfo": "Hiace van",
          "numberOfSeats": 7, "hasAC": True, "hasHeating": False, "pin": "121212"})
ENV["vehicle"] = data(r, "vehicleId")
call("Get vehicles", "GET", "/api/v1/vehicle/", ENV["owner"])
call("Update vehicle comfort", "PATCH", "/api/v1/vehicle/update", ENV["owner"],
     {"vehicleId": ENV["vehicle"], "numberOfSeats": 7, "hasAC": True, "hasHeating": True, "pin": "121212"})

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
         {"vehicleNumber": f"FLEET-{SUF}B", "vehicleInfo": "Corolla",
          "numberOfSeats": 4, "hasAC": True, "hasHeating": True, "pin": "131313"})
ENV["vehicle2"] = data(r, "vehicleId")

for key, mob, name, gender, dev in (("psg1", PSG1_M, "Ayesha", "female", "p1"),
                                    ("psg2", PSG2_M, "Bilal", "male", "p2")):
    r = call(f"Register passenger ({gender})", "POST", "/api/v1/passenger/register", OPEN_TOKEN,
             {"deviceId": dev, "mobile": mob, "name": name, "gender": gender, "password": "Golang@12122"})
    call("Verify passenger otp", "POST", "/api/v1/otp/verify", OPEN_TOKEN,
         {"mobile": mob, "otp": data(r, "tempOTP"), "operation": "ACTIVATE_PASSENGER"})
    r = call(f"Login passenger {name}", "POST", "/api/v1/passenger/login", OPEN_TOKEN,
             {"deviceId": dev, "mobile": mob, "password": "Golang@12122"})
    ENV[key] = data(r, "sessionId")
    ENV[key + "_id"] = data(r, "passenger", "id")

call("Passenger profile", "GET", "/api/v1/passenger/info", ENV["psg1"])
call("Passenger forgot password", "GET", f"/api/v1/passenger/password/forgot?mobile_number={quote(PSG1_M)}", OPEN_TOKEN)

print("=" * 100)
print("TRAVEL FORM")
print("=" * 100)
call("Set weekly travel form", "PUT", "/api/v1/passenger/schedule", ENV["psg1"], {"preferences": [
    {"dayOfWeek": 1, "direction": "pickup", "isEnabled": True, "location": "Bahria Town Phase 4",
     "lat": 33.5121, "lng": 73.0951, "scheduledTime": "07:15:00"},
    {"dayOfWeek": 1, "direction": "drop", "isEnabled": True, "location": "F-8 Markaz",
     "lat": 33.7101, "lng": 73.0441, "scheduledTime": "17:30:00"},
    {"dayOfWeek": 2, "direction": "pickup", "isEnabled": True, "location": "Bahria Town Phase 4",
     "lat": 33.5121, "lng": 73.0951, "scheduledTime": "07:15:00"},
    {"dayOfWeek": 2, "direction": "drop", "isEnabled": False, "location": "", "lat": 0, "lng": 0,
     "scheduledTime": ""},
]}, note="tuesday drop is switched off, this passenger rides in only in the morning")
call("Get weekly travel form", "GET", "/api/v1/passenger/schedule", ENV["psg1"])
call("Passenger 2 travel form", "PUT", "/api/v1/passenger/schedule", ENV["psg2"], {"preferences": [
    {"dayOfWeek": 1, "direction": "pickup", "isEnabled": True, "location": "DHA Phase 2",
     "lat": 33.5350, "lng": 73.1350, "scheduledTime": "07:20:00"},
]})
call("Bad day of week (refused)", "PUT", "/api/v1/passenger/schedule", ENV["psg1"],
     {"preferences": [{"dayOfWeek": 9, "direction": "pickup", "isEnabled": True,
                       "location": "X", "lat": 0, "lng": 0, "scheduledTime": "07:00:00"}]},
     expect=400, note="day of week is 1..7")

print("=" * 100)
print("GROUP")
print("=" * 100)
call("Create group with no vehicle (refused)", "POST", "/api/v1/group", ENV["owner"],
     {"name": "Empty Fleet", "description": "no vehicles", "vehicleIds": []},
     expect=400, note="a fleet has to be born with at least one vehicle")
r = call("Create group", "POST", "/api/v1/group", ENV["owner"],
         {"name": f"Morning School Fleet {SUF}", "description": "Islamabad school run",
          "vehicleIds": [ENV["vehicle"]]})
ENV["group"] = data(r, "groupId")
call("My groups", "GET", "/api/v1/group/mine?page=1", ENV["owner"])
call("Search groups (driver)", "GET", "/api/v1/group/search?page=1", ENV["drv2"])
call("Search groups (passenger)", "GET", "/api/v1/passenger/groups/search?page=1", ENV["psg1"])
call("Driver join request", "POST", "/api/v1/group/request/driver", ENV["drv2"],
     {"groupId": ENV["group"], "joinType": "both", "vehicleIds": [ENV["vehicle2"]]})
call("Duplicate join request (refused)", "POST", "/api/v1/group/request/driver", ENV["drv2"],
     {"groupId": ENV["group"], "joinType": "both", "vehicleIds": [ENV["vehicle2"]]},
     expect=400, note="a request is already waiting")
call("Passenger join request", "POST", "/api/v1/passenger/group/request", ENV["psg1"], {"groupId": ENV["group"]})
call("Passenger 2 join request", "POST", "/api/v1/passenger/group/request", ENV["psg2"], {"groupId": ENV["group"]})
call("Passenger my groups", "GET", "/api/v1/passenger/groups?page=1", ENV["psg1"])

r = call("List pending requests", "GET", f"/api/v1/group/requests?group_id={ENV['group']}", ENV["owner"])
pend = (r or {}).get("data") or {}
mem = [m for m in (pend.get("members") or []) if m["status"] == "pending"]
veh = [v for v in (pend.get("vehicles") or []) if v["status"] == "pending"]
psg = [p for p in (pend.get("passengers") or []) if p["status"] == "pending"]

decisions = []
for m in mem:
    decisions.append({"memberType": "driver", "memberId": m["id"], "action": "approve"})
for v in veh:
    decisions.append({"memberType": "vehicle", "memberId": v["id"], "action": "approve"})
for p in psg:
    decisions.append({"memberType": "passenger", "memberId": p["id"], "action": "approve"})

call("Decide requests in bulk", "PATCH", "/api/v1/group/requests", ENV["owner"],
     {"groupId": ENV["group"], "decisions": decisions})
call("Re-approve the same rows (all skipped)", "PATCH", "/api/v1/group/requests", ENV["owner"],
     {"groupId": ENV["group"], "decisions": decisions},
     note="already approved, every line comes back in skipped with the reason")

call("Promote an outsider (skipped)", "PATCH", "/api/v1/group/submanagers", ENV["owner"],
     {"groupId": ENV["group"], "promote": [ENV["psg1_id"]], "demote": []},
     note="only an approved driver-member of THIS fleet can be promoted")
call("Promote sub manager", "PATCH", "/api/v1/group/submanagers", ENV["owner"],
     {"groupId": ENV["group"], "promote": [ENV["drv2_id"]], "demote": []})
call("Group details", "GET", f"/api/v1/group?group_id={ENV['group']}", ENV["owner"])
call("Passenger travel forms (manager view)", "GET",
     f"/api/v1/group/passengers/schedules?group_id={ENV['group']}&day_of_week=1&direction=pickup&page=1", ENV["owner"])
call("Passenger token on a driver route (refused)", "GET", f"/api/v1/group?group_id={ENV['group']}", ENV["psg1"],
     expect=401, note="the middleware rejects the wrong token type before the handler is ever reached")

print("=" * 100)
print("SHIFTS")
print("=" * 100)
shift_body = {
    "groupId": ENV["group"], "vehicleId": ENV["vehicle"], "driverId": ENV["owner_id"],
    "direction": "pickup",
    "startDatetime": f"{D1} 07:00:00", "estimatedEndDatetime": f"{D1} 08:30:00",
    "startLocation": "Bahria Town Phase 4", "endLocation": "Roots School F-8",
    "routeDetails": "Via Expressway", "makeTemplate": True,
    "templateName": "Monday morning pickup", "daysOfWeek": [1, 2, 3, 4, 5],
    "stops": [
        {"location": "Bahria Town Phase 4", "lat": 33.5121, "lng": 73.0951, "scheduledTime": "07:15:00",
         "seats": [{"seatNumber": 1, "gender": "female", "passengerId": ENV["psg1_id"]}]},
        {"location": "DHA Phase 2", "lat": 33.5350, "lng": 73.1350, "scheduledTime": "07:35:00",
         "seats": [{"seatNumber": 2, "gender": "male", "passengerId": ENV["psg2_id"]}]},
        {"location": "Roots School F-8", "lat": 33.7101, "lng": 73.0441, "scheduledTime": "08:20:00",
         "seats": []},
    ],
}
r = call("Create pickup shift (+template)", "POST", "/api/v1/shift", ENV["owner"], shift_body)
ENV["shift"] = data(r, "shiftId")
ENV["template"] = data(r, "templateId")

bad = json.loads(json.dumps(shift_body))
bad["stops"][0]["seats"][0]["gender"] = "male"
bad["makeTemplate"] = False
call("Wrong gender on a seat (refused)", "POST", "/api/v1/shift", ENV["owner"], bad,
     expect=400, note="the seat's gender and its passenger must agree")

clash = json.loads(json.dumps(shift_body))
clash["startDatetime"] = f"{D1} 07:30:00"
clash["estimatedEndDatetime"] = f"{D1} 09:00:00"
clash["makeTemplate"] = False
clash["stops"] = [{"location": "Somewhere", "lat": 33.6, "lng": 73.1, "scheduledTime": "07:40:00", "seats": []}]
call("Overlapping vehicle (refused)", "POST", "/api/v1/shift", ENV["owner"], clash,
     expect=400, note="one vehicle cannot be in two places at once")

drop_body = {
    "groupId": ENV["group"], "vehicleId": ENV["vehicle"], "driverId": ENV["owner_id"],
    "direction": "drop",
    "startDatetime": f"{D1} 17:00:00", "estimatedEndDatetime": f"{D1} 18:30:00",
    "startLocation": "Roots School F-8", "endLocation": "Bahria Town Phase 4",
    "routeDetails": "Via Kashmir Highway", "makeTemplate": False, "daysOfWeek": [1],
    "stops": [
        {"location": "Roots School F-8", "lat": 33.7101, "lng": 73.0441, "scheduledTime": "17:00:00", "seats": []},
        {"location": "F-8 Markaz", "lat": 33.7101, "lng": 73.0441, "scheduledTime": "17:30:00",
         "seats": [{"seatNumber": 1, "gender": "female", "passengerId": ENV["psg1_id"]}]},
        {"location": "DHA Phase 2", "lat": 33.5350, "lng": 73.1350, "scheduledTime": "18:10:00",
         "seats": [{"seatNumber": 2, "gender": "male", "passengerId": ENV["psg2_id"]}]},
    ],
}
r = call("Create drop shift", "POST", "/api/v1/shift", ENV["owner"], drop_body)
ENV["drop_shift"] = data(r, "shiftId")

call("Shift detail", "GET", f"/api/v1/shift/detail?shift_id={ENV['shift']}", ENV["owner"])
call("Group shift roster", "GET", f"/api/v1/shift?group_id={ENV['group']}&page=1", ENV["owner"])
call("My shifts (driver)", "GET", "/api/v1/shift/mine?page=1", ENV["owner"])
call("My shifts (passenger)", "GET", "/api/v1/passenger/shifts?page=1", ENV["psg1"])
call("Passenger reads shift detail", "GET", f"/api/v1/passenger/shift/detail?shift_id={ENV['shift']}", ENV["psg1"])

call("Swap male rider for female", "PUT", "/api/v1/shift/seats", ENV["owner"], {
    "shiftId": ENV["shift"], "seats": [
        {"seatNumber": 2, "gender": "female", "passengerId": ENV["psg1_id"], "stopSequence": 1},
        {"seatNumber": 1, "gender": "male", "passengerId": ENV["psg2_id"], "stopSequence": 2},
    ]}, note="both seats change gender and occupant in one call")
call("Seat a passenger twice (refused)", "PUT", "/api/v1/shift/seats", ENV["owner"], {
    "shiftId": ENV["shift"], "seats": [
        {"seatNumber": 3, "gender": "female", "passengerId": ENV["psg1_id"], "stopSequence": 1},
    ]}, expect=400, note="they already hold seat 2 on this shift")
call("Free a seat", "PUT", "/api/v1/shift/seats", ENV["owner"], {
    "shiftId": ENV["shift"], "seats": [{"seatNumber": 1, "gender": "", "passengerId": "", "stopSequence": 0}]})
call("Reschedule the shift", "PATCH", "/api/v1/shift", ENV["owner"], {
    "shiftId": ENV["shift"],
    "startDatetime": f"{D1} 07:20:00", "estimatedEndDatetime": f"{D1} 08:50:00",
    "routeDetails": "Via Expressway, delayed for roadworks",
    "stopTimes": [{"sequenceNumber": 1, "scheduledTime": "07:35:00"},
                  {"sequenceNumber": 2, "scheduledTime": "07:55:00"},
                  {"sequenceNumber": 3, "scheduledTime": "08:40:00"}]})

sub_body = {
    "groupId": ENV["group"], "vehicleId": ENV["vehicle2"], "driverId": ENV["drv2_id"],
    "direction": "pickup",
    "startDatetime": f"{D2} 07:00:00", "estimatedEndDatetime": f"{D2} 08:30:00",
    "startLocation": "Gulberg Greens", "endLocation": "Roots School F-8",
    "routeDetails": "Via Islamabad Highway", "makeTemplate": False, "daysOfWeek": [2],
    "stops": [
        {"location": "Gulberg Greens", "lat": 33.6180, "lng": 73.1560, "scheduledTime": "07:20:00",
         "seats": [{"seatNumber": 1, "gender": "female", "passengerId": ENV["psg1_id"]}]},
        {"location": "Roots School F-8", "lat": 33.7101, "lng": 73.0441, "scheduledTime": "08:20:00", "seats": []},
    ],
}
r = call("Sub manager builds a shift", "POST", "/api/v1/shift", ENV["drv2"], sub_body)
ENV["sub_shift"] = data(r, "shiftId")
call("Sub manager decides membership (refused)", "PATCH", "/api/v1/group/requests", ENV["drv2"],
     {"groupId": ENV["group"], "decisions": [{"memberType": "driver", "memberId": ENV["drv2_id"], "action": "remove"}]},
     expect=400, note="a sub manager holds the shift permissions and none of the membership ones")

call("Get shift templates", "GET", f"/api/v1/shift/templates?group_id={ENV['group']}", ENV["owner"])

print("=" * 100)
print("RIDESHARE (existing carpool)")
print("=" * 100)
r = call("Create carpool ride", "POST", "/api/v1/ride/create", ENV["owner"], {
    "startDatetime": f"{D2} 15:30:00", "estimatedEndDatetime": f"{D2} 17:00:00",
    "numberOfSeats": 3, "startLocation": "Location A", "endLocation": "Location B",
    "routePoints": ["LocationA1", "LocationA2"], "fare": 20.5,
    "routeDetails": "Via Highway 1", "vehicleId": ENV["vehicle"],
    "makeTemplate": True, "isRecurring": False, "frequency": 1, "period": 1, "daysOfWeek": [1]})
ENV["ride"] = data(r, "id")
call("Carpool ride clashing with a shift (refused)", "POST", "/api/v1/ride/create", ENV["owner"], {
    "startDatetime": f"{D1} 07:30:00", "estimatedEndDatetime": f"{D1} 08:00:00",
    "numberOfSeats": 2, "startLocation": "Location A", "endLocation": "Location B",
    "routePoints": ["LocationA1"], "fare": 20.5, "routeDetails": "Overlaps the morning shift",
    "vehicleId": ENV["vehicle"], "makeTemplate": False, "isRecurring": False,
    "frequency": 1, "period": 1, "daysOfWeek": [1]},
    expect=400, note="the vehicle is on a fleet shift at that hour")
call("Driver rides", "GET", "/api/v1/driver/rides?page=1&status=all", ENV["owner"])
call("Get one ride", "GET", f"/api/v1/ride?ride_id={ENV['ride']}", OPEN_TOKEN)
call("Filtered rides", "GET", "/api/v1/ride/filtered?page=1&search=LocationA1", OPEN_TOKEN)
call("Ride templates", "GET", "/api/v1/ride/templates", ENV["owner"])
r = call("Passenger ride request", "POST", "/api/v1/passenger/ride/request", OPEN_TOKEN, {
    "startDatetime": f"{D2} 15:30:00", "estimatedEndDatetime": f"{D2} 17:00:00",
    "numberOfSeats": 2, "startLocation": "Location A", "endLocation": "Location B",
    "routeDetails": "via gt road", "contactNumber": "+923301221121"})
ENV["ride_request"] = data(r, "requestId")
call("Get ride requests", "GET", "/api/v1/driver/ride/requests?page=1", ENV["owner"])
call("Update ride seats", "PATCH", f"/api/v1/driver/ride/update?ride_id={ENV['ride']}", ENV["owner"],
     {"numberOfSeats": 2})
call("Get announcements", "GET", "/api/v1/announcements", OPEN_TOKEN)
call("Driver notifications", "GET", "/api/v1/user/notifications", ENV["owner"])

print("=" * 100)
print("ADMIN OVERSIGHT")
print("=" * 100)
call("Admin overview (populated)", "GET", "/api/v1/admin/overview", adm)
call("Admin list passengers", "GET", "/api/v1/admin/passengers?page=1", adm)
call("Admin passenger profile", "GET", f"/api/v1/admin/passenger?passenger_id={ENV['psg1_id']}", adm)
call("Admin list groups", "GET", "/api/v1/admin/groups?page=1", adm)
call("Admin group details", "GET", f"/api/v1/admin/group?group_id={ENV['group']}", adm)
call("Admin list shifts", "GET", "/api/v1/admin/shifts?page=1", adm)
call("Admin shift details", "GET", f"/api/v1/admin/shift?shift_id={ENV['shift']}", adm)
call("Admin list vehicles", "GET", "/api/v1/admin/vehicles?page=1", adm)
call("Admin list drivers", "GET", "/api/v1/admin/drivers?page=1", adm)
call("Admin list rides", "GET", "/api/v1/admin/rides?page=1", adm)
call("Admin suspend passenger", "PATCH", f"/api/v1/admin/passenger/status?passenger_id={ENV['psg2_id']}", adm,
     {"status": "inactive"})
call("Suspended passenger cannot log in", "POST", "/api/v1/passenger/login", OPEN_TOKEN,
     {"deviceId": "p2", "mobile": PSG2_M, "password": "Golang@12122"},
     expect=400, note="a suspended account is refused a token")
call("Admin restore passenger", "PATCH", f"/api/v1/admin/passenger/status?passenger_id={ENV['psg2_id']}", adm,
     {"status": "active"})
call("Admin suspend driver", "PATCH", f"/api/v1/admin/driver/status?driver_id={ENV['drv2_id']}", adm,
     {"status": "inactive"})
call("Admin restore driver", "PATCH", f"/api/v1/admin/driver/status?driver_id={ENV['drv2_id']}", adm,
     {"status": "active"})
call("Admin shut a fleet down", "PATCH", f"/api/v1/admin/group/status?group_id={ENV['group']}", adm,
     {"status": "inactive"})
call("Admin restore the fleet", "PATCH", f"/api/v1/admin/group/status?group_id={ENV['group']}", adm,
     {"status": "active"})
call("Delete a driver who owns a fleet (refused)", "DELETE",
     f"/api/v1/admin/driver?user_id={ENV['owner_id']}", adm,
     expect=400, note="the fleet would be left with a phantom owner")

print("=" * 100)
print("TEARDOWN")
print("=" * 100)
call("Cancel the sub manager shift", "DELETE", f"/api/v1/shift?shift_id={ENV['sub_shift']}", ENV["owner"])
call("Driver leaves the fleet", "DELETE", f"/api/v1/group/leave?group_id={ENV['group']}", ENV["drv2"])
call("Passenger leaves the fleet", "DELETE", f"/api/v1/passenger/group/leave?group_id={ENV['group']}", ENV["psg2"])
call("Delete shift template", "DELETE", f"/api/v1/shift/template?shift_template_id={ENV['template']}", ENV["owner"])
call("Delete the group with shifts (refused)", "DELETE", f"/api/v1/group?group_id={ENV['group']}", ENV["owner"],
     expect=400, note="cancel the upcoming shifts first")
call("Cancel remaining shifts", "DELETE", f"/api/v1/shift?shift_id={ENV['shift']}", ENV["owner"])
call("Cancel drop shift", "DELETE", f"/api/v1/shift?shift_id={ENV['drop_shift']}", ENV["owner"])
call("Delete the group", "DELETE", f"/api/v1/group?group_id={ENV['group']}", ENV["owner"])
call("Admin deletes a passenger", "DELETE", f"/api/v1/admin/passenger?passenger_id={ENV['psg2_id']}", adm)
call("Delete custom role", "DELETE", f"/api/v1/admin/role?role_id={ENV['role']}", adm)
call("Passenger logout", "GET", "/api/v1/passenger/logout", ENV["psg1"])
call("Driver logout", "GET", "/api/v1/driver/logout", ENV["owner"])

print()
print("=" * 100)
ok = sum(1 for r in RESULTS if r["ok"])
print(f"RESULT: {ok}/{len(RESULTS)} behaved as expected")
if FAILS:
    print(f"\n{len(FAILS)} UNEXPECTED:")
    for f in FAILS:
        print(f"  - {f[0]}: {f[1]} {f[2]} → got {f[3]}, expected {f[4]}")
        print(f"      {f[5]}")

out = "/tmp/claude-1000/-home-raotalha-Code-PersonalCode-sathsawaribackend/461379a6-41df-48ee-a93f-36ece93a6803/scratchpad/responses.json"
with open(out, "w") as f:
    json.dump({"results": RESULTS, "env": ENV}, f, indent=2)
print(f"\nrecorded {len(RESULTS)} calls → {out}")
sys.exit(1 if FAILS else 0)
