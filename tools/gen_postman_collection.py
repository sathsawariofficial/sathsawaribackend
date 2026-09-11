import json
import os
from datetime import datetime, timedelta, timezone

BASE = "{{base-url}}"
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
# the responses captured by tools/flow_test.py against a live server
RECORDED_PATH = os.path.join(ROOT, ".sath", "responses.json")

# dates the bodies start from, editable as collection variables
PKT_TODAY = (datetime.now(timezone.utc) + timedelta(hours=5)).date()
START_DATE = (PKT_TODAY + timedelta(days=2)).isoformat()
END_DATE = (PKT_TODAY + timedelta(days=16)).isoformat()
RIDE_DATE = (PKT_TODAY + timedelta(days=10)).isoformat()


def url(path, query=None):
    segs = [s for s in path.strip("/").split("/") if s]
    u = {"raw": BASE + "/" + "/".join(segs), "host": [BASE], "path": segs}
    if query:
        u["raw"] += "?" + "&".join(f"{q['key']}={q['value']}" for q in query if not q.get("disabled"))
        u["query"] = query
    return u


def q(key, value, disabled=False):
    d = {"key": key, "value": value}
    if disabled:
        d["disabled"] = True
    return d


def at(name, lat, lng, time):
    return {"location": name, "lat": lat, "lng": lng, "time": time}


BAHRIA = ("Bahria Town Phase 4", 33.5121, 73.0951)
DHA = ("DHA Phase 2", 33.5350, 73.1350)
ROOTS = ("Roots School F-8", 33.7101, 73.0441)
MARKAZ = ("F-8 Markaz", 33.7090, 73.0400)
G11 = ("G-11 Markaz", 33.6680, 72.9980)
BLUE = ("Blue Area", 33.7100, 73.0600)


# Real responses captured by driving the whole product against a live database
# (tools/flow_test.py). Keyed by the label used there, so an example is never invented.
try:
    with open(RECORDED_PATH) as _f:
        RECORDED = {r["label"]: r for r in json.load(_f)["results"]}
except Exception:
    RECORDED = {}


def example(label):
    """Build a postman saved-response block out of one recorded call."""
    rec = RECORDED.get(label)
    if not rec:
        return None

    status_text = {200: "OK", 400: "Bad Request", 401: "Unauthorized"}.get(rec["status"], "Response")
    name = ("SUCCESS" if rec["status"] == 200 else "ERROR") + " · " + label
    if rec.get("note"):
        name += " · " + rec["note"][:70]

    original = {
        "method": rec["method"],
        "header": [],
        "url": url(rec["path"].split("?")[0],
                   [q(k, v) for k, v in (
                       [kv.split("=", 1) for kv in rec["path"].split("?", 1)[1].split("&")]
                       if "?" in rec["path"] else [])]),
    }
    if rec.get("request") is not None:
        original["body"] = {
            "mode": "raw",
            "raw": json.dumps(rec["request"], indent=4),
            "options": {"raw": {"language": "json"}},
        }

    return {
        "name": name,
        "originalRequest": original,
        "status": status_text,
        "code": rec["status"],
        "_postman_previewlanguage": "json",
        "header": [{"key": "Content-Type", "value": "application/json; charset=utf-8"}],
        "cookie": [],
        "body": json.dumps(rec["response"], indent=4),
    }


def req(name, method, path, token, body=None, query=None, tests=None, desc=None, examples=None):
    """One request. examples are the flow_test labels whose recorded responses it
    carries, typically the success plus the refusals that prove its rules."""
    r = {
        "auth": {"type": "bearer", "bearer": [{"key": "token", "value": "{{" + token + "}}", "type": "string"}]},
        "method": method,
        "header": [],
        "url": url(path, query),
    }
    if body is not None:
        r["body"] = {
            "mode": "raw",
            "raw": json.dumps(body, indent=4),
            "options": {"raw": {"language": "json"}},
        }
    if desc:
        r["description"] = desc

    saved = []
    for label in (examples or [name]):
        block = example(label)
        if block:
            saved.append(block)

    item = {"name": name, "request": r, "response": saved}
    if tests:
        item["event"] = [{
            "listen": "test",
            "script": {"exec": tests, "type": "text/javascript", "packages": {}},
        }]
    return item


def save(var, expr, label=None):
    return [
        "const r = pm.response.json();",
        f"if (r.data && ({expr}) !== undefined && ({expr}) !== null) {{",
        f'    pm.environment.set("{var}", {expr});',
        f'    console.log("{label or var} saved:", {expr});',
        "} else {",
        f'    console.error("{var} not found in response");',
        "}",
    ]


def save_from_list(var, list_expr, match_field, match_var, value_field):
    """Picks one row out of a list response by an id already in the environment."""
    return [
        "const r = pm.response.json();",
        f"const rows = (r.data && {list_expr}) || [];",
        f'const row = rows.find(x => x.{match_field} === pm.environment.get("{match_var}"));',
        "if (row) {",
        f'    pm.environment.set("{var}", row.{value_field});',
        f'    console.log("{var} saved:", row.{value_field});',
        "} else {",
        f'    console.error("{var} not found in response");',
        "}",
    ]


def driver_account(prefix, label, name, mobile, device, pin, session_var, id_var, otp_var):
    return [
        req(f"{prefix} · Register", "POST", "/api/v1/driver/register", "open_token",
            body={"deviceId": device, "mobile": mobile, "name": name, "password": "Golang@12122", "gender": "male"},
            tests=save(otp_var, "r.data.tempOTP"), examples=[f"Register {label}"]),
        req(f"{prefix} · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
            body={"mobile": mobile, "otp": "{{" + otp_var + "}}", "operation": "ACTIVATE_DRIVER"},
            examples=["Verify driver otp" if label == "owner driver" else "Verify driver 2 otp"]),
        req(f"{prefix} · Login", "POST", "/api/v1/driver/login", "open_token",
            body={"deviceId": device, "mobile": mobile, "password": "Golang@12122"},
            tests=save(session_var, "r.data.sessionId") + save(id_var, "r.data.driver.id"),
            examples=["Login owner driver" if label == "owner driver" else "Login second driver"]),
        req(f"{prefix} · Set Pin", "POST", "/api/v1/driver/pin", session_var,
            query=[q("pin", pin)], examples=["Set pin" if label == "owner driver" else "Second driver pin"]),
    ]


def passenger_account(prefix, name, gender, mobile, device, session_var, id_var, otp_var):
    return [
        req(f"{prefix} · Register", "POST", "/api/v1/passenger/register", "open_token",
            body={"deviceId": device, "mobile": mobile, "name": name, "gender": gender, "password": "Golang@12122"},
            tests=save(otp_var, "r.data.tempOTP"), examples=[f"Register passenger {name}"]),
        req(f"{prefix} · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
            body={"mobile": mobile, "otp": "{{" + otp_var + "}}", "operation": "ACTIVATE_PASSENGER"},
            examples=[f"Verify passenger {name} otp"]),
        req(f"{prefix} · Login", "POST", "/api/v1/passenger/login", "open_token",
            body={"deviceId": device, "mobile": mobile, "password": "Golang@12122"},
            tests=save(session_var, "r.data.sessionId") + save(id_var, "r.data.passenger.id"),
            examples=[f"Login passenger {name}"]),
    ]


