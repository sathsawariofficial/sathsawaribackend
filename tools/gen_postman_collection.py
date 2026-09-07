import json

BASE = "{{base-url}}"
import os
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
# the responses captured by tools/flow_test.py against a live server
RECORDED_PATH = os.path.join(ROOT, ".sath", "responses.json")


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


# Real responses captured by driving the whole product against a live database
# (tools/flow.py). Keyed by the label used there, so an example is never invented.
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
    name = "SUCCESS" if rec["status"] == 200 else "ERROR"
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



# Collection request name -> the label(s) it was recorded under in tools/flow.py.
# A request can carry several examples, typically the success plus the refusal that
# proves a rule.
EXAMPLE_MAP = {
    "Admin Login": ["Admin login"],
    "Platform Overview": ["Admin overview (populated)"],
    "List Passengers": ["Admin list passengers"],
    "Passenger Profile": ["Admin passenger profile"],
    "Suspend A Passenger": ["Admin suspend passenger", "Suspended passenger cannot log in"],
    "Suspend A Driver": ["Admin suspend driver"],
    "Delete A Passenger": ["Admin deletes a passenger"],
    "List Groups": ["Admin list groups"],
    "Group Details": ["Admin group details"],
    "Shut A Group Down": ["Admin shut a fleet down"],
    "List Shifts": ["Admin list shifts"],
    "Shift Details": ["Admin shift details"],
    "Get Roles": ["List roles"],
    "Get Permissions": ["List permissions"],
    "Create Role": ["Create role"],
    "Set Role Permissions (bulk)": ["Set role permissions"],
    "Update Role": ["Update role", "Rename a system role (refused)"],
    "Delete Role": ["Delete custom role", "Delete a system role (refused)"],

    "Owner Driver · Register": ["Register owner driver"],
    "Owner Driver · Verify OTP": ["Verify driver otp"],
    "Owner Driver · Login": ["Login owner driver"],
    "Owner Driver · Set Pin": ["Set pin"],
    "Owner Driver · Register Vehicle (7 seats, AC)": ["Register vehicle (7 seats)"],
    "Second Driver · Register": ["Register second driver"],
    "Second Driver · Verify OTP": ["Verify driver 2 otp"],
    "Second Driver · Login": ["Login second driver"],
    "Second Driver · Set Pin": ["Second driver pin"],
    "Second Driver · Register Vehicle (4 seats)": ["Register second vehicle"],
    "Update Vehicle (seats / AC / heating)": ["Update vehicle comfort"],
    "Get Vehicles": ["Get vehicles"],
    "Passenger · Register (female)": ["Register passenger (female)"],
    "Passenger · Verify OTP": ["Verify passenger otp"],
    "Passenger · Login": ["Login passenger Ayesha"],
    "Passenger 2 · Register (male)": ["Register passenger (male)"],
    "Passenger 2 · Verify OTP": ["Verify passenger otp"],
    "Passenger 2 · Login": ["Login passenger Bilal"],
    "Passenger · Profile": ["Passenger profile"],
    "Passenger · Forgot Password": ["Passenger forgot password"],
    "Passenger · Logout": ["Passenger logout"],
    "Passenger · Reset Password (change)": ["Passenger reset password"],

    "Set Weekly Form (bulk)": ["Set weekly travel form", "Bad day of week (refused)"],
    "Get Weekly Form": ["Get weekly travel form"],
    "Passenger 2 · Set Weekly Form": ["Passenger 2 travel form"],

    "Create Group": ["Create group", "Create group with no vehicle (refused)"],
    "My Groups": ["My groups"],
    "Find Groups To Join (driver)": ["Search groups (driver)"],
    "Driver · Request To Join (driver + vehicle)": ["Driver join request", "Duplicate join request (refused)"],
    "Passenger · Request To Join": ["Passenger join request"],
    "Passenger 2 · Request To Join": ["Passenger 2 join request"],
    "List Pending Requests": ["List pending requests"],
    "Decide Requests (bulk)": ["Decide requests in bulk", "Re-approve the same rows (all skipped)"],
    "Promote Sub Manager (bulk)": ["Promote sub manager", "Promote an outsider (skipped)"],
    "Group Details (rosters)": ["Group details", "Passenger token on a driver route (refused)"],
    "Passenger Travel Forms (manager view)": ["Passenger travel forms (manager view)"],
    "Find Groups To Join (passenger)": ["Search groups (passenger)"],
    "My Groups (passenger)": ["Passenger my groups"],
    "Driver Leaves The Group": ["Driver leaves the fleet"],
    "Passenger Leaves The Group": ["Passenger leaves the fleet"],
    "Delete Group": ["Delete the group", "Delete the group with shifts (refused)"],

    "Create Morning PICKUP Shift (+ template)": ["Create pickup shift (+template)", "Wrong gender on a seat (refused)"],
    "Create Evening DROP Shift": ["Create drop shift"],
    "Shift Detail": ["Shift detail"],
    "Group Shift Roster": ["Group shift roster"],
    "My Shifts (driver)": ["My shifts (driver)"],
    "My Shifts (passenger)": ["My shifts (passenger)"],
    "Passenger · Shift Detail": ["Passenger reads shift detail"],
    "Swap Male Rider For Female (one call)": ["Swap male rider for female", "Seat a passenger twice (refused)"],
    "Free A Seat": ["Free a seat"],
    "Sub Manager Builds A Shift": ["Sub manager builds a shift", "Sub manager decides membership (refused)"],
    "Clash Check · Same Vehicle, Overlapping Time": ["Overlapping vehicle (refused)"],
    "Reschedule Shift (move the time)": ["Reschedule the shift"],
    "Cancel Shift": ["Cancel the sub manager shift"],
    "Get Shift Templates": ["Get shift templates"],
    "Delete Shift Template": ["Delete shift template"],
    "Create Ride": ["Create carpool ride", "Carpool ride clashing with a shift (refused)"],
    "Driver Rides": ["Driver rides"],
    "Get One Ride": ["Get one ride"],
    "Filtered Rides": ["Filtered rides"],
    "Ride Templates": ["Ride templates"],
    "Update Ride": ["Update ride seats"],
    "Passenger Ride Request": ["Passenger ride request"],
    "Get Ride Requests": ["Get ride requests"],
    "Get Announcements": ["Get announcements"],
    "Driver Notifications": ["Driver notifications"],
    "Admin · List Rides": ["Admin list rides"],
    "Admin · List Drivers": ["Admin list drivers"],
    "Admin · List Vehicles": ["Admin list vehicles"],
}

