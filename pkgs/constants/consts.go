package constants

// response message
const (
	// SUCCESS
	Registered_Successfully       = "%v registered successfully"
	Updated_Successfully          = "%v updated successfully"
	Deleted_Successfully          = "%v updated successfully"
	Loggedin_Successfully         = "%s loggedin successfully"
	Loggedout_Successfully        = "%s loggedout successfully"
	Created_Successfully          = "%s created successfully"
	Added_Successfully            = "%s added successfully"
	Request_Received_Successfully = "Your request has been received."
	SENT_OTP_Successfully         = "OTP has been sent to your registered number"
	Success                       = "Success"
	Success_Info                  = "%s successfully"

	// ERROR
	Registeration_Failed    = "Failed to register %s"
	Update_Failed           = "Failed to update %s"
	DELETE_Failed           = "Failed to update %s"
	Creation_Failed         = "Failed to create %s"
	Missing_Data            = "%v is missing"
	Invalid_Data            = "%v is invalid"
	Failed_To_Do_Job        = "Failed to %s"
	Login_Failed            = "Login failed, please use correct credentials"
	Unable_To_Do_Job        = "Unable to %s"
	Driver_Not_Found        = "Unknown driver"
	Vehicle_Not_Found       = "vehicle details are invalid"
	Ride_Not_Found          = "There are no rides available"
	Invalid_Password        = "Invalid password"
	Invalid_Session         = "Invalid session"
	General_Error           = "Error occured please, try again later"
	Unknown_Error           = "Unknown error"
	Operation_Not_Permitted = "Operation is not permitted"
	Unverified_Driver       = "Unverified driver"
	Perform_this_operation  = "Perform this operation"
	General_Unknown         = "Unknown %s"
	Not_Found               = "%s not found"

	// ride share capacity errors
	Ride_Seats_Exceed_Vehicle = "Number of seats cannot be more than the vehicle's %d seat(s)"
	Ride_Seats_Below_Booked   = "Number of seats cannot be less than the %d seat(s) already booked"
	Ride_Clash                = "%s already has a ride from %s to %s"
	Ride_Series_Overlap       = "The rides of this series would overlap each other on %s, each ride has to end before the next one starts"
	Ride_Has_Ended            = "This ride has already ended and cannot be made active again"

	// pick & drop errors
	Passenger_Not_Found             = "Unknown passenger"
	Service_Not_Found               = "Pick & Drop service not found"
	Request_Not_Found               = "Request not found"
	Advertisement_Not_Found         = "Advertisement not found"
	Not_Service_Owner               = "Only the owner of a Pick & Drop service can do this"
	Not_Service_Member              = "You are not a member of a Pick & Drop service"
	Already_Owns_Service            = "You own a Pick & Drop service, so you cannot start or join another one"
	Already_In_Service              = "You already belong to a Pick & Drop service, leave it first"
	Pending_Service_Request         = "You already have a pending Pick & Drop request, withdraw it first"
	Owner_Cannot_Leave              = "The owner cannot leave their own Pick & Drop service, disable it instead"
	Vehicle_Already_In_Service      = "Vehicle %s is already offered to a Pick & Drop service"
	Vehicle_Not_In_Service          = "This vehicle is not part of your Pick & Drop service"
	Driver_Not_In_Service           = "This driver is not an approved driver of your Pick & Drop service"
	Passenger_Not_In_Service        = "%s is not an approved passenger of your Pick & Drop service"
	Advertisement_Limit             = "%d Pick & Drop vehicle(s) allow %d advertisement(s), delete one or add a vehicle first"
	Service_Has_Active_Shifts       = "This Pick & Drop service still has %d active shift(s), delete them first"
	On_Active_Shifts                = "%s still assigned to %d active shift(s), remove that assignment first"
	Request_Already_Decided         = "This request is already %s"
	Only_Approved_Removable         = "Only an approved member can be removed"
	Vehicle_Driver_Not_Approved     = "Approve the driver of vehicle %s before the vehicle itself"
	Vehicles_Join_Needs_Vehicle     = "Joining with vehicles only needs at least one vehicle on offer"
	Last_Vehicle_Of_Vehicles_Member = "You joined with vehicles only, so your last vehicle cannot be taken out, leave the service instead"
	Driver_Vehicles_Only            = "%s joined your Pick & Drop service with vehicles only and cannot drive a shift"

	// shift errors
	Shift_Not_Found           = "Shift not found"
	Vehicle_Seats_Missing     = "Vehicle %s must have its number of seats set before it can be used for a shift"
	Vehicle_Capacity_Exceeded = "Vehicle %s has %d seat(s), it cannot carry %d passenger(s)"
	Resource_Shift_Clash      = "%s is already on the shift %q on %s from %s to %s"
	Resource_Ride_Clash       = "%s already has a ride from %s to %s that overlaps this shift"
	Not_Shift_Driver          = "You are not the driver of this shift"
	Not_Shift_Passenger       = "You are not a passenger of this shift"
	Not_An_Occurrence         = "This shift does not run on %s"
	Occurrence_Started        = "This trip has already started"
	Shift_In_The_Past         = "A shift cannot start in the past"
	Passenger_Stop_Missing    = "%d passenger(s) are waiting at stop %d, move them before shortening the route"
	Shift_Not_Active          = "This shift is %s and can no longer be changed"
	Passenger_Not_On_Shift    = "Passenger %s is not on this shift"
	Passenger_Already_On      = "%s is already on this shift"
	End_Date_In_The_Past      = "The end date of a shift cannot be in the past"
)