# ---------------------------------------------------------------- 0 admin ----
admin = {"name": "0 · Admin Console", "item": [
    req("Admin Login", "POST", "/api/v1/admin/login", "open_token",
        body={"username": "twssawari", "password": "vR7!xK2@pQ9#Lm4$Zw8^Ty1&Nc5*Hs3%Df6!Ba"},
        tests=save("admin_session", "r.data.sessionId"), examples=["Admin login"]),

    req("Platform Overview", "GET", "/api/v1/admin/overview", "admin_session",
        desc=("Drivers, passengers, vehicles, carpool rides, Pick & Drop services and shifts, each as a total "
              "with the live slice of it, plus advertisements, open shift requests and join requests still "
              "waiting on a decision. Pick & Drop stays run by each service's owner, the admin only ever looks."),
        examples=["Admin overview (populated)", "Admin overview"]),

    req("List Passengers", "GET", "/api/v1/admin/passengers", "admin_session",
        query=[q("page", "1"), q("search", "", True), q("status", "", True)],
        desc="search matches the name or the mobile number, status filters active, inactive or pending.",
        examples=["Admin list passengers"]),

    req("Passenger Profile", "GET", "/api/v1/admin/passenger", "admin_session",
        query=[q("passenger_id", "{{passenger1_id}}")], examples=["Admin passenger profile"]),

    req("Suspend A Passenger", "PATCH", "/api/v1/admin/passenger/status", "admin_session",
        query=[q("passenger_id", "{{passenger3_id}}")], body={"status": "inactive"},
        desc="A suspended account cannot log in, nothing it was part of is destroyed. Send active to restore it.",
        examples=["Admin suspend passenger", "Suspended passenger's session is refused",
                  "Suspended passenger cannot log in", "Admin restore passenger"]),

    req("Suspend A Driver", "PATCH", "/api/v1/admin/driver/status", "admin_session",
        query=[q("driver_id", "{{driver2_id}}")], body={"status": "inactive"},
        examples=["Admin suspend driver", "Admin restore driver"]),

    req("Delete A Driver", "DELETE", "/api/v1/admin/driver", "admin_session",
        query=[q("user_id", "{{owner_id}}")],
        desc="Refused while the driver owns a Pick & Drop service or drives an active shift.",
        examples=["Delete a driver who owns a service (refused)"]),

    req("Delete A Passenger", "DELETE", "/api/v1/admin/passenger", "admin_session",
        query=[q("passenger_id", "{{passenger3_id}}")],
        desc=("Archives the account, takes it off its active shifts, ends its Pick & Drop membership and "
              "archives its availability and shift requests. Its travel history stays."),
        examples=["Admin deletes a passenger"]),

    req("List Vehicles", "GET", "/api/v1/admin/vehicles", "admin_session", query=[q("page", "1")],
        examples=["Admin list vehicles"]),
    req("List Drivers", "GET", "/api/v1/admin/drivers", "admin_session", query=[q("page", "1")],
        examples=["Admin list drivers"]),
    req("List Rides", "GET", "/api/v1/admin/rides", "admin_session", query=[q("page", "1")],
        examples=["Admin list rides"]),

    req("Broadcast A Notification", "POST", "/api/v1/admin/broadcast", "admin_session",
        body={"userType": 1, "title": "Fare update", "message": "Fares are unchanged this month",
              "notificationType": "information"},
        desc="userType is 1 (driver) or 2 (passenger). notificationType is information or marketing.",
        examples=["Admin broadcasts a notification"]),

    req("Post An Announcement", "POST", "/api/v1/admin/announcement", "admin_session",
        body={"title": "Eid holiday hours", "message": "Reduced service Eid week", "type": "general", "link": ""},
        desc="Shows up on GET /api/v1/announcements, the public board every app reads on launch.",
        examples=["Admin posts an announcement"]),

    req("Get-In-Touch Requests", "GET", "/api/v1/admin/approch", "admin_session",
        query=[q("type", "contact"), q("page", "1")],
        desc="Submissions from the public get-in-touch form. type is contact or complain.",
        examples=["Get-in-touch requests"]),

    req("List Pick & Drop Services", "GET", "/api/v1/admin/pickdrop/services", "admin_session",
        query=[q("page", "1"), q("search", "", True), q("status", "", True)],
        desc=("Every service on the platform, active or disabled, with the same roster counts an owner sees on "
              "their own. Read only — Pick & Drop stays run by each service's owner."),
        examples=["Admin lists every Pick & Drop service", "Admin searches services by name"]),
    req("Pick & Drop Service Detail", "GET", "/api/v1/admin/pickdrop/service", "admin_session",
        query=[q("service_id", "{{service_id}}")],
        desc="The service, its full roster breakdown (approved by kind, pending, vehicles-only counted apart), and every active shift it runs.",
        examples=["Admin reads one service", "Read a service that does not exist (refused)"]),
    req("List Advertisements", "GET", "/api/v1/admin/pickdrop/advertisements", "admin_session",
        query=[q("page", "1"), q("search", "", True)],
        examples=["Admin lists advertisements"]),
    req("List Shifts", "GET", "/api/v1/admin/shifts", "admin_session",
        query=[q("page", "1"), q("search", "", True), q("day_of_week", "", True),
               q("start_time", "", True), q("end_time", "", True), q("status", "", True)],
        desc="Every shift on the platform, in any service, filtered exactly like the owner's own shift search. status is active, completed or all (default all).",
        examples=["Admin lists every shift on the platform", "Admin searches shifts by service name"]),
    req("List Open Shift Requests", "GET", "/api/v1/admin/shift/requests", "admin_session",
        query=[q("page", "1"), q("search", "", True), q("day_of_week", "", True),
               q("start_time", "", True), q("end_time", "", True)],
        desc="Every rider requirement not yet met by a shift, addressed to one service or still open to any of them.",
        examples=["Admin lists open shift requests"]),
]}