def req(name, method, path, token, body=None, query=None, tests=None, desc=None, examples=None):
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
    for label in (examples or EXAMPLE_MAP.get(name) or [name]):
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


items = []

# ---------------------------------------------------------------- 0 admin ----
admin = {"name": "0 · Admin Console", "item": [
    req("Admin Login", "POST", "/api/v1/admin/login", "open_token",
        body={"username": "twssawari", "password": "vR7!xK2@pQ9#Lm4$Zw8^Ty1&Nc5*Hs3%Df6!Ba"},
        tests=save("admin_session", "r.data.sessionId"),
        desc="Saves admin_session. Run this first if you want to touch roles or permissions."),

    req("Platform Overview", "GET", "/api/v1/admin/overview", "admin_session",
        desc=("The whole product counted in one query: drivers, passengers, vehicles, fleets, "
              "shifts and carpool rides, each as a total with the live slice of it, plus how many "
              "join requests are sitting unanswered across every fleet.\n\n"
              "This is the admin's first screen.")),

    req("List Passengers", "GET", "/api/v1/admin/passengers", "admin_session",
        query=[q("page", "1"), q("search", "", True), q("status", "", True)],
        tests=[
            "const r = pm.response.json();",
            "if (r.data && r.data.passengers && r.data.passengers.length) {",
            '    pm.environment.set("admin_passenger_id", r.data.passengers[0].id);',
            "}",
        ],
        desc=("Passengers were invisible to the admin until this feature gave them real accounts. "
              "search matches the name or the mobile number, status filters active, inactive or "
              "pending.")),

    req("Passenger Profile", "GET", "/api/v1/admin/passenger", "admin_session",
        query=[q("passenger_id", "{{passenger1_id}}")],
        desc="One account explained: the fleets they ride with and the standing travel form they filled in, which is what a support call actually needs."),

    req("Suspend A Passenger", "PATCH", "/api/v1/admin/passenger/status", "admin_session",
        query=[q("passenger_id", "{{passenger1_id}}")],
        body={"status": "inactive"},
        desc=("The softer moderation tool. A suspended account cannot log in or be seated on a "
              "shift, but nothing it was part of is destroyed. Send active to bring it back.")),

    req("Suspend A Driver", "PATCH", "/api/v1/admin/driver/status", "admin_session",
        query=[q("driver_id", "{{driver2_id}}")],
        body={"status": "inactive"},
        desc="Same idea for a driver, the alternative to deleting them outright."),

    req("Delete A Passenger", "DELETE", "/api/v1/admin/passenger", "admin_session",
        query=[q("passenger_id", "{{admin_passenger_id}}")],
        desc=("Permanent. The account is archived, its seats on upcoming shifts are freed with the "
              "taken counts corrected, its fleet memberships end and its travel form is cleared, "
              "so nothing anywhere is left pointing at somebody who no longer exists.")),

    req("List Groups", "GET", "/api/v1/admin/groups", "admin_session",
        query=[q("page", "1"), q("search", "", True), q("status", "", True)],
        desc="Every fleet with who runs it and how big it is. The member, vehicle, passenger and shift counts come back as subselects, so the page costs one query."),

    req("Group Details", "GET", "/api/v1/admin/group", "admin_session",
        query=[q("group_id", "{{group_id}}")],
        desc=("The full rosters, and unlike the manager's own view this is NOT filtered by status "
              "— an admin looking into a complaint needs to see who was turned away and who left, "
              "not just who is in.")),

    req("Shut A Group Down", "PATCH", "/api/v1/admin/group/status", "admin_session",
        query=[q("group_id", "{{group_id}}")],
        body={"status": "inactive"},
        desc=("The right hammer for a fleet that is misbehaving. Switched off it cannot be found, "
              "joined or built on, while its history stays intact. Send active to restore it.")),

    req("List Shifts", "GET", "/api/v1/admin/shifts", "admin_session",
        query=[q("page", "1"), q("group_id", "", True), q("driver_id", "", True),
               q("direction", "", True), q("status", "", True),
               q("start_time", "", True), q("estimated_end_time", "", True)],
        tests=[
            "const r = pm.response.json();",
            "if (r.data && r.data.shifts && r.data.shifts.length) {",
            '    pm.environment.set("admin_shift_id", r.data.shifts[0].id);',
            "}",
        ],
        desc="Every shift on the platform, filterable the way somebody chasing a complaint would want it."),

    req("Shift Details", "GET", "/api/v1/admin/shift", "admin_session",
        query=[q("shift_id", "{{shift_id}}")],
        desc="The whole trip: who is driving, who is aboard, and the route in the order it is driven."),

    req("Get Roles", "GET", "/api/v1/admin/roles", "admin_session",
        query=[q("page", "1")],
        tests=[
            "const r = pm.response.json();",
            "if (r.data && r.data.roles) {",
            '    const owner = r.data.roles.find(x => x.name === "group_owner");',
            '    const sub = r.data.roles.find(x => x.name === "group_submanager");',
            '    if (owner) pm.environment.set("owner_role_id", owner.id);',
            '    if (sub) pm.environment.set("submanager_role_id", sub.id);',
            '    console.log("seeded roles:", r.data.roles.map(x => x.name));',
            "}",
        ],
        desc="group_owner and group_submanager are seeded at boot with their permissions already mapped. This also saves their ids."),

    req("Get Permissions", "GET", "/api/v1/admin/permissions", "admin_session",
        query=[q("page", "1")],
        tests=[
            "const r = pm.response.json();",
            "if (r.data && r.data.permissions) {",
            "    const byCode = {};",
            "    r.data.permissions.forEach(p => byCode[p.code] = p.id);",
            '    pm.environment.set("permission_ids_json", JSON.stringify(byCode));',
            '    console.log("permissions:", Object.keys(byCode));',
            "}",
        ],
        desc="The eight seeded permission codes. Saves a code -> id map you can use when remapping a role."),

    req("Create Role", "POST", "/api/v1/admin/role", "admin_session",
        body={"name": "group_dispatcher", "description": "Can build shifts but not cancel them"},
        tests=save("custom_role_id", "r.data.id"),
        desc="Proves point 14: a brand new role can be added with no code change."),

    req("Set Role Permissions (bulk)", "PUT", "/api/v1/admin/role/permissions", "admin_session",
        body={"roleId": "{{custom_role_id}}", "permissionIds": ["{{permission_create_shift_id}}"]},
        desc="Replaces the whole permission set of a role in one call. Read permission_ids_json from Get Permissions and paste the ids you want. Sending an empty list strips the role back to nothing."),

    req("Update Role", "PATCH", "/api/v1/admin/role", "admin_session",
        query=[q("role_id", "{{custom_role_id}}")],
        body={"description": "Builds shifts for a fleet"},
        desc="A system role can have its description changed but never its name, because the group checks look those roles up by name."),

    req("Delete Role", "DELETE", "/api/v1/admin/role", "admin_session",
        query=[q("role_id", "{{custom_role_id}}")],
        desc="Refused for a system role, and refused while any member still holds the role."),
]}