// keys
const (
	Sessoin_KEY            = "session"
	Operaion_KEY           = "operation"
	User_KEY               = "user_id"
	Vehicle_KEY            = "vehicle_id"
	Pin_Key                = "pin"
	Encrypted_User_KEY     = "encrypted_user_id"
	Page_Key               = "page"
	Start_Time_Key         = "start_time"
	Extimated_End_Time_Key = "estimated_end_time"
	Start_Loc_Key          = "start_location"
	End_Loc_Key            = "end_location"
	Search_Loc_Key         = "search"
	Status_Key             = "status"
	Ride_Key               = "ride_id"
	Ride_Request_Key       = "request_id"
	Seat_Key               = "seat_id"
	Booking_Key            = "booking_id"
	Ride_Template_Key      = "ride_template_id"
	Language_Key           = "lang"
	Type_Key               = "type"
	Shift_Key              = "shift_id"
	Service_Key            = "service_id"
	Request_Key            = "request_id"
	Advertisement_Key      = "advertisement_id"
	Day_Of_Week_Key        = "day_of_week"
	Days_Of_Week_Key       = "days_of_week"
	Date_Key               = "date"
	Start_Date_Key         = "start_date"
	End_Date_Key           = "end_date"
	End_Time_Key           = "end_time"
	Exclude_Shift_Key      = "exclude_shift_id"
	Driver_Key             = "driver_id"
	Passenger_Key          = "passenger_id"
)

// ride status
const (
	Ride_Status_Active   = "active"
	Ride_Status_InActive = "inactive"
	Ride_Status_All      = "all"
)

const (
	Status_Active          = "active"
	Status_InActive        = "inactive"
	Status_PendingApproval = "pending"
)

// notification titles
const (
	NOTIFICATION_TITLE_RIDE_CREATION  = "Ride Created"
	NOTIFICATION_TITLE_PIN_CREATION   = "Pin Created"
	NOTIFICATION_TITLE_RIDE_BOOKED    = "Ride Booked"
	NOTIFICATION_TITLE_SMS_TO_SERVICE = "SathSawari sent an OTP"

	NOTIFICATION_TITLE_PICKDROP_REQUEST  = "New Pick & Drop Request"
	NOTIFICATION_TITLE_PICKDROP_DECISION = "Pick & Drop Request Update"
	NOTIFICATION_TITLE_PICKDROP_UPDATE   = "Pick & Drop Update"
	NOTIFICATION_TITLE_SHIFT_CREATED     = "Shift Scheduled"
	NOTIFICATION_TITLE_SHIFT_UPDATED     = "Shift Updated"
	NOTIFICATION_TITLE_SHIFT_DELETED     = "Shift Cancelled"
	NOTIFICATION_TITLE_SHIFT_ASSIGNED    = "Shift Assigned"
	NOTIFICATION_TITLE_SHIFT_PASSENGER   = "Shift Seat Update"
	NOTIFICATION_TITLE_SHIFT_ABSENCE     = "Passenger Attendance"
	NOTIFICATION_TITLE_SHIFT_DRIVER      = "Driver Update"
	NOTIFICATION_TITLE_SHIFT_REMINDER    = "Shift Starting Soon"
)