# ------------------------------------------------------------- 1 accounts ----
accounts = {"name": "1 · Accounts", "item": (
    driver_account("Owner Driver", "owner driver", "Owner Driver", "+923400000001", "flow", "121212",
                   "owner_session", "owner_id", "owner_otp") + [
        req("Owner Driver · Profile", "GET", "/api/v1/driver/info", "owner_session", examples=["Driver profile"]),
        req("Owner Driver · Register Vehicle (7 seats)", "POST", "/api/v1/vehicle/register", "owner_session",
            body={"vehicleNumber": "PD-0001A", "vehicleInfo": "Hiace van", "numberOfSeats": 7,
                  "hasAC": True, "hasHeating": False, "pin": "121212"},
            tests=save("vehicle_id", "r.data.vehicleId"),
            desc="A vehicle needs a seat count before it can go on a shift, and a ride can never offer more seats than it has.",
            examples=["Register vehicle (7 seats)"]),
        req("Owner Driver · Register Small Vehicle (1 seat)", "POST", "/api/v1/vehicle/register", "owner_session",
            body={"vehicleNumber": "PD-0001C", "vehicleInfo": "Mehran", "numberOfSeats": 1,
                  "hasAC": False, "hasHeating": False, "pin": "121212"},
            tests=save("vehicle3_id", "r.data.vehicleId"), examples=["Register small vehicle (1 seat)"]),
        req("Owner Driver · Get Vehicles", "GET", "/api/v1/vehicle/", "owner_session", examples=["Get vehicles"]),
        req("Update Vehicle (seats / number)", "PATCH", "/api/v1/vehicle/update", "owner_session",
            body={"vehicleId": "{{vehicle_id}}", "numberOfSeats": 8, "hasAC": True, "hasHeating": False, "pin": "121212"},
            desc=("Every active shift on the vehicle takes the new seat count, and the vehicle cannot drop below "
                  "the passengers an active shift already carries in it. A shift is named after its vehicle, so "
                  "a new vehicleNumber renames its active shifts and their upcoming trips."),
            examples=["Grow the vehicle", "Shrink the vehicle under a seated shift (refused)", "Rename the vehicle"]),
    ] +
    driver_account("Second Driver", "second driver", "Second Driver", "+923400000002", "flow2", "131313",
                   "driver2_session", "driver2_id", "driver2_otp") + [
        req("Second Driver · Register Vehicle (4 seats)", "POST", "/api/v1/vehicle/register", "driver2_session",
            body={"vehicleNumber": "PD-0001B", "vehicleInfo": "Corolla", "numberOfSeats": 4,
                  "hasAC": True, "hasHeating": True, "pin": "131313"},
            tests=save("vehicle2_id", "r.data.vehicleId"), examples=["Register second vehicle"]),
        req("Second Driver · Register Another Vehicle", "POST", "/api/v1/vehicle/register", "driver2_session",
            body={"vehicleNumber": "PD-0001D", "vehicleInfo": "Cultus", "numberOfSeats": 4,
                  "hasAC": True, "hasHeating": False, "pin": "131313"},
            tests=save("vehicle4_id", "r.data.vehicleId"), examples=["Register fourth vehicle"]),
        req("Vehicle Owner · Register", "POST", "/api/v1/driver/register", "open_token",
            body={"deviceId": "flow3", "mobile": "+923400000006", "name": "Vehicle Owner",
                  "password": "Golang@12122", "gender": "male"},
            tests=save("driver3_otp", "r.data.tempOTP"),
            desc="A driver who will join a service with their vehicles only.",
            examples=["Register vehicle owner"]),
        req("Vehicle Owner · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
            body={"mobile": "+923400000006", "otp": "{{driver3_otp}}", "operation": "ACTIVATE_DRIVER"},
            examples=["Verify vehicle owner otp"]),
        req("Vehicle Owner · Login", "POST", "/api/v1/driver/login", "open_token",
            body={"deviceId": "flow3", "mobile": "+923400000006", "password": "Golang@12122"},
            tests=save("driver3_session", "r.data.sessionId") + save("driver3_id", "r.data.driver.id"),
            examples=["Login vehicle owner"]),
        req("Vehicle Owner · Set Pin", "POST", "/api/v1/driver/pin", "driver3_session",
            query=[q("pin", "141414")], examples=["Vehicle owner pin"]),
        req("Vehicle Owner · Register A Van", "POST", "/api/v1/vehicle/register", "driver3_session",
            body={"vehicleNumber": "PD-0001E", "vehicleInfo": "Hiace van", "numberOfSeats": 10,
                  "hasAC": True, "hasHeating": False, "pin": "141414"},
            tests=save("vehicle5_id", "r.data.vehicleId"), examples=["Vehicle owner registers a van"]),
        req("Vehicle Owner · Register A Car", "POST", "/api/v1/vehicle/register", "driver3_session",
            body={"vehicleNumber": "PD-0001F", "vehicleInfo": "Corolla", "numberOfSeats": 4,
                  "hasAC": True, "hasHeating": True, "pin": "141414"},
            tests=save("vehicle6_id", "r.data.vehicleId"), examples=["Vehicle owner registers a car"]),
    ] +
    passenger_account("Passenger 1", "Ayesha", "female", "+923400000003", "p1",
                      "passenger_session", "passenger1_id", "passenger1_otp") +
    passenger_account("Passenger 2", "Bilal", "male", "+923400000004", "p2",
                      "passenger2_session", "passenger2_id", "passenger2_otp") +
    passenger_account("Passenger 3", "Sana", "female", "+923400000005", "p3",
                      "passenger3_session", "passenger3_id", "passenger3_otp") + [
        req("Passenger · Profile", "GET", "/api/v1/passenger/info", "passenger_session", examples=["Passenger profile"]),
        req("Passenger · Forgot Password", "GET", "/api/v1/passenger/password/forgot", "open_token",
            query=[q("mobile_number", "%2B923400000003")], examples=["Passenger forgot password"]),
        req("Passenger · Logout", "GET", "/api/v1/passenger/logout", "passenger_session", examples=["Passenger logout"]),
        req("Driver · Logout", "GET", "/api/v1/driver/logout", "owner_session", examples=["Driver logout"]),
    ]
)}

# ------------------------------------------------------ 2 pick & drop owner ----
owner = {"name": "2 · Pick & Drop · Owner", "item": [
    req("Enable Pick & Drop", "POST", "/api/v1/pickdrop", "owner_session",
        body={"name": "Islamabad School Run", "description": "Morning and afternoon school runs",
              "vehicleIds": ["{{vehicle_id}}"]},
        tests=save("service_id", "r.data.serviceId"),
        desc=("The driver becomes the one and only owner of the service. vehicleIds is optional, vehicles can "
              "be added later. A driver owns or belongs to at most one service, so this is refused for an "
              "owner, a member, or a driver with a request waiting."),
        examples=["Enable Pick & Drop", "Enable a second service (refused)",
                  "Enable a service while a request is open (refused)", "Vehicles only member enables a service (refused)"]),

    req("My Service", "GET", "/api/v1/pickdrop", "owner_session",
        desc=("role is owner, driver or none. An owner gets the counts, including the advertisement limit, "
              "which is the approved vehicles times two, worked out on every read."),
        examples=["My service (owner)", "My service after adding a vehicle", "My service after disabling",
                  "Passenger token on a driver route (refused)"]),

    req("Add Own Vehicles", "POST", "/api/v1/pickdrop/vehicles", "owner_session",
        body={"vehicleIds": ["{{vehicle3_id}}"]},
        desc="The owner's own vehicles go straight in, approved.",
        examples=["Add another own vehicle", "Member driver adds vehicles as owner (refused)"]),

    req("Remove A Vehicle", "DELETE", "/api/v1/pickdrop/vehicle", "owner_session",
        query=[q("vehicle_id", "{{vehicle3_id}}")],
        desc="Refused while the vehicle is on an active shift.", examples=["Owner takes a vehicle out"]),

    req("Join Requests · Drivers", "GET", "/api/v1/pickdrop/requests", "owner_session",
        query=[q("type", "driver"), q("page", "1"), q("status", "", True)],
        tests=save_from_list("driver2_request_id", "r.data.requests", "driverId", "driver2_id", "requestId"),
        desc="status defaults to pending, send all or any membership status to see decided ones.",
        examples=["Pending driver requests"]),

    req("Join Requests · Vehicles", "GET", "/api/v1/pickdrop/requests", "owner_session",
        query=[q("type", "vehicle"), q("page", "1")],
        tests=save_from_list("vehicle2_request_id", "r.data.requests", "vehicleId", "vehicle2_id", "requestId"),
        desc=("ownerJoinType is owner, driver or vehicles. A vehicle whose owner joined with vehicles only is "
              "always driven by somebody else."),
        examples=["Pending vehicle requests", "Pending vehicle requests (vehicles only member)"]),

    req("Join Requests · Passengers", "GET", "/api/v1/pickdrop/requests", "owner_session",
        query=[q("type", "passenger"), q("page", "1")],
        tests=(save_from_list("passenger1_request_id", "r.data.requests", "passengerId", "passenger1_id", "requestId") +
               ["{"] + save_from_list("passenger2_request_id", "r.data.requests", "passengerId", "passenger2_id", "requestId") + ["}"]),
        desc="Each passenger comes with the weekly demand they filled in, once they have.",
        examples=["Pending passenger requests"]),

    req("Join Request Detail", "GET", "/api/v1/pickdrop/request", "owner_session",
        query=[q("type", "driver"), q("request_id", "{{driver2_request_id}}")], examples=["Join request detail"]),

    req("Decide Requests (bulk)", "PATCH", "/api/v1/pickdrop/requests", "owner_session",
        body={"decisions": [
            {"type": "driver", "requestId": "{{driver2_request_id}}", "action": "approve"},
            {"type": "vehicle", "requestId": "{{vehicle2_request_id}}", "action": "approve"},
            {"type": "passenger", "requestId": "{{passenger1_request_id}}", "action": "approve"},
            {"type": "passenger", "requestId": "{{passenger2_request_id}}", "action": "approve"},
        ]},
        desc=("approve, reject or remove, for drivers, vehicles and passengers in one call. Drivers are settled "
              "first, then vehicles, then passengers, whatever order they are sent in. A line that cannot be "
              "applied comes back in skipped with the reason instead of failing the call: a vehicle whose "
              "driver is not approved, a request already decided, removing somebody who is not approved. "
              "Removing a passenger takes them off the service's shifts, removing a driver or vehicle is "
              "refused while they are on an active shift."),
        examples=["Decide requests in bulk", "Approve a vehicle before its driver (skipped)",
                  "Remove a rejected passenger (skipped)", "Re-decide the same requests (all skipped)",
                  "Approve a vehicles only member and their vehicles"]),

    req("Available Drivers", "GET", "/api/v1/pickdrop/drivers/available", "owner_session",
        query=[q("page", "1"), q("days_of_week", "1,2,3,4,5,6,7", True), q("start_date", "{{start_date}}", True),
               q("end_date", "", True), q("start_time", "07:00", True), q("end_time", "08:00", True),
               q("exclude_shift_id", "", True)],
        desc=("The owner and the approved drivers, never a member who joined with vehicles only. Send the "
              "schedule of the shift you are about to build, all of days_of_week, start_date, start_time and "
              "end_time, and each driver says isAvailable and, if not, the conflict."),
        examples=["Available drivers", "Available drivers for a schedule", "Available drivers leave out vehicles only members"]),

    req("Available Vehicles", "GET", "/api/v1/pickdrop/vehicles/available", "owner_session",
        query=[q("page", "1"), q("days_of_week", "", True), q("start_date", "", True),
               q("start_time", "", True), q("end_time", "", True)],
        desc=("Approved vehicles with hasValidSeats, a vehicle without seats cannot go on a shift. ownerJoinType "
              "vehicles means the vehicle's owner does not drive, pick another driver for it."),
        examples=["Available vehicles"]),

    req("Available Passengers (weekly demand)", "GET", "/api/v1/pickdrop/passengers/available", "owner_session",
        query=[q("page", "1"), q("day_of_week", "", True)],
        desc="Approved passengers with the days and places they need the service.",
        examples=["Available passengers with weekly demand"]),

    req("Create Advertisement", "POST", "/api/v1/pickdrop/advertisement", "owner_session",
        body={"title": "Bahria to F-8 school run", "description": "Morning pick up in an AC van", "fare": 6000,
              "daysOfWeek": [1, 2, 3, 4, 5], "startTime": "06:30", "endTime": "08:30",
              "locations": [at(*BAHRIA, "06:45"), at(*DHA, "07:10"), at(*ROOTS, "08:00")]},
        tests=save("advertisement_id", "r.data.advertisementId"),
        desc=("Up to ten places, each with the time it is reached, all between startTime and endTime. An owner "
              "may hold approved vehicles times two advertisements."),
        examples=["Create advertisement", "Create second advertisement", "Advertisement over the limit (refused)",
                  "Advertisement with eleven places (refused)"]),

    req("My Advertisements", "GET", "/api/v1/pickdrop/advertisements", "owner_session",
        query=[q("page", "1")], examples=["My advertisements"]),

    req("Delete Advertisement", "DELETE", "/api/v1/pickdrop/advertisement", "owner_session",
        query=[q("advertisement_id", "{{advertisement_id}}")],
        desc="Archived to del_pick_drop_advertisements, then deleted.",
        examples=["Delete advertisement", "Delete the last advertisement"]),

    req("Leave (owner)", "DELETE", "/api/v1/pickdrop/leave", "owner_session",
        desc="An owner cannot leave, they disable the service instead.",
        examples=["Owner leaves their own service (refused)"]),

    req("Disable Pick & Drop", "DELETE", "/api/v1/pickdrop", "owner_session",
        desc=("Refused while the service has active shifts. Otherwise every open request and membership is "
              "closed and its advertisements archived, so its drivers and passengers are free to join another."),
        examples=["Disable the service", "Disable the service with active shifts (refused)"]),
]}