# ------------------------------------------------------- 1 accounts ----------
accounts = {"name": "1 · Accounts & Vehicles", "item": [
    req("Owner Driver · Register", "POST", "/api/v1/driver/register", "open_token",
        body={"deviceId": "post_man", "mobile": "+923405421301", "name": "Owner Driver",
              "password": "Golang@12122", "gender": "male"},
        tests=save("owner_otp", "r.data.tempOTP") + ['pm.environment.set("owner_mobile", "+923405421301");'],
        desc="Step 1 of the flow. The owner of the fleet is an ordinary driver."),

    req("Owner Driver · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "{{owner_mobile}}", "otp": "{{owner_otp}}", "operation": "ACTIVATE_DRIVER"}),

    req("Owner Driver · Login", "POST", "/api/v1/driver/login", "open_token",
        body={"deviceId": "post_man", "mobile": "{{owner_mobile}}", "password": "Golang@12122"},
        tests=save("owner_session", "r.data.sessionId") + [
            'if (r.data && r.data.driver) pm.environment.set("owner_driver_id", r.data.driver.id);'],
        desc="Saves owner_session and owner_driver_id."),

    req("Owner Driver · Set Pin", "POST", "/api/v1/driver/pin", "owner_session",
        query=[q("pin", "121212")],
        desc="The pin is needed to register a vehicle."),

    req("Owner Driver · Register Vehicle (7 seats, AC)", "POST", "/api/v1/vehicle/register", "owner_session",
        body={"vehicleNumber": "FLEET-001", "vehicleInfo": "Hiace van",
              "numberOfSeats": 7, "hasAC": True, "hasHeating": False, "pin": "121212"},
        tests=save("owner_vehicle_id", "r.data.vehicleId"),
        desc=("numberOfSeats, hasAC and hasHeating are what let this vehicle join a fleet. Seats are counted excluding the driver, so this van seats 7 passengers.\n\nNOTE: a vehicle is created already active and registration sends NO otp, so tempOTP comes back empty. There is no ACTIVATE_VEHICLE step to run.")),

    req("Second Driver · Register", "POST", "/api/v1/driver/register", "open_token",
        body={"deviceId": "post_man_2", "mobile": "+923405421302", "name": "Second Driver",
              "password": "Golang@12122", "gender": "male"},
        tests=save("driver2_otp", "r.data.tempOTP") + ['pm.environment.set("driver2_mobile", "+923405421302");'],
        desc="This driver will ask to join the fleet and later be made a sub manager."),

    req("Second Driver · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "{{driver2_mobile}}", "otp": "{{driver2_otp}}", "operation": "ACTIVATE_DRIVER"}),

    req("Second Driver · Login", "POST", "/api/v1/driver/login", "open_token",
        body={"deviceId": "post_man_2", "mobile": "{{driver2_mobile}}", "password": "Golang@12122"},
        tests=save("driver2_session", "r.data.sessionId") + [
            'if (r.data && r.data.driver) pm.environment.set("driver2_id", r.data.driver.id);']),

    req("Second Driver · Set Pin", "POST", "/api/v1/driver/pin", "driver2_session",
        query=[q("pin", "131313")]),

    req("Second Driver · Register Vehicle (4 seats)", "POST", "/api/v1/vehicle/register", "driver2_session",
        body={"vehicleNumber": "FLEET-002", "vehicleInfo": "Corolla",
              "numberOfSeats": 4, "hasAC": True, "hasHeating": True, "pin": "131313"},
        tests=save("driver2_vehicle_id", "r.data.vehicleId")),

    req("Update Vehicle (seats / AC / heating)", "PATCH", "/api/v1/vehicle/update", "owner_session",
        body={"vehicleId": "{{owner_vehicle_id}}", "numberOfSeats": 7, "hasAC": True,
              "hasHeating": True, "pin": "121212"},
        desc="hasAC and hasHeating are pointers on update, so leaving one out is different from switching it off."),

    req("Get Vehicles", "GET", "/api/v1/vehicle", "owner_session",
        desc="Now returns numberOfSeats, hasAC and hasHeating."),

    req("Passenger · Register (female)", "POST", "/api/v1/passenger/register", "open_token",
        body={"deviceId": "post_man_p1", "mobile": "+923405421401", "name": "Ayesha",
              "gender": "female", "password": "Golang@12122"},
        tests=save("passenger1_otp", "r.data.tempOTP") + [
            'pm.environment.set("passenger1_mobile", "+923405421401");',
            'if (r.data) pm.environment.set("passenger1_id", r.data.passengerId);'],
        desc="Passengers had no account at all before this feature. Gender matters because seats are gender locked."),

    req("Passenger · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "{{passenger1_mobile}}", "otp": "{{passenger1_otp}}", "operation": "ACTIVATE_PASSENGER"},
        desc="ACTIVATE_PASSENGER is its own operation, separate from ACTIVATE_DRIVER, so one number can hold both kinds of account."),

    req("Passenger · Login", "POST", "/api/v1/passenger/login", "open_token",
        body={"deviceId": "post_man_p1", "mobile": "{{passenger1_mobile}}", "password": "Golang@12122"},
        tests=save("passenger_session", "r.data.sessionId") + [
            'if (r.data && r.data.passenger) pm.environment.set("passenger1_id", r.data.passenger.id);'],
        desc="Issues a passenger_token, a different token type from a driver's."),

    req("Passenger 2 · Register (male)", "POST", "/api/v1/passenger/register", "open_token",
        body={"deviceId": "post_man_p2", "mobile": "+923405421402", "name": "Bilal",
              "gender": "male", "password": "Golang@12122"},
        tests=save("passenger2_otp", "r.data.tempOTP") + [
            'pm.environment.set("passenger2_mobile", "+923405421402");',
            'if (r.data) pm.environment.set("passenger2_id", r.data.passengerId);'],
        desc="A second passenger of the other gender, used later to show the gender swap on a seat."),

    req("Passenger 2 · Verify OTP", "POST", "/api/v1/otp/verify", "open_token",
        body={"mobile": "{{passenger2_mobile}}", "otp": "{{passenger2_otp}}", "operation": "ACTIVATE_PASSENGER"}),

    req("Passenger 2 · Login", "POST", "/api/v1/passenger/login", "open_token",
        body={"deviceId": "post_man_p2", "mobile": "{{passenger2_mobile}}", "password": "Golang@12122"},
        tests=save("passenger2_session", "r.data.sessionId") + [
            'if (r.data && r.data.passenger) pm.environment.set("passenger2_id", r.data.passenger.id);']),

    req("Passenger · Profile", "GET", "/api/v1/passenger/info", "passenger_session"),

    req("Passenger · Forgot Password", "GET", "/api/v1/passenger/password/forgot", "open_token",
        query=[q("mobile_number", "{{passenger1_mobile}}")],
        tests=save("passenger_reset_otp", "r.data.tempOTP"),
        desc="Verify with operation PASSENGER_FORGOT_PASSWORD and send the new password in the same body."),

    req("Passenger · Reset Password (change)", "POST", "/api/v1/passenger/password/reset", "passenger_session",
        body={"oldPassword": "Golang@12122", "newPassword": "AbC!123456"},
        tests=save("passenger_change_otp", "r.data.tempOTP"),
        desc="Parks the new password against the otp. Confirm it with operation PASSENGER_UPDATE_PASSWORD on /otp/verify."),

    req("Passenger · Logout", "GET", "/api/v1/passenger/logout", "passenger_session"),
]}

# -------------------------------------------------- 2 travel form ------------
form = {"name": "2 · Passenger Travel Form (point 7)", "item": [
    req("Set Weekly Form (bulk)", "PUT", "/api/v1/passenger/schedule", "passenger_session",
        body={"preferences": [
            {"dayOfWeek": 1, "direction": "pickup", "isEnabled": True, "location": "Bahria Town Phase 4",
             "lat": 33.5121, "lng": 73.0951, "scheduledTime": "07:15:00"},
            {"dayOfWeek": 1, "direction": "drop", "isEnabled": True, "location": "F-8 Markaz",
             "lat": 33.7101, "lng": 73.0441, "scheduledTime": "17:30:00"},
            {"dayOfWeek": 2, "direction": "pickup", "isEnabled": True, "location": "Bahria Town Phase 4",
             "lat": 33.5121, "lng": 73.0951, "scheduledTime": "07:15:00"},
            {"dayOfWeek": 2, "direction": "drop", "isEnabled": False, "location": "", "lat": 0, "lng": 0,
             "scheduledTime": ""},
            {"dayOfWeek": 3, "direction": "pickup", "isEnabled": True, "location": "Gulberg Greens",
             "lat": 33.6180, "lng": 73.1560, "scheduledTime": "07:40:00"},
            {"dayOfWeek": 3, "direction": "drop", "isEnabled": True, "location": "Blue Area",
             "lat": 33.7180, "lng": 73.0640, "scheduledTime": "18:00:00"},
        ]},
        desc=("The whole week in one call, one row per day per direction.\n\n"
              "Tuesday shows requirement 17.2: pickup is on, drop is off, so this passenger rides in "
              "the morning and makes their own way home.\n\n"
              "Wednesday shows requirement 17.3: picked up from Gulberg Greens but dropped at Blue Area, "
              "a different place from the morning.\n\n"
              "This form never creates a shift. It is only the sheet the manager reads while building one.")),

    req("Get Weekly Form", "GET", "/api/v1/passenger/schedule", "passenger_session"),

    req("Passenger 2 · Set Weekly Form", "PUT", "/api/v1/passenger/schedule", "passenger2_session",
        body={"preferences": [
            {"dayOfWeek": 1, "direction": "pickup", "isEnabled": True, "location": "DHA Phase 2",
             "lat": 33.5350, "lng": 73.1350, "scheduledTime": "07:20:00"},
            {"dayOfWeek": 1, "direction": "drop", "isEnabled": True, "location": "F-8 Markaz",
             "lat": 33.7101, "lng": 73.0441, "scheduledTime": "17:30:00"},
        ]}),
]}

# ------------------------------------------------------- 3 group -------------
group = {"name": "3 · Group (Fleet)", "item": [
    req("Create Group", "POST", "/api/v1/group", "owner_session",
        body={"name": "Morning School Fleet", "description": "Islamabad school run",
              "vehicleIds": ["{{owner_vehicle_id}}"]},
        tests=save("group_id", "r.data.groupId"),
        desc=("The creating driver becomes the owner, which in this product is the same thing as the "
              "manager. A group must be born with at least one vehicle, and every vehicle must already "
              "have declared its seats.")),

    req("My Groups", "GET", "/api/v1/group/mine", "owner_session",
        query=[q("page", "1")]),

    req("Find Groups To Join (driver)", "GET", "/api/v1/group/search", "driver2_session",
        query=[q("page", "1"), q("search", "", True)],
        tests=[
            "const r = pm.response.json();",
            "if (r.data && r.data.groups && r.data.groups.length) {",
            '    pm.environment.set("group_id", r.data.groups[0].id);',
            '    console.log("found:", r.data.groups.map(g => g.name + " [" + (g.myStatus || "not requested") + "]"));',
            "}",
            "",
        ],
        desc=("This is how a group id is discovered at all. Without it a driver could never "
              "learn the id that the join request needs.\n\n"
              "Each row carries myStatus, so the app can tell 'ask to join' apart from 'waiting' "
              "without a second call. Empty means they have never asked.\n\n"
              "Pass search to filter by name. Passengers hit the same thing at "
              "GET /api/v1/passenger/groups/search.")),

    req("Driver · Request To Join (driver + vehicle)", "POST", "/api/v1/group/request/driver", "driver2_session",
        body={"groupId": "{{group_id}}", "joinType": "both", "vehicleIds": ["{{driver2_vehicle_id}}"]},
        desc=("joinType is one of both, driver_only or vehicle_only (requirement 3).\n\n"
              "driver_only must carry no vehicles. vehicle_only and both must carry at least one.\n\n"
              "Nothing is granted here, the owner still decides. A driver who was turned down before "
              "can simply ask again.")),

    req("Passenger · Request To Join", "POST", "/api/v1/passenger/group/request", "passenger_session",
        body={"groupId": "{{group_id}}"},
        desc="Requirement 23. Note this runs on a passenger token, not a driver one."),

    req("Passenger 2 · Request To Join", "POST", "/api/v1/passenger/group/request", "passenger2_session",
        body={"groupId": "{{group_id}}"}),

    req("List Pending Requests", "GET", "/api/v1/group/requests", "owner_session",
        query=[q("group_id", "{{group_id}}")],
        tests=[
            "const r = pm.response.json();",
            "if (r.data) {",
            '    if (r.data.members && r.data.members.length) pm.environment.set("member_request_id", r.data.members[0].id);',
            '    if (r.data.vehicles && r.data.vehicles.length) pm.environment.set("vehicle_request_id", r.data.vehicles[0].id);',
            "    if (r.data.passengers && r.data.passengers.length) {",
            '        pm.environment.set("passenger_request_id", r.data.passengers[0].id);',
            '        if (r.data.passengers.length > 1) pm.environment.set("passenger2_request_id", r.data.passengers[1].id);',
            "    }",
            '    console.log("pending:", (r.data.members||[]).length, "drivers,", (r.data.vehicles||[]).length, "vehicles,", (r.data.passengers||[]).length, "passengers");',
            "}",
        ],
        desc="Needs group.manage_members. Saves the membership row ids the next call decides on."),

    req("Decide Requests (bulk)", "PATCH", "/api/v1/group/requests", "owner_session",
        body={"groupId": "{{group_id}}", "decisions": [
            {"memberType": "driver", "memberId": "{{member_request_id}}", "action": "approve"},
            {"memberType": "vehicle", "memberId": "{{vehicle_request_id}}", "action": "approve"},
            {"memberType": "passenger", "memberId": "{{passenger_request_id}}", "action": "approve"},
            {"memberType": "passenger", "memberId": "{{passenger2_request_id}}", "action": "approve"},
        ]},
        desc=("One call settles the whole queue. memberId is the id of the membership row, not of the "
              "driver or passenger behind it.\n\n"
              "action is approve, reject or remove. approve and reject only act on somebody still "
              "waiting, remove only on somebody already inside.\n\n"
              "The response splits into applied and skipped, and every skipped line says why, so a "
              "half valid batch still tells you exactly what happened. Deciding on a vehicle "
              "additionally needs group.manage_vehicles.")),

    req("Promote Sub Manager (bulk)", "PATCH", "/api/v1/group/submanagers", "owner_session",
        body={"groupId": "{{group_id}}", "promote": ["{{driver2_id}}"], "demote": []},
        desc=("Requirement 16. The rule you insisted on is enforced here: a driver can only be promoted "
              "if they are ALREADY an approved member of this same fleet. Anybody else comes back in "
              "skipped with the reason, never promoted.\n\n"
              "The owner can never be demoted, that would leave the fleet with nobody in charge.\n\n"
              "A sub manager gets the four shift permissions and none of the membership ones.")),

    req("Group Details (rosters)", "GET", "/api/v1/group", "owner_session",
        query=[q("group_id", "{{group_id}}")],
        desc="The fleet with its driver, vehicle and passenger rosters. Only somebody already inside the fleet may look."),

    req("Passenger Travel Forms (manager view)", "GET", "/api/v1/group/passengers/schedules", "owner_session",
        query=[q("group_id", "{{group_id}}"), q("day_of_week", "1"), q("direction", "pickup"), q("page", "1")],
        desc=("This is the sheet requirement 10 describes: who wants to travel on this day in this "
              "direction, from where, at what time, and which gender seat they need.\n\n"
              "Leave day_of_week and direction off to get the whole week.\n\n"
              "Gated on shift.assign_seats rather than plain membership, since it carries home "
              "addresses, so the owner and sub managers see it and an ordinary joined driver does not.")),

    req("Find Groups To Join (passenger)", "GET", "/api/v1/passenger/groups/search", "passenger_session",
        query=[q("page", "1")],
        desc="The same discovery list seen through a passenger token."),

    req("My Groups (passenger)", "GET", "/api/v1/passenger/groups", "passenger_session",
        query=[q("page", "1")],
        desc="The fleets a passenger has been let into or is still waiting on, so they can see where their request stands."),

    req("Driver Leaves The Group", "DELETE", "/api/v1/group/leave", "driver2_session",
        query=[q("group_id", "{{group_id}}")],
        desc=("The brief says people can join AND leave, and until now only the manager could "
              "take somebody out.\n\n"
              "Refused while shifts are still expecting this driver behind the wheel, since a "
              "driver cannot be swapped out automatically. Cancel or reassign those first.\n\n"
              "Their vehicles leave the fleet with them, and any sub manager role is dropped. "
              "The owner can never leave, they delete the group instead.\n\n"
              "Run this AFTER you have finished with the sub manager shift, or it will refuse.")),

    req("Passenger Leaves The Group", "DELETE", "/api/v1/passenger/group/leave", "passenger_session",
        query=[q("group_id", "{{group_id}}")],
        desc=("A departing passenger's seats on upcoming shifts are freed automatically and each "
              "affected shift's taken count is put back in step, so they never leave a ghost "
              "sitting in a seat nobody can fill.\n\n"
              "The same release happens when a manager removes a passenger through the bulk "
              "decide endpoint.")),

    req("Delete Group", "DELETE", "/api/v1/group", "owner_session",
        query=[q("group_id", "{{group_id}}")],
        desc="Refused while the fleet still has shifts ahead of it. Cancel those first. Run this last, it ends the flow."),
]}

# ------------------------------------------------------- 4 shifts ------------
shifts = {"name": "4 · Shifts", "item": [
    req("Create Morning PICKUP Shift (+ template)", "POST", "/api/v1/shift", "owner_session",
        body={
            "groupId": "{{group_id}}",
            "vehicleId": "{{owner_vehicle_id}}",
            "driverId": "{{owner_driver_id}}",
            "direction": "pickup",
            "startDatetime": "2026-09-14 07:00:00",
            "estimatedEndDatetime": "2026-09-14 08:30:00",
            "startLocation": "Bahria Town Phase 4",
            "endLocation": "Roots School F-8",
            "routeDetails": "Via Expressway",
            "makeTemplate": True,
            "templateName": "Monday morning pickup",
            "daysOfWeek": [1, 2, 3, 4, 5],
            "stops": [
                {"location": "Bahria Town Phase 4", "lat": 33.5121, "lng": 73.0951,
                 "scheduledTime": "07:15:00",
                 "seats": [{"seatNumber": 1, "gender": "female", "passengerId": "{{passenger1_id}}"}]},
                {"location": "DHA Phase 2", "lat": 33.5350, "lng": 73.1350,
                 "scheduledTime": "07:35:00",
                 "seats": [{"seatNumber": 2, "gender": "male", "passengerId": "{{passenger2_id}}"}]},
                {"location": "Roots School F-8", "lat": 33.7101, "lng": 73.0441,
                 "scheduledTime": "08:20:00", "seats": []},
            ],
        },
        tests=save("shift_id", "r.data.shiftId") + ['if (r.data) pm.environment.set("template_id", r.data.templateId);'],
        desc=("The stops are the route, in the order you send them, and each seat hangs off the stop "
              "its passenger waits at (your ordered stop requirement). The last stop here is the "
              "destination and carries nobody.\n\n"
              "Every seat of the vehicle gets a row, so this 7 seat van comes back with 2 assigned "
              "and 5 empty seats a manager can fill later.\n\n"
              "makeTemplate saves the shape for reuse, exactly like ride templates. There is no "
              "separate build from template api, the app refetches a template and posts it back here.\n\n"
              "Rejected if the vehicle or the driver is already promised at that hour, on another "
              "shift OR on a carpool ride (requirements 19 to 22).")),

    req("Create Evening DROP Shift", "POST", "/api/v1/shift", "owner_session",
        body={
            "groupId": "{{group_id}}",
            "vehicleId": "{{owner_vehicle_id}}",
            "driverId": "{{owner_driver_id}}",
            "direction": "drop",
            "startDatetime": "2026-09-14 17:00:00",
            "estimatedEndDatetime": "2026-09-14 18:30:00",
            "startLocation": "Roots School F-8",
            "endLocation": "Bahria Town Phase 4",
            "routeDetails": "Via Kashmir Highway",
            "makeTemplate": False,
            "daysOfWeek": [1, 2, 3, 4, 5],
            "stops": [
                {"location": "Roots School F-8", "lat": 33.7101, "lng": 73.0441,
                 "scheduledTime": "17:00:00", "seats": []},
                {"location": "F-8 Markaz", "lat": 33.7101, "lng": 73.0441,
                 "scheduledTime": "17:30:00",
                 "seats": [{"seatNumber": 1, "gender": "female", "passengerId": "{{passenger1_id}}"}]},
                {"location": "DHA Phase 2", "lat": 33.5350, "lng": 73.1350,
                 "scheduledTime": "18:10:00",
                 "seats": [{"seatNumber": 2, "gender": "male", "passengerId": "{{passenger2_id}}"}]},
            ],
        },
        tests=save("drop_shift_id", "r.data.shiftId"),
        desc=("The same people going home is a COMPLETELY SEPARATE shift with its own route "
              "(requirements 12, 13 and 17.4). Note the stop order is reversed and the drop points "
              "differ from the morning pickup points.\n\n"
              "It does not clash with the morning shift because the time windows do not overlap.")),

    req("Shift Detail", "GET", "/api/v1/shift/detail", "owner_session",
        query=[q("shift_id", "{{shift_id}}")],
        desc=("Returns the shift, its stops in order, and every seat.\n\n"
              "The shift carries both the driver's number and the number of the manager who built it, "
              "which is the contact information you asked to be on every shift.\n\n"
              "Visible to any approved member of the fleet, and to a passenger holding a seat on it.")),

    req("Group Shift Roster", "GET", "/api/v1/shift", "owner_session",
        query=[q("group_id", "{{group_id}}"), q("direction", "pickup"), q("status", "active"),
               q("page", "1"), q("driver_id", "", True),
               q("start_time", "2026-09-14 00:00:00", True),
               q("estimated_end_time", "2026-09-20 23:59:59", True)],
        desc="The manager's roster view. Every filter is optional."),

    req("My Shifts (driver)", "GET", "/api/v1/shift/mine", "owner_session",
        query=[q("page", "1")],
        desc="What this driver is driving."),

    req("My Shifts (passenger)", "GET", "/api/v1/passenger/shifts", "passenger_session",
        query=[q("page", "1")],
        desc="The same handler seen through a passenger token: the shifts they hold a seat on, morning and evening both."),

    req("Passenger · Shift Detail", "GET", "/api/v1/passenger/shift/detail", "passenger_session",
        query=[q("shift_id", "{{shift_id}}")],
        desc="A passenger can open a shift they are seated on, and gets the driver and manager contacts with it."),

    req("Swap Male Rider For Female (one call)", "PUT", "/api/v1/shift/seats", "owner_session",
        body={"shiftId": "{{shift_id}}", "seats": [
            {"seatNumber": 2, "gender": "female", "passengerId": "{{passenger1_id}}", "stopSequence": 1},
            {"seatNumber": 1, "gender": "male", "passengerId": "{{passenger2_id}}", "stopSequence": 2},
        ]},
        desc=("The swap you asked for. Seat 2 held a male rider and now holds a female one, and seat 1 "
              "goes the other way, in a single request.\n\n"
              "The rule that still holds: a seat's declared gender and the person sitting in it must "
              "always agree, so nobody ever lands in a seat kept for the other gender. Try changing "
              "one gender field and leaving the passenger, it will be refused.\n\n"
              "stopSequence points at the stop that passenger now waits at, counting from 1.\n\n"
              "Send passengerId as \"\" to free a seat. Everybody whose place actually changed is "
              "notified, people who kept their seat are left alone.")),

    req("Free A Seat", "PUT", "/api/v1/shift/seats", "owner_session",
        body={"shiftId": "{{shift_id}}", "seats": [
            {"seatNumber": 2, "gender": "", "passengerId": "", "stopSequence": 0},
        ]},
        desc="An empty passengerId frees the seat and clears the gender it was kept for. seatsTaken is recomputed."),

    req("Sub Manager Builds A Shift", "POST", "/api/v1/shift", "driver2_session",
        body={
            "groupId": "{{group_id}}",
            "vehicleId": "{{driver2_vehicle_id}}",
            "driverId": "{{driver2_id}}",
            "direction": "pickup",
            "startDatetime": "2026-09-15 07:00:00",
            "estimatedEndDatetime": "2026-09-15 08:30:00",
            "startLocation": "Gulberg Greens",
            "endLocation": "Roots School F-8",
            "routeDetails": "Via Islamabad Highway",
            "makeTemplate": False,
            "daysOfWeek": [2],
            "stops": [
                {"location": "Gulberg Greens", "lat": 33.6180, "lng": 73.1560,
                 "scheduledTime": "07:20:00",
                 "seats": [{"seatNumber": 1, "gender": "female", "passengerId": "{{passenger1_id}}"}]},
                {"location": "Roots School F-8", "lat": 33.7101, "lng": 73.0441,
                 "scheduledTime": "08:20:00", "seats": []},
            ],
        },
        tests=save("submanager_shift_id", "r.data.shiftId"),
        desc=("Requirement 16 end to end: the sub manager takes the shift building load off the owner. "
              "Same endpoint, different token.\n\n"
              "Try the group membership calls on this token and they will come back "
              "'Operation is not permitted', because a sub manager holds the shift permissions and "
              "none of the membership ones.")),

    req("Clash Check · Same Vehicle, Overlapping Time", "POST", "/api/v1/shift", "owner_session",
        body={
            "groupId": "{{group_id}}",
            "vehicleId": "{{owner_vehicle_id}}",
            "driverId": "{{owner_driver_id}}",
            "direction": "pickup",
            "startDatetime": "2026-09-14 07:30:00",
            "estimatedEndDatetime": "2026-09-14 09:00:00",
            "startLocation": "Somewhere else",
            "endLocation": "Another school",
            "routeDetails": "Overlaps the morning shift on purpose",
            "makeTemplate": False,
            "daysOfWeek": [1],
            "stops": [{"location": "Somewhere else", "lat": 33.6, "lng": 73.1,
                       "scheduledTime": "07:40:00", "seats": []}],
        },
        desc=("EXPECTED TO FAIL with 'Vehicle already has a ride or a shift scheduled for this "
              "duration'. This proves requirements 19 and 20.\n\n"
              "Create a carpool ride on this same vehicle at this time and the shift is refused too, "
              "and the mirror holds: creating a carpool ride while a shift is on gets refused as well "
              "(requirements 21 and 22).")),

    req("Reschedule Shift (move the time)", "PATCH", "/api/v1/shift", "owner_session",
        body={
            "shiftId": "{{shift_id}}",
            "startDatetime": "2026-09-14 07:20:00",
            "estimatedEndDatetime": "2026-09-14 08:50:00",
            "routeDetails": "Via Expressway, delayed for roadworks",
            "stopTimes": [
                {"sequenceNumber": 1, "scheduledTime": "07:35:00"},
                {"sequenceNumber": 2, "scheduledTime": "07:55:00"},
                {"sequenceNumber": 3, "scheduledTime": "08:40:00"},
            ],
        },
        desc=("Pushes the whole trip back twenty minutes WITHOUT losing the seat plan. Cancelling "
              "and rebuilding would throw the seats away, and is refused outright inside the two "
              "hour window.\n\n"
              "Every field except shiftId is optional, but the time window has to be sent whole or "
              "not at all. stopTimes matches stops by their sequence number, so stop ids and the "
              "seats hanging off them survive.\n\n"
              "Moving the window re-runs the full clash check, ignoring this shift's own row. "
              "Everybody aboard is notified.\n\n"
              "The vehicle and driver are deliberately NOT changeable here, swapping either can "
              "invalidate every seat, so that stays a cancel and rebuild.")),

    req("Cancel Shift", "DELETE", "/api/v1/shift", "owner_session",
        query=[q("shift_id", "{{submanager_shift_id}}")],
        desc=("Soft cancel. Refused when the shift starts within 2 hours, mirroring the carpool ride "
              "guard, so use a shift comfortably in the future.\n\n"
              "The driver and every seated passenger are told, with the manager's contact in the message.")),
]}

# --------------------------------------------------- 5 templates -------------
templates = {"name": "5 · Templates", "item": [
    req("Get Shift Templates", "GET", "/api/v1/shift/templates", "owner_session",
        query=[q("group_id", "{{group_id}}")],
        tests=[
            "const r = pm.response.json();",
            "if (r.data && r.data.templates && r.data.templates.length) {",
            '    pm.environment.set("template_id", r.data.templates[0].id);',
            '    console.log("templates:", r.data.templates.length);',
            "}",
        ],
        desc=("A template comes back with its stops, its seat plan and its vehicle.\n\n"
              "This is how a shift is rebuilt: the app fetches a template, prefills the create form "
              "with it, changes the date, and posts it to POST /api/v1/shift. Exactly the pattern your "
              "ride templates already use, which is why there is no build from template endpoint.\n\n"
              "Template seats point at their stop by sequence rather than by id, so a template stays "
              "valid for a brand new shift.")),

    req("Delete Shift Template", "DELETE", "/api/v1/shift/template", "owner_session",
        query=[q("shift_template_id", "{{template_id}}")],
        desc="Deletes the template with its stops and seats. Needs shift.manage_templates."),
]}


# ----------------------------------------------------- 6 rideshare ----------
# The existing carpool feature. Only one line of it changed for the fleet work,
# the clash check, but its saved responses had drifted from what the server
# actually returns, so they are recaptured here.
rideshare = {"name": "6 · Rideshare (existing carpool)", "item": [
    req("Create Ride", "POST", "/api/v1/ride/create", "owner_session",
        body={
            "startDatetime": "2026-09-15 15:30:00", "estimatedEndDatetime": "2026-09-15 17:00:00",
            "numberOfSeats": 3, "startLocation": "Location A", "endLocation": "Location B",
            "routePoints": ["LocationA1", "LocationA2"], "fare": 20.5,
            "routeDetails": "Via Highway 1", "vehicleId": "{{owner_vehicle_id}}",
            "makeTemplate": True, "isRecurring": False, "frequency": 1, "period": 1,
            "daysOfWeek": [1],
        },
        tests=save("ride_id", "r.data.id"),
        desc=("Unchanged except for one added guard: the vehicle and the driver are now also "
              "checked against the fleet shifts, so a carpool ride cannot be created on top of a "
              "shift. The second saved response shows that refusal.")),

    req("Driver Rides", "GET", "/api/v1/driver/rides", "owner_session",
        query=[q("page", "1"), q("status", "all")]),

    req("Get One Ride", "GET", "/api/v1/ride", "open_token",
        query=[q("ride_id", "{{ride_id}}")]),

    req("Filtered Rides", "GET", "/api/v1/ride/filtered", "open_token",
        query=[q("page", "1"), q("search", "LocationA1")]),

    req("Ride Templates", "GET", "/api/v1/ride/templates", "owner_session"),

    req("Update Ride", "PATCH", "/api/v1/driver/ride/update", "owner_session",
        query=[q("ride_id", "{{ride_id}}")],
        body={"numberOfSeats": 2}),

    req("Passenger Ride Request", "POST", "/api/v1/passenger/ride/request", "open_token",
        body={
            "startDatetime": "2026-09-15 15:30:00", "estimatedEndDatetime": "2026-09-15 17:00:00",
            "numberOfSeats": 2, "startLocation": "Location A", "endLocation": "Location B",
            "routeDetails": "via gt road", "contactNumber": "+923301221121",
        }),

    req("Get Ride Requests", "GET", "/api/v1/driver/ride/requests", "owner_session",
        query=[q("page", "1")]),

    req("Get Announcements", "GET", "/api/v1/announcements", "open_token"),

    req("Driver Notifications", "GET", "/api/v1/user/notifications", "owner_session"),

    req("Admin · List Rides", "GET", "/api/v1/admin/rides", "admin_session",
        query=[q("page", "1")]),

    req("Admin · List Drivers", "GET", "/api/v1/admin/drivers", "admin_session",
        query=[q("page", "1")]),

    req("Admin · List Vehicles", "GET", "/api/v1/admin/vehicles", "admin_session",
        query=[q("page", "1")],
        desc="Now carries numberOfSeats, hasAC and hasHeating, and its totalPages is correct (it used to count a query that already had the page limit applied)."),
]}

collection = {
    "info": {
        "_postman_id": "b7c41f02-5e6a-4d38-9c11-3a7f0e2b4d91",
        "name": "SathSawari · Shift Management",
        "description": (
            "Every endpoint added by the group / fleet shift management feature, in the order you "
            "would actually exercise them.\n\n"
            "SETUP\n"
            "1. Use the same environment as your existing Rideshare collection, it needs base-url and "
            "open_token.\n"
            "2. Run the folders top to bottom. Almost every id is captured into an environment "
            "variable by a test script, so the flow chains on its own.\n"
            "3. The OTPs come back in the response as tempOTP while you are not on production, so the "
            "verify steps chain automatically too.\n\n"
            "TOKENS USED\n"
            "  open_token          public endpoints\n"
            "  admin_session       admin, roles and permissions\n"
            "  owner_session       the driver who owns the fleet, ie the manager\n"
            "  driver2_session     a second driver, later a sub manager\n"
            "  passenger_session   a female passenger\n"
            "  passenger2_session  a male passenger\n\n"
            "Folder 4 contains two requests that are MEANT to fail, they are the proof of the clash "
            "and gender rules. Each one says so in its description."
        ),
        "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
    },
    "item": [admin, accounts, form, group, shifts, templates, rideshare],
    "variable": [
        {"key": "base-url", "value": "http://localhost:8080", "type": "string"},
    ],
}

out = os.path.join(ROOT, "SathSawari-ShiftManagement.postman_collection.json")
with open(out, "w") as f:
    json.dump(collection, f, indent=2)

n = sum(len(folder["item"]) for folder in collection["item"])
print(f"wrote {out}")
print(f"folders: {len(collection['item'])}, requests: {n}")