// notification message
const (
	NOTIFICATION_MESSAGE_RIDE_CREATION      = "You have created a ride"
	NOTIFICATION_MESSAGE_PIN_CREATION       = "Please remember your secure PIN “%s” for future operations in the app"
	NOTIFICATION_MESSSGE_RIDE_BOOKED_DRIVER = "%v seat(s) booked for vehicle %s."
	NOTIFICATION_MESSAGE_PIN_UPDATED        = "Your pin has been updated, Please do not share it with anyone"
	NOTIFICATION_MESSAGE_SMS_TO_SERVICE     = "<#> Sathsawari verification code: %s. Do not share this code with anyone.\n%s"

	// pick & drop membership notifications
	NOTIFICATION_MESSAGE_PICKDROP_JOIN_REQUEST     = "%s has asked to join %s as a %s"
	NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_OFFER    = "%s has offered %d vehicle(s) to %s"
	NOTIFICATION_MESSAGE_PICKDROP_APPROVED         = "Your request to join %s has been approved"
	NOTIFICATION_MESSAGE_PICKDROP_REJECTED         = "Your request to join %s has been rejected"
	NOTIFICATION_MESSAGE_PICKDROP_REMOVED          = "You have been removed from %s"
	NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_APPROVED = "Vehicle %s has been approved for %s"
	NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_REJECTED = "Vehicle %s has been rejected for %s"
	NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_REMOVED  = "Vehicle %s has been removed from %s"
	NOTIFICATION_MESSAGE_PICKDROP_VEHICLE_LEFT     = "%s has taken vehicle %s out of %s"
	NOTIFICATION_MESSAGE_PICKDROP_LEFT             = "%s has left %s"
	NOTIFICATION_MESSAGE_PICKDROP_WITHDRAWN        = "%s has withdrawn their request to join %s"
	NOTIFICATION_MESSAGE_PICKDROP_CLOSED           = "%s has been closed by its owner"

	// shift notifications
	NOTIFICATION_MESSAGE_SHIFT_PASSENGER_ADDED    = "You have been added to the shift %s of %s (%s). You are picked from %s at %s. Driver %s, vehicle %s."
	NOTIFICATION_MESSAGE_SHIFT_PASSENGER_REMOVED  = "You have been removed from the shift %s of %s."
	NOTIFICATION_MESSAGE_SHIFT_PASSENGER_MOVED    = "Your stop on the shift %s of %s is now %s at %s."
	NOTIFICATION_MESSAGE_SHIFT_DRIVER_ASSIGNED    = "You are driving the shift %s of %s (%s) from %s to %s with vehicle %s."
	NOTIFICATION_MESSAGE_SHIFT_DRIVER_UNASSIGNED  = "You are no longer driving the shift %s of %s."
	NOTIFICATION_MESSAGE_SHIFT_UPDATED            = "The shift %s of %s now runs %s from %s to %s. Driver %s, vehicle %s."
	NOTIFICATION_MESSAGE_SHIFT_PASSENGERS_CHANGE  = "The passengers of the shift %s of %s have changed, %d of %d seat(s) are now taken."
	NOTIFICATION_MESSAGE_SHIFT_DELETED            = "The shift %s of %s has been cancelled."
	NOTIFICATION_MESSAGE_SHIFT_ABSENT             = "%s will be absent from the shift %s on %s (%s at %s)."
	NOTIFICATION_MESSAGE_SHIFT_PRESENT            = "%s will travel on the shift %s on %s after all (%s at %s)."
	NOTIFICATION_MESSAGE_SHIFT_DRIVER_UPDATE      = "%s, %s (vehicle %s, shift %s): %s. Sent at %s."
	NOTIFICATION_MESSAGE_SHIFT_REMINDER_DRIVER    = "Your shift %s of %s starts at %s from %s with vehicle %s."
	NOTIFICATION_MESSAGE_SHIFT_REMINDER_RIDER     = "Your shift %s of %s starts at %s. Be at %s by %s. Driver %s, vehicle %s."
	NOTIFICATION_MESSAGE_SHIFT_VEHICLE_ASSIGNED   = "Your vehicle %s is on the shift %s of %s (%s) from %s to %s, driven by %s."
	NOTIFICATION_MESSAGE_SHIFT_VEHICLE_UNASSIGNED = "Your vehicle %s is no longer on the shift %s of %s."
)