# ----------------------------------------------------- 3 pick & drop driver ----
member = {"name": "3 · Pick & Drop · Joining Driver", "item": [
    req("Find Services", "GET", "/api/v1/pickdrop/search", "driver2_session",
        query=[q("page", "1"), q("search", "School", True)], examples=["Search services (driver)"]),

    req("Request To Join", "POST", "/api/v1/pickdrop/request", "driver2_session",
        body={"serviceId": "{{service_id}}", "joinType": "driver", "vehicleIds": ["{{vehicle2_id}}"]},
        desc=("joinType driver (the default) joins as somebody who can drive the service's shifts. vehicleIds "
              "is optional, each proposed vehicle becomes its own request the owner approves separately, after "
              "the driver."),
        examples=["Driver join request (with a vehicle)", "Duplicate join request (refused)",
                  "Owner asks to join a service (refused)"]),

    req("Request To Join (vehicles only)", "POST", "/api/v1/pickdrop/request", "driver3_session",
        body={"serviceId": "{{service_id}}", "joinType": "vehicles", "vehicleIds": ["{{vehicle5_id}}", "{{vehicle6_id}}"]},
        desc=("joinType vehicles joins only through the vehicles offered, at least one. The driver belongs to the "
              "service but is never put behind the wheel of its shifts and never listed as an available driver. "
              "The owner approves the driver, then each vehicle, and puts the vehicles on shifts with other "
              "drivers, different vehicles on different shifts. The last vehicle cannot be taken out, the member "
              "leaves the service instead. Belonging to one service still rules out owning or joining another."),
        examples=["Join with vehicles only", "Join with vehicles only and no vehicle (refused)"]),

    req("My Service (member)", "GET", "/api/v1/pickdrop", "driver2_session",
        desc="membership.joinType is driver or vehicles.",
        examples=["My service (member driver)", "My service (vehicles only member)"]),

    req("Offer Vehicles", "POST", "/api/v1/pickdrop/vehicles/offer", "driver2_session",
        body={"vehicleIds": ["{{vehicle4_id}}"]},
        desc="A member driver proposes more of their vehicles, each waits for the owner.",
        examples=["Offer another vehicle"]),

    req("Withdraw A Vehicle", "DELETE", "/api/v1/pickdrop/vehicle/offer", "driver2_session",
        query=[q("vehicle_id", "{{vehicle4_id}}")],
        desc=("Withdraws a waiting offer, or takes an approved vehicle out when it is on no active shift. A member "
              "who joined with vehicles only keeps at least one."),
        examples=["Withdraw the offered vehicle", "Vehicles only member takes a vehicle out",
                  "Withdraw a vehicle that is not offered (refused)", "Withdraw a vehicle that is on a shift (refused)",
                  "Take the last vehicle of a vehicles only member out (refused)"]),

    req("Leave The Service", "DELETE", "/api/v1/pickdrop/leave", "driver2_session",
        desc="Refused while the driver, or a vehicle they brought, is on an active shift.",
        examples=["Driver leaves the service", "Vehicles only member leaves the service",
                  "Driver leaves while on a shift (refused)"]),
]}

# -------------------------------------------------- 4 pick & drop passenger ----
passenger = {"name": "4 · Pick & Drop · Passenger", "item": [
    req("Search Advertisements (public)", "GET", "/api/v1/pickdrop/advertisements/search", "open_token",
        query=[q("page", "1"), q("search", "Bahria", True), q("day_of_week", "1", True),
               q("start_time", "", True), q("end_time", "", True)],
        desc="Open to anybody. search matches the advertised places, day_of_week and the time window narrow it further.",
        examples=["Search advertisements (public)"]),

    req("Find Services", "GET", "/api/v1/passenger/pickdrop/search", "passenger_session",
        query=[q("page", "1"), q("search", "School", True)], examples=["Search services (passenger)"]),

    req("Request To Join", "POST", "/api/v1/passenger/pickdrop/request", "passenger_session",
        body={"serviceId": "{{service_id}}"},
        desc="A passenger belongs to at most one service and holds one open request at a time.",
        examples=["Passenger Ayesha join request", "Passenger duplicate join request (refused)"]),

    req("Passenger 2 · Request To Join", "POST", "/api/v1/passenger/pickdrop/request", "passenger2_session",
        body={"serviceId": "{{service_id}}"}, examples=["Passenger Bilal join request"]),

    req("Passenger 3 · Request To Join", "POST", "/api/v1/passenger/pickdrop/request", "passenger3_session",
        body={"serviceId": "{{service_id}}"}, examples=["Passenger Sana join request"]),

    req("My Membership", "GET", "/api/v1/passenger/pickdrop", "passenger_session",
        examples=["Passenger membership (approved)", "Passenger membership (pending)"]),

    req("Set Weekly Availability", "PUT", "/api/v1/passenger/availability", "passenger_session",
        body={"days": [
            {"dayOfWeek": 1, "isRequired": True, "locations": [at(*BAHRIA, "06:55"), at(*ROOTS, "13:30")]},
            {"dayOfWeek": 2, "isRequired": True, "locations": [at(*BAHRIA, "06:55")]},
            {"dayOfWeek": 3, "isRequired": False, "locations": []},
        ]},
        desc=("Open once the owner has approved the passenger. Each day is required or not, a required day "
              "carries one to six places each with a time. Days not sent stay as they were."),
        examples=["Set weekly availability", "Availability before approval (refused)",
                  "Seven places in a day (refused)", "Places on a day that is not required (refused)"]),

    req("Get Weekly Availability", "GET", "/api/v1/passenger/availability", "passenger_session",
        desc="Always the whole week, Monday (1) to Sunday (7).", examples=["Get weekly availability"]),

    req("Passenger 2 · Set Weekly Availability", "PUT", "/api/v1/passenger/availability", "passenger2_session",
        body={"days": [{"dayOfWeek": d, "isRequired": True, "locations": [at(*DHA, "07:20")]} for d in (1, 2, 3, 4, 5)]},
        examples=["Passenger 2 weekly availability"]),

    req("Leave The Service", "DELETE", "/api/v1/passenger/pickdrop/leave", "passenger2_session",
        desc="Withdraws a waiting request, or leaves: the passenger is taken off every active shift of the service and the owner is told.",
        examples=["Passenger leaves the service"]),
]}

# -------------------------------------------------------- 5 shifts owner ----
shift_body = {
    "driverId": "{{owner_id}}", "vehicleId": "{{vehicle_id}}",
    "daysOfWeek": [1, 2, 3, 4, 5, 6, 7], "startDate": "{{start_date}}", "endDate": "{{end_date}}",
    "locations": [at(*BAHRIA, "07:00"), at(*DHA, "07:25"), at(*ROOTS, "08:00")],
    "passengers": [{"passengerId": "{{passenger1_id}}", "locationSequence": 1},
                   {"passengerId": "{{passenger2_id}}", "locationSequence": 2}],
}

shifts_owner = {"name": "5 · Shifts · Owner", "item": [
    req("Create Shift", "POST", "/api/v1/shift", "owner_session", body=shift_body,
        tests=save("shift_id", "r.data.shiftId"),
        desc=("One recurring shift, named after its vehicle's number: the days it runs, from startDate until endDate (empty runs until changed), "
              "and a route of 2 to 20 places reached one after another. The trip lasts from the first place's "
              "time to the last. Each passenger waits at the place with their locationSequence.\n\n"
              "Only the owner creates shifts. The driver is the owner or an approved driver, the vehicle an "
              "approved vehicle with seats, the passengers approved passengers who fit in it. None of them may be "
              "on another shift whose trips overlap on a date both run, in any service, and the driver and "
              "vehicle may not be on a carpool ride during a trip. Touching times do not overlap."),
        examples=["Create shift", "Shift in the past (refused)", "More passengers than seats (refused)",
                  "Passenger outside the service (refused)", "Member driver builds a shift (refused)",
                  "Vehicles only member put behind the wheel (refused)"]),

    req("Clash · Same Driver", "POST", "/api/v1/shift", "owner_session",
        body=dict(shift_body, vehicleId="{{vehicle3_id}}", passengers=[],
                  locations=[at(*DHA, "07:30"), at(*ROOTS, "08:30")]),
        desc="MEANT TO FAIL: the owner is already driving the morning run at that time.",
        examples=["Driver on an overlapping shift (refused)"]),

    req("Clash · Same Vehicle", "POST", "/api/v1/shift", "owner_session",
        body=dict(shift_body, driverId="{{driver2_id}}", passengers=[],
                  locations=[at(*DHA, "07:30"), at(*ROOTS, "08:30")]),
        desc="MEANT TO FAIL: the vehicle is on the morning run at that time.",
        examples=["Vehicle on an overlapping shift (refused)"]),

    req("Clash · Same Passenger", "POST", "/api/v1/shift", "owner_session",
        body=dict(shift_body, driverId="{{driver2_id}}", vehicleId="{{vehicle2_id}}",
                  locations=[at(*DHA, "07:30"), at(*ROOTS, "08:30")],
                  passengers=[{"passengerId": "{{passenger1_id}}", "locationSequence": 1}]),
        desc="MEANT TO FAIL: another driver and vehicle, but the passenger is on the morning run.",
        examples=["Passenger on an overlapping shift (refused)"]),

    req("Clash · A Carpool Ride", "POST", "/api/v1/shift", "owner_session",
        body=dict(shift_body, vehicleId="{{vehicle3_id}}", passengers=[],
                  daysOfWeek=[datetime.fromisoformat(RIDE_DATE).isoweekday()],
                  startDate="{{ride_date}}", endDate="{{ride_date}}",
                  locations=[at(*BAHRIA, "12:10"), at(*ROOTS, "12:40")]),
        desc="MEANT TO FAIL: the owner drives a carpool ride at that hour (create it in folder 8 first).",
        examples=["Shift on top of a ride (refused)"]),

    req("Create Touching Shift", "POST", "/api/v1/shift", "owner_session",
        body=dict(shift_body, driverId="{{driver2_id}}", vehicleId="{{vehicle2_id}}",
                  locations=[at(*ROOTS, "08:00"), at(*MARKAZ, "08:30")],
                  passengers=[{"passengerId": "{{passenger1_id}}", "locationSequence": 1}]),
        tests=save("shift_c_id", "r.data.shiftId"),
        desc="Starts the minute the morning run ends. Touching trips do not overlap, so the same passenger fits.",
        examples=["Create touching shift (member driver)"]),

    req("Shifts On A Vehicles Only Member's Vehicles", "POST", "/api/v1/shift", "owner_session",
        body=dict(shift_body, driverId="{{driver2_id}}", vehicleId="{{vehicle6_id}}", passengers=[],
                  locations=[at(*DHA, "10:00"), at(*ROOTS, "10:30")]),
        desc=("A member who joined with vehicles only never drives, so their vehicles go on shifts with the "
              "owner or another driver behind the wheel. Two of their vehicles can be on two shifts at the "
              "same time, each with its own driver. The vehicle's owner is told, sees those shifts under "
              "/shift/mine and can read them."),
        examples=["Shift on the member's van, owner driving", "Shift on the member's car at the same time, second driver"]),

    req("Shift Detail", "GET", "/api/v1/shift/detail", "owner_session",
        query=[q("shift_id", "{{shift_id}}")],
        desc="For the owner and the driver of the shift: route, passengers with their stops and numbers, seats.",
        examples=["Shift detail (owner)", "Shift detail after the vehicle grew", "Shift detail after the vehicle was renamed",
                  "Shift after a passenger left", "Read a deleted shift (refused)"]),

    req("Service Shifts", "GET", "/api/v1/shift", "owner_session",
        query=[q("page", "1"), q("status", "", True), q("day_of_week", "", True), q("search", "", True),
               q("start_time", "06:30", True), q("end_time", "08:05", True)],
        desc=("status is active (default), completed or all. day_of_week keeps the shifts running that day. "
              "search matches the vehicle number (which is the shift's name) or any place on the route. "
              "start_time and end_time work like the ride search: trips starting at or after start_time and "
              "over by end_time, either can be sent alone."),
        examples=["Service shifts", "Service shifts for a day", "Service shifts in a time window",
                  "Service shifts by vehicle number"]),

    req("Edit Shift", "PATCH", "/api/v1/shift", "owner_session",
        body={"shiftId": "{{shift_id}}",
              "locations": [at(*BAHRIA, "06:55"), at(*DHA, "07:20"), at(*ROOTS, "07:55")]},
        desc=("Send only what changes: driverId, vehicleId (the shift is renamed after the new vehicle), "
              "daysOfWeek, startDate, endDate (empty makes it open ended) or locations (replaces the route). "
              "Every edit runs the same checks as creating, trips not yet started are rewritten, and the "
              "driver, the passengers and the owner of the vehicle are told."),
        examples=["Edit the route", "Move a shift onto another (refused)", "Member driver edits a shift (refused)"]),

    req("Change Passengers (bulk)", "PUT", "/api/v1/shift/passengers", "owner_session",
        body={"shiftId": "{{shift_id}}", "add": [{"passengerId": "{{passenger2_id}}", "locationSequence": 2}],
              "move": [{"passengerId": "{{passenger1_id}}", "locationSequence": 3}], "remove": []},
        desc=("Adds, moves and removes in one transaction, judged against the seats once every change lands. "
              "A removed passenger's row is kept as history, adding them again starts a new one."),
        examples=["Add and move passengers", "Remove a passenger", "Add a passenger already on the shift (refused)",
                  "Add a passenger outside the service (refused)"]),

    req("Attendance Of A Trip", "GET", "/api/v1/shift/attendance", "owner_session",
        query=[q("shift_id", "{{shift_id}}"), q("date", "{{start_date}}")],
        desc="Everyone is present unless they marked themselves absent.",
        examples=["Attendance of a trip (owner)", "Attendance of a trip you do not run (refused)"]),

    req("Trips", "GET", "/api/v1/shift/occurrences", "owner_session",
        query=[q("page", "1"), q("shift_id", "{{shift_id}}", True)],
        desc="Trips already written, with present and absent counts. A driver who is not an owner sees the trips they drove.",
        examples=["Trips of a shift", "Trips of the service"]),

    req("Shift History", "GET", "/api/v1/shift/history", "owner_session", query=[q("page", "1")],
        desc="Every shift the service ran, deleted ones included with deletedAt.", examples=["Shift history"]),

    req("Passenger History", "GET", "/api/v1/shift/passengers/history", "owner_session",
        query=[q("page", "1"), q("passenger_id", "{{passenger2_id}}", True)],
        desc="Each stint of a passenger on a shift: when they joined the service, when they were added and removed, how often they travelled.",
        examples=["Passenger history (owner)"]),

    req("Search Shift Requests", "GET", "/api/v1/shift/requests", "owner_session",
        query=[q("page", "1"), q("search", "", True), q("day_of_week", "", True),
               q("start_time", "", True), q("end_time", "", True)],
        desc=("Passengers' requirements, open ones and the ones addressed to this service. search matches any "
              "requested place, start_time and end_time work like the ride search."),
        examples=["Search shift requests (owner)", "Search shift requests by place", "Search shift requests in a time window"]),

    req("Delete Shift", "DELETE", "/api/v1/shift", "owner_session",
        query=[q("shift_id", "{{shift_c_id}}")],
        desc="Archived to del_shifts, deleted, and the driver and passengers are told. Its trips stay as travel history.",
        examples=["Delete the touching shift", "Delete shift", "Delete today's shift",
                  "Delete the shift on the member's van", "Delete the shift on the member's car"]),
]}