// genders
const (
	Gender_Male   = "male"
	Gender_Female = "female"
	Production    = "production"
)

// data
const (
	Postgress_PK_Size         = 2730
	MOBILE_NUMBER_QUERY       = "mobile_number"
	OTP_OPERATION             = "otp_operation"
	FIREBASE_CREDENTIALS_JSON = "FIREBASE_CREDENTIALS_JSON"
)

// limits
const (
	Name_Min_Len              = 3
	Name_Max_Len              = 128
	Username_Min_Len          = 5
	Username_Max_Len          = 50
	Password_Min_Len          = 8
	Password_Max_Len          = 20
	Long_Password_Min_Len     = 8
	Long_Password_Max_Len     = 70
	Pin_Len                   = 6
	VehicleNumber_Min_Len     = 1
	VehicleNumber_Max_Len     = 20
	VehicleInfo_Min_Len       = 1
	VehicleInfo_Max_Len       = 200
	MobileNumber_Min_Len      = 7
	MobileNumber_Max_Len      = 15
	Page_Min_Value            = 1
	Page_Max_Value            = 1000
	Number_Of_Seats_Min_Value = 1
	Number_Of_Seats_Max_Value = 80
	RouteDetails_Min_Len      = 3
	RouteDetails_Max_Len      = 200
	RoutePoints_Max_Len       = 5
	Fare_Min_Len              = 0
	Fare_Max_Len              = 100000
	Rating_Min_Value          = 1
	Rating_Max_Value          = 5
	General_Min_Len           = 1
	General_Max_Len           = 50
	Email_Min_Len             = 6
	Email_Max_Len             = 254
	Message_Min_Len           = 5
	Message_Max_Len           = 500
	OTP_Min_Len               = 6
	OTP_Max_Len               = 6
	Max_Start_Date_Gap        = 30
	Max_End_Date_Gap          = 2
	Max_Daily_Frequency       = 30
	Max_Weekly_Frequency      = 4
	Max_Monthly_Frequency     = 3

	Ride_Cancel_Min_Hours_Before_Start = 2

	// pick & drop and shift limits
	Service_Name_Min_Len            = 3
	Service_Name_Max_Len            = 100
	Title_Min_Len                   = 3
	Title_Max_Len                   = 100
	Description_Max_Len             = 500
	Location_Min_Len                = 1
	Location_Max_Len                = 200
	Bulk_Request_Max_Len            = 100
	Day_Of_Week_Min_Value           = 1
	Day_Of_Week_Max_Value           = 7
	Advertisement_Locations_Min_Len = 2
	Advertisement_Locations_Max_Len = 10
	Advertisements_Per_Vehicle      = 2
	Shift_Request_Locations_Min_Len = 2
	Shift_Request_Locations_Max_Len = 10
	Availability_Locations_Max_Len  = 6
	Shift_Locations_Min_Len         = 2
	Shift_Locations_Max_Len         = 20
	Shift_Reminder_Minutes          = 15
	Occurrence_Backfill_Days        = 1
	Occurrence_Horizon_Days         = 1
)

const (
	DateTimeLayout = "2006-01-02 15:04:05"

	// a pick & drop schedule is a calendar date plus a wall clock time in the
	// business time zone, kept as sortable strings the same way rides keep theirs
	Date_Layout            = "2006-01-02"
	Clock_Layout           = "15:04"
	Minute_Datetime_Layout = "2006-01-02 15:04"
)

const (
	DRIVER_TOKEN    = "driver_token"
	ADMIN_TOKEN     = "admin_token"
	PASSENGER_TOKEN = "passenger_token"
	OPEN_TOKEN      = "open_token"
	NO_TOKEN_TYPE   = "no_token_type" // such token whose type dont matter
)