# ------------------------------------------------------- 6 shifts driver ----
shifts_driver = {"name": "6 · Shifts · Driver", "item": [
    req("My Shifts", "GET", "/api/v1/shift/mine", "driver2_session",
        query=[q("page", "1"), q("status", "", True), q("day_of_week", "", True), q("search", "", True),
               q("start_time", "", True), q("end_time", "", True)],
        desc="The shifts the caller drives and the shifts their vehicles are on, with the same filters as the service's shifts.",
        examples=["My shifts (member driver)", "My shifts (owner driving)", "My shifts (vehicles only member)"]),

    req("Shift I Drive", "GET", "/api/v1/shift/detail", "driver2_session",
        query=[q("shift_id", "{{shift_c_id}}")],
        desc="Readable by the owner, the shift's driver and the owner of its vehicle.",
        examples=["Driver reads the shift they drive", "Vehicle owner reads a shift their vehicle is on",
                  "Driver reads a shift they do not drive (refused)"]),

    req("Send Location Update", "POST", "/api/v1/shift/location", "driver2_session",
        body={"shiftId": "{{shift_today_id}}", "message": "Running five minutes late",
              "location": "Gulberg Greens main gate", "lat": 33.6180, "lng": 73.1560},
        desc="Only the driver of the shift, only on a day it runs, and only its passengers travelling that day are told.",
        examples=["Driver sends a location update", "Update from a driver who is not driving (refused)",
                  "Update on a shift with no trip today (refused)"]),
]}

# ---------------------------------------------------- 7 shifts passenger ----
shifts_passenger = {"name": "7 · Shifts · Passenger", "item": [
    req("My Shifts", "GET", "/api/v1/passenger/shifts", "passenger_session",
        query=[q("page", "1"), q("status", "", True), q("day_of_week", "", True), q("search", "", True),
               q("start_time", "", True), q("end_time", "", True)],
        examples=["My shifts (passenger)"]),

    req("Shift Detail", "GET", "/api/v1/passenger/shift/detail", "passenger_session",
        query=[q("shift_id", "{{shift_id}}")],
        desc="myStop is the caller's own stop, other passengers are shown without their numbers.",
        examples=["Passenger reads shift detail", "Passenger reads a shift they are not on (refused)"]),

    req("Mark Absent", "PATCH", "/api/v1/passenger/shift/attendance", "passenger_session",
        body={"shiftId": "{{shift_id}}", "date": "{{start_date}}", "status": "absent"},
        desc=("status absent or present, for a date the shift runs that has not started. The driver of that trip "
              "and the owner are told, the other passengers can see it."),
        examples=["Mark absent", "Mark absent again", "Mark absent on a day the shift does not run (refused)",
                  "Mark absent on a shift you are not on (refused)", "Mark absent after the trip started (refused)"]),

    req("Trip Attendance", "GET", "/api/v1/passenger/shift/attendance", "passenger2_session",
        query=[q("shift_id", "{{shift_id}}"), q("date", "{{start_date}}")],
        examples=["Attendance of a trip (other passenger)"]),

    req("Travel History", "GET", "/api/v1/passenger/shift/history", "passenger_session",
        query=[q("page", "1")],
        desc="Every trip that started, with the driver, vehicle and route of that day. It outlives leaving and deleted shifts.",
        examples=["Travel history", "Travel history after the shift was deleted"]),

    req("Create Shift Request", "POST", "/api/v1/passenger/shift/request", "passenger3_session",
        body={"serviceId": "", "daysOfWeek": [1, 2, 3, 4, 5], "contactNumber": "+923001234567",
              "note": "A seat for my daughter", "locations": [at(*G11, "07:30"), at(*BLUE, "08:15")]},
        tests=save("shift_request_id", "r.data.requestId"),
        desc="serviceId is optional, a request addressed to one service is shown only to its owner.",
        examples=["Create shift request", "Create shift request for one service", "Shift request with one place (refused)"]),

    req("My Shift Requests", "GET", "/api/v1/passenger/shift/requests", "passenger3_session",
        query=[q("page", "1"), q("search", "", True), q("day_of_week", "", True),
               q("start_time", "07:00", True), q("end_time", "09:00", True)],
        examples=["My shift requests", "My shift requests in a time window"]),

    req("Delete Shift Request", "DELETE", "/api/v1/passenger/shift/request", "passenger3_session",
        query=[q("request_id", "{{shift_request_id}}")], examples=["Delete shift request"]),

    req("Notifications", "GET", "/api/v1/passenger/notifications", "passenger_session",
        examples=["Passenger notifications"]),
]}

# ---------------------------------------------------------- 8 ride share ----
ride_body = {
    "startDatetime": "{{ride_date}} 12:00:00", "estimatedEndDatetime": "{{ride_date}} 13:00:00",
    "numberOfSeats": 3, "startLocation": "Location A", "endLocation": "Location B",
    "routePoints": ["LocationA1", "LocationA2"], "fare": 20.5, "routeDetails": "Via Highway 1",
    "vehicleId": "{{vehicle_id}}", "makeTemplate": True, "isRecurring": False,
    "frequency": 1, "period": 1, "daysOfWeek": [1],
}