// pick & drop membership status, shared by driver, vehicle and passenger rows. A row
// is never reused, every request is its own row so the membership history survives
const (
	Membership_Status_Pending   = "pending"
	Membership_Status_Approved  = "approved"
	Membership_Status_Rejected  = "rejected"
	Membership_Status_Withdrawn = "withdrawn"
	Membership_Status_Left      = "left"
	Membership_Status_Removed   = "removed"
	Membership_Status_Closed    = "closed"
	Membership_Status_All       = "all"
)

// join request kinds and the decisions an owner can take on them
const (
	Request_Type_Driver    = "driver"
	Request_Type_Vehicle   = "vehicle"
	Request_Type_Passenger = "passenger"

	Request_Action_Approve = "approve"
	Request_Action_Reject  = "reject"
	Request_Action_Remove  = "remove"

	// how a driver belongs to a service: able to drive its shifts, or only through the
	// vehicles they brought, which other drivers then drive
	Join_Type_Driver   = "driver"
	Join_Type_Vehicles = "vehicles"
)

// what the caller is to a pick & drop service
const (
	Service_Role_Owner     = "owner"
	Service_Role_Driver    = "driver"
	Service_Role_Passenger = "passenger"
	Service_Role_None      = "none"
)

// shift, shift passenger, occurrence and attendance states
const (
	Shift_Status_Active    = "active"
	Shift_Status_Completed = "completed"
	Shift_Status_Deleted   = "deleted"
	Shift_Status_All       = "all"

	Shift_Passenger_Active  = "active"
	Shift_Passenger_Removed = "removed"

	Removal_Reason_Owner           = "removed_by_owner"
	Removal_Reason_Left_Service    = "left_service"
	Removal_Reason_Removed_Service = "removed_from_service"
	Removal_Reason_Shift_Deleted   = "shift_deleted"
	Removal_Reason_Account_Deleted = "account_deleted"

	Occurrence_Status_Scheduled = "scheduled"
	Occurrence_Status_Completed = "completed"

	Attendance_Present = "present"
	Attendance_Absent  = "absent"
)

// OTP Operations
// NOTE: if u add anything here then also add in VerifyOTPOperations function in utils
const (
	ACTIVATE_DRIVER_OPERATION  = "ACTIVATE_DRIVER"
	ACTIVATE_VEHICLE_OPERATION = "ACTIVATE_VEHICLE"
	UPDATE_PASSWORD_OPERATION  = "UPDATE_PASSWORD"
	FORGOT_PASSWORD_OPERATION  = "FORGOT_PASSWORD"
	UPDATE_PIN_OPERATION       = "UPDATE_PIN"
	FORGOT_PIN_OPERATION       = "FORGOT_PIN"
	BOOK_RIDE_OPERATION        = "BOOK_RIDE"

	// a passenger account is a different record from a driver account even when
	// both are held by the same mobile number, so they carry their own operations
	ACTIVATE_PASSENGER_OPERATION        = "ACTIVATE_PASSENGER"
	PASSENGER_UPDATE_PASSWORD_OPERATION = "PASSENGER_UPDATE_PASSWORD"
	PASSENGER_FORGOT_PASSWORD_OPERATION = "PASSENGER_FORGOT_PASSWORD"
)

// SMS Keys
const (
	SMS_KEY_MOBILE_NUMBER = "mobileNumber"
	SMS_KEY_MESSAGE       = "message"
)

// Notifications Key
const (
	NOTIFICATION_KEY_RIDE_ID         = "rideId"
	NOTIFICATION_KEY_SHIFT_ID        = "shiftId"
	NOTIFICATION_KEY_SERVICE_ID      = "serviceId"
	NOTIFICATION_KEY_REQUEST_ID      = "requestId"
	NOTIFICATION_KEY_OCCURRENCE_DATE = "occurrenceDate"
	NOTIFICATION_KEY_LOCATION        = "location"
	NOTIFICATION_KEY_LAT             = "lat"
	NOTIFICATION_KEY_LNG             = "lng"
)

// deep linking url types
const (
	LIKE_TYPE_RIDE_URL         = "RIDE_URL"
	LIKE_TYPE_RIDE_REQUEST_URL = "RIDE_REQUEST_URL"
)