rideshare = {"name": "8 · Ride Share", "item": [
    req("Create Ride", "POST", "/api/v1/ride/create", "owner_session", body=ride_body,
        tests=save("ride_id", "r.data.id"),
        desc=("A ride goes on the driver's own vehicle and never offers more seats than it has. A vehicle whose "
              "seats were never recorded keeps working as before until they are. The driver cannot be on "
              "another ride at the same time on any vehicle, the vehicle cannot carry another ride, and "
              "neither can be on a shift trip. A recurring series is checked date by date, against "
              "everything already booked and against its own rides, before anything is written, and one "
              "clashing date refuses the whole series and names that date. Touching times do not overlap."),
        examples=["Create carpool ride", "Ride with more seats than the vehicle (refused)",
                  "Ride on another driver's vehicle (refused)", "Ride on top of a shift trip (refused)",
                  "Overlapping ride on the same vehicle (refused)",
                  "Same driver on another vehicle at the same time (refused)", "Ride right after another ride",
                  "Recurring series landing on a ride (refused)", "The refused series left nothing behind",
                  "Recurring series landing on a shift trip (refused)",
                  "Recurring series that runs into itself (refused)", "Create a recurring ride series",
                  "Vehicle owner's ride on top of their vehicle's shift (refused)"]),

    req("Update Ride", "PATCH", "/api/v1/driver/ride/update", "owner_session",
        query=[q("ride_id", "{{ride_id}}")], body={"numberOfSeats": 2},
        desc=("Only the driver of the ride. Seats stay between the seats already booked and the vehicle's seats. "
              "status inactive cancels the ride through the cancellation guards and archives it. status active "
              "only matters for a ride the worker closed: one that has ended stays closed, and one still ahead "
              "must not clash with anything the driver or vehicle took on meanwhile. An empty status changes nothing."),
        examples=["Update ride seats", "Update ride above the vehicle (refused)", "Update somebody else's ride (refused)",
                  "Empty status leaves the ride alone", "Cancel a ride"]),

    req("Driver Rides", "GET", "/api/v1/driver/rides", "owner_session",
        query=[q("page", "1"), q("status", "all")], examples=["Driver rides"]),
    req("Get One Ride", "GET", "/api/v1/ride", "open_token", query=[q("ride_id", "{{ride_id}}")],
        tests=save("ride_code", "r.data.ride.code"),
        examples=["Get one ride", "Get a ride series", "The ride is still active"]),
    req("Filtered Rides", "GET", "/api/v1/ride/filtered", "open_token",
        query=[q("page", "1"), q("search", "LocationA1"),
               q("start_time", "{{ride_date}} 00:00:00"), q("estimated_end_time", "{{ride_date}} 23:59:59")],
        desc="Without start_time and estimated_end_time the search covers the next seven days only.",
        examples=["Filtered rides"]),
    req("Ride Templates", "GET", "/api/v1/ride/templates", "owner_session",
        tests=save_from_list("ride_template_id", "r.data", "rideId", "ride_id", "id"),
        examples=["Ride templates"]),
    req("Delete Ride Template", "DELETE", "/api/v1/ride/template", "owner_session",
        query=[q("ride_template_id", "{{ride_template_id}}")], examples=["Delete ride template"]),
    req("Passenger Ride Request", "POST", "/api/v1/passenger/ride/request", "open_token",
        body={"startDatetime": "{{ride_date}} 15:30:00", "estimatedEndDatetime": "{{ride_date}} 17:00:00",
              "numberOfSeats": 2, "startLocation": "Location A", "endLocation": "Location B",
              "routeDetails": "via gt road", "contactNumber": "+923301221121"},
        tests=save("ride_request_id", "r.data.requestId"), examples=["Passenger ride request"]),
    req("Get A Posted Ride Request", "GET", "/api/v1/ride/request", "open_token",
        query=[q("request_id", "{{ride_request_id}}")],
        desc="request_id accepts the id or the short code from the openUrl above.",
        examples=["Get a posted ride request"]),
    req("Get Ride Requests", "GET", "/api/v1/driver/ride/requests", "owner_session",
        query=[q("page", "1")], examples=["Get ride requests"]),
    req("Get Announcements", "GET", "/api/v1/announcements", "open_token", examples=["Get announcements"]),
    req("Driver Notifications", "GET", "/api/v1/user/notifications", "owner_session",
        examples=["Driver notifications"]),

    req("Driver Books A Seat For A Phone Caller", "POST", "/api/v1/driver/seat/book", "owner_session",
        body={"rideId": "{{ride_id}}", "name": "Phone Caller", "mobileNumber": "+923001112233",
              "code": "{{ride_code}}", "seats": 1, "isBook": True},
        desc=("The older way to fill a ride board post: someone calls the driver directly, and the driver enters "
              "the booking on their behalf. code is the ride's own code, read off Get One Ride. Send isBook "
              "false with the same body to release the seat again."),
        tests=save("driver_booking_id", "r.data.bookingId"),
        examples=["Driver books a seat for a phone caller"]),
    req("Driver's Bookings On A Ride", "GET", "/api/v1/driver/bookings", "owner_session",
        query=[q("ride_id", "{{ride_id}}")], examples=["Driver's bookings on the ride"]),
    req("Reserve (Confirm) A Booking", "GET", "/api/v1/driver/booking/reserve", "owner_session",
        query=[q("booking_id", "{{driver_booking_id}}")], examples=["Reserve (confirm) a booking"]),
    req("Passenger Books A Seat Directly", "POST", "/api/v1/passenger/seat/book", "open_token",
        body={"rideId": "{{ride_id}}", "name": "Direct Rider", "mobileNumber": "+923001112244",
              "seats": 1, "code": "{{ride_code}}"},
        desc="The open-board equivalent: a rider with the ride's code books straight from the app, no phone call needed.",
        examples=["Passenger books a seat directly"]),
    req("Rate The Driver", "POST", "/api/v1/driver/rate", "open_token",
        body={"driverId": "{{owner_id}}", "rideId": "{{ride_id}}", "mobileNumber": "+923001112244", "rating": 5},
        desc="Left after a ride. Public because the rider who is rating may not hold a driver session.",
        examples=["Rate the driver"]),

    req("Cancel Ride Series", "DELETE", "/api/v1/ride/series", "owner_session",
        query=[q("ride_id", "{{ride_id}}")],
        desc=("Cancels every ride of the series the given ride belongs to, the parent and all its repeats. Rides "
              "with bookings or starting within the cancellation window are skipped with the reason. Kept last "
              "in this folder because it cancels the ride the requests above use."),
        examples=["Cancel the ride series"]),
]}

# ------------------------------------------------ 9 account lifecycle ----
# a throwaway driver and passenger, kept separate from the accounts every other
# folder depends on, so deleting them here never breaks a request that runs later
lifecycle = {"name": "9 · Account Lifecycle & Support", "item": [
    req("Lifecycle Driver · Register", "POST", "/api/v1/driver/register", "open_token",
        body={"deviceId": "lc", "mobile": "+923400000008", "name": "Lifecycle Driver",
              "password": "Golang@12122", "gender": "male"},
        desc="A disposable account used only to demonstrate the self-service calls below.",
        examples=["Lifecycle driver · Register"]),
    req("Resend OTP", "POST", "/api/v1/otp/resend", "open_token",
        query=[q("mobile_number", "+923400000008"), q("otp_operation", "ACTIVATE_DRIVER")],
        desc="Issues a fresh code for the same operation; the one from registration no longer verifies.",
        tests=save("lifecycle_otp", "r.data.tempOTP"), examples=["Resend OTP"]),
    req("Lifecycle Driver · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "+923400000008", "otp": "{{lifecycle_otp}}", "operation": "ACTIVATE_DRIVER"},
        examples=["Lifecycle driver · Verify OTP"]),
    req("Lifecycle Driver · Login", "POST", "/api/v1/driver/login", "open_token",
        body={"deviceId": "lc", "mobile": "+923400000008", "password": "Golang@12122"},
        tests=save("lifecycle_driver_session", "r.data.sessionId"), examples=["Lifecycle driver · Login"]),
    req("Lifecycle Driver · Set Pin", "POST", "/api/v1/driver/pin", "lifecycle_driver_session",
        query=[q("pin", "151515")], examples=["Lifecycle driver · Set pin"]),

    req("Driver · Forgot Password", "GET", "/api/v1/driver/password/forgot", "open_token",
        query=[q("mobile_number", "+923400000008")],
        desc=("Always 200 whether or not the number is registered, so a caller can't probe for accounts by it. "
              "Sends a confirmation OTP — nothing changes yet."),
        tests=save("lifecycle_reset_otp", "r.data.tempOTP"), examples=["Driver · Forgot password"]),
    req("Driver · Confirm The New Password", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "+923400000008", "otp": "{{lifecycle_reset_otp}}",
              "operation": "FORGOT_PASSWORD", "password": "Golang@12123"},
        desc="The new password travels with this call — nothing was cached by Forgot Password above.",
        examples=["Driver · Confirm the new password"]),

    req("Driver · Change Password, Wrong Current One", "POST", "/api/v1/driver/password/reset",
        "lifecycle_driver_session",
        body={"oldPassword": "not the real password", "newPassword": "Golang@12124"},
        desc="Authenticated, needs the current password. This one is wrong, so no OTP is even sent.",
        examples=["Driver · Change password, wrong current one (refused)"]),
    req("Driver · Change Password", "POST", "/api/v1/driver/password/reset", "lifecycle_driver_session",
        body={"oldPassword": "Golang@12123", "newPassword": "Golang@12124"},
        desc="Correct current password: sends a confirmation OTP. The password is still the old one until it's verified.",
        tests=save("lifecycle_reset_otp", "r.data.tempOTP"), examples=["Driver · Change password"]),
    req("Driver · Confirm The Password Change", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "+923400000008", "otp": "{{lifecycle_reset_otp}}", "operation": "UPDATE_PASSWORD"},
        desc="No password in the body this time — the new one was already cached by Change Password above.",
        examples=["Driver · Confirm the password change"]),

    req("Driver · Forgot Pin", "GET", "/api/v1/driver/pin/forgot", "open_token",
        query=[q("mobile_number", "+923400000008")],
        tests=save("lifecycle_reset_otp", "r.data.tempOTP"), examples=["Driver · Forgot pin"]),
    req("Driver · Confirm The New Pin", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "+923400000008", "otp": "{{lifecycle_reset_otp}}", "operation": "FORGOT_PIN", "pin": "161616"},
        examples=["Driver · Confirm the new pin"]),
    req("Driver · Change Pin", "POST", "/api/v1/driver/pin/reset", "lifecycle_driver_session",
        body={"oldPin": "161616", "newPin": "171717"},
        tests=save("lifecycle_reset_otp", "r.data.tempOTP"), examples=["Driver · Change pin"]),
    req("Driver · Confirm The Pin Change", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "+923400000008", "otp": "{{lifecycle_reset_otp}}", "operation": "UPDATE_PIN"},
        examples=["Driver · Confirm the pin change"]),

    req("Driver · Deactivate/Reactivate Own Profile", "PATCH", "/api/v1/driver/status", "lifecycle_driver_session",
        query=[q("status", "inactive"), q("pin", "171717")],
        desc="Self-service, distinct from an admin suspension — send status active to switch back on.",
        examples=["Driver · Deactivate own profile", "Driver · Reactivate own profile"]),

    req("Lifecycle Passenger · Register", "POST", "/api/v1/passenger/register", "open_token",
        body={"deviceId": "lcp", "mobile": "+923400000009", "name": "Lifecycle Passenger",
              "gender": "female", "password": "Golang@12122"},
        tests=save("lifecycle_passenger_otp", "r.data.tempOTP"), examples=["Lifecycle passenger · Register"]),
    req("Lifecycle Passenger · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "+923400000009", "otp": "{{lifecycle_passenger_otp}}", "operation": "ACTIVATE_PASSENGER"},
        examples=["Lifecycle passenger · Verify OTP"]),
    req("Lifecycle Passenger · Login", "POST", "/api/v1/passenger/login", "open_token",
        body={"deviceId": "lcp", "mobile": "+923400000009", "password": "Golang@12122"},
        tests=save("lifecycle_passenger_session", "r.data.sessionId"), examples=["Lifecycle passenger · Login"]),
    req("Passenger · Change Password", "POST", "/api/v1/passenger/password/reset", "lifecycle_passenger_session",
        body={"oldPassword": "Golang@12122", "newPassword": "Golang@12123"},
        desc="Two-step like the driver's: a confirmation OTP first, nothing changes until it's verified.",
        tests=save("lifecycle_passenger_reset_otp", "r.data.tempOTP"), examples=["Passenger · Change password"]),
    req("Passenger · Confirm The Password Change", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "+923400000009", "otp": "{{lifecycle_passenger_reset_otp}}",
              "operation": "PASSENGER_UPDATE_PASSWORD"},
        examples=["Passenger · Confirm the password change"]),
    req("Passenger · Delete Own Profile", "DELETE", "/api/v1/passenger/delete", "lifecycle_passenger_session",
        examples=["Passenger · Delete own profile"]),
    req("Driver · Delete Own Profile", "DELETE", "/api/v1/driver/delete", "lifecycle_driver_session",
        query=[q("pin", "171717")], examples=["Driver · Delete own profile"]),

    req("Get In Touch", "POST", "/api/v1/approach", "open_token",
        body={"name": "Interested Fleet Owner", "number": "+923001234599", "email": "owner@example.com",
              "message": "We run 12 vans in Lahore, interested in Pick & Drop", "type": "contact"},
        desc="Public contact form, unrelated to any account — type is contact or complain.",
        examples=["Get in touch"]),
]}

collection = {
    "info": {
        "_postman_id": "b7c41f02-5e6a-4d38-9c11-3a7f0e2b4d91",
        "name": "SathSawari · Pick & Drop and Shifts",
        "description": (
            "Pick & Drop services, their recurring shifts, and the ride share rules they share, in the order "
            "you would actually exercise them.\n\n"
            "Pick & Drop and shifts are two different things: a driver enables a Pick & Drop service and "
            "becomes its only owner, and the owner builds shifts, the recurring trips of that service.\n\n"
            "SETUP\n"
            "1. Use the same environment as your existing Rideshare collection, it needs base-url and open_token.\n"
            "2. Run the folders top to bottom. Ids are captured into environment variables by test scripts.\n"
            "3. OTPs come back as tempOTP while you are not on production, so verification chains too.\n"
            "4. The dates come from the collection variables start_date, end_date and ride_date, change them "
            "if they have passed. Dates are YYYY-MM-DD, times HH:MM, days of week 1 (Monday) to 7 (Sunday), "
            "all in Pakistan time.\n\n"
            "TOKENS USED\n"
            "  open_token          public endpoints\n"
            "  admin_session       admin console\n"
            "  owner_session       the driver who owns the Pick & Drop service\n"
            "  driver2_session     a driver who joins the service\n"
            "  driver3_session     a driver who joins with vehicles only\n"
            "  passenger_session   Ayesha, a passenger\n"
            "  passenger2_session  Bilal, a passenger\n"
            "  passenger3_session  Sana, a passenger who is turned away\n\n"
            "Every request carries the real responses recorded by tools/flow_test.py, the success and every "
            "refusal that proves one of its rules. Requests whose description starts MEANT TO FAIL are there "
            "to show a rule refusing."
        ),
        "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
    },
    "item": [admin, accounts, owner, member, passenger, shifts_owner, shifts_driver, shifts_passenger, rideshare,
             lifecycle],
    "variable": [
        {"key": "base-url", "value": "http://localhost:8080", "type": "string"},
        {"key": "start_date", "value": START_DATE, "type": "string"},
        {"key": "end_date", "value": END_DATE, "type": "string"},
        {"key": "ride_date", "value": RIDE_DATE, "type": "string"},
    ],
}

out = os.path.join(ROOT, "SathSawari-ShiftManagement.postman_collection.json")
with open(out, "w") as f:
    json.dump(collection, f, indent=2)

n = sum(len(folder["item"]) for folder in collection["item"])
print(f"wrote {out}")
print(f"folders: {len(collection['item'])}, requests: {n}")
