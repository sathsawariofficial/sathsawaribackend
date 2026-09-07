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

	// group and shift errors
	Passenger_Not_Found    = "Unknown passenger"
	Group_Not_Found        = "Group not found"
	Shift_Not_Found        = "Shift not found"
	Role_Not_Found         = "Role not found"
	Permission_Not_Found   = "Permission not found"
	Not_Group_Member       = "You are not a member of this group"
	Already_Group_Member   = "You are already a member of this group"
	Request_Already_Exists = "A request has already been sent"
	Seat_Gender_Mismatch   = "Seat %d is reserved for a %s passenger"
	Seat_Already_Taken     = "Seat %d is already assigned"
	Driver_Busy            = "Driver already has a ride or a shift scheduled for this duration"
	Vehicle_Busy           = "Vehicle already has a ride or a shift scheduled for this duration"
	Vehicle_Seats_Missing  = "Vehicle must have its number of seats set before it can join a group"
	Not_System_Role        = "System roles cannot be deleted"
	Owner_Cannot_Leave     = "The owner cannot leave their own group, delete it instead"
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
	Group_Key              = "group_id"
	Shift_Key              = "shift_id"
	Shift_Template_Key     = "shift_template_id"
	Role_Key               = "role_id"
	Permission_Key         = "permission_id"
	Direction_Key          = "direction"
	Day_Of_Week_Key        = "day_of_week"
	Date_Key               = "date"
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
	NOTIFICATION_TITLE_RIDE_CREATION   = "Ride Created"
	NOTIFICATION_TITLE_PIN_CREATION    = "Pin Created"
	NOTIFICATION_TITLE_RIDE_BOOKED     = "Ride Booked"
	NOTIFICATION_TITLE_SMS_TO_SERVICE  = "SathSawari sent an OTP"
	NOTIFICATION_TITLE_SHIFT_CREATED   = "Shift Scheduled"
	NOTIFICATION_TITLE_SHIFT_UPDATED   = "Shift Updated"
	NOTIFICATION_TITLE_SHIFT_CANCELLED = "Shift Cancelled"
	NOTIFICATION_TITLE_GROUP_REQUEST   = "New Group Request"
	NOTIFICATION_TITLE_GROUP_DECISION  = "Group Request Update"
)

// notification message
const (
	NOTIFICATION_MESSAGE_RIDE_CREATION      = "You have created a ride"
	NOTIFICATION_MESSAGE_PIN_CREATION       = "Please remember your secure PIN “%s” for future operations in the app"
	NOTIFICATION_MESSSGE_RIDE_BOOKED_DRIVER = "%v seat(s) booked for vehicle %s."
	NOTIFICATION_MESSAGE_PIN_UPDATED        = "Your pin has been updated, Please do not share it with anyone"
	NOTIFICATION_MESSAGE_SMS_TO_SERVICE     = "<#> Sathsawari verification code: %s. Do not share this code with anyone.\n%s"

	// shift notifications, they always carry the driver's number and the number of
	// the manager or sub manager who built the shift so riders know who to contact
	NOTIFICATION_MESSAGE_SHIFT_PASSENGER = "Your %s shift on %s. Vehicle %s, you are picked from %s at %s. Driver %s (%s). For any issue contact %s (%s)."
	NOTIFICATION_MESSAGE_SHIFT_DRIVER    = "Your %s shift on %s. Vehicle %s with %d passenger(s), %d stop(s), starting %s. For any issue contact %s (%s)."
	NOTIFICATION_MESSAGE_SHIFT_CANCELLED = "Your %s shift on %s with vehicle %s has been cancelled. For any issue contact %s (%s)."

	// group membership notifications
	NOTIFICATION_MESSAGE_GROUP_JOIN_REQUEST = "%s has requested to join your group %s"
	NOTIFICATION_MESSAGE_GROUP_APPROVED     = "Your request to join the group %s has been approved"
	NOTIFICATION_MESSAGE_GROUP_REJECTED     = "Your request to join the group %s has been rejected"
	NOTIFICATION_MESSAGE_GROUP_REMOVED      = "You have been removed from the group %s"
	NOTIFICATION_MESSAGE_SUBMANAGER_ADDED   = "You are now a sub manager of the group %s"
	NOTIFICATION_MESSAGE_SUBMANAGER_REMOVED = "You are no longer a sub manager of the group %s"
	NOTIFICATION_MESSAGE_GROUP_LEFT         = "Someone has left your group %s"
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

	// group and shift limits
	Group_Name_Min_Len                  = 3
	Group_Name_Max_Len                  = 100
	Description_Max_Len                 = 500
	Role_Name_Min_Len                   = 3
	Role_Name_Max_Len                   = 50
	Permission_Code_Min_Len             = 3
	Permission_Code_Max_Len             = 100
	Location_Min_Len                    = 1
	Location_Max_Len                    = 200
	Shift_Stops_Max_Len                 = 60
	Bulk_Request_Max_Len                = 100
	Day_Of_Week_Min_Value               = 1
	Day_Of_Week_Max_Value               = 7
	Shift_Cancel_Min_Hours_Before_Start = 2
)

const (
	DateTimeLayout = "2006-01-02 15:04:05"
)

const (
	DRIVER_TOKEN    = "driver_token"
	ADMIN_TOKEN     = "admin_token"
	PASSENGER_TOKEN = "passenger_token"
	OPEN_TOKEN      = "open_token"
	NO_TOKEN_TYPE   = "no_token_type" // such token whose type dont matter
)

// shift direction, a pickup and a drop off are always two separate shifts
const (
	Shift_Direction_Pickup = "pickup"
	Shift_Direction_Drop   = "drop"
)

// group membership status, shared by driver, vehicle and passenger membership
const (
	Membership_Status_Pending  = "pending"
	Membership_Status_Approved = "approved"
	Membership_Status_Rejected = "rejected"
	Membership_Status_Left     = "left"
	Membership_Status_Removed  = "removed"
)

// how a driver joins a group
const (
	Join_Type_Driver_Only  = "driver_only"
	Join_Type_Vehicle_Only = "vehicle_only"
	Join_Type_Both         = "both"
)

// member kinds and actions used by the bulk membership decision api
const (
	Member_Type_Driver    = "driver"
	Member_Type_Vehicle   = "vehicle"
	Member_Type_Passenger = "passenger"

	Membership_Action_Approve = "approve"
	Membership_Action_Reject  = "reject"
	Membership_Action_Remove  = "remove"
)

// shift seat status
const (
	Seat_Status_Empty    = "empty"
	Seat_Status_Assigned = "assigned"
)

// system roles seeded at boot, admin may add more
const (
	ROLE_GROUP_OWNER      = "group_owner"
	ROLE_GROUP_SUBMANAGER = "group_submanager"
)

// system permissions seeded at boot, admin may add more and remap them to roles
const (
	PERMISSION_GROUP_MANAGE_MEMBERS     = "group.manage_members"
	PERMISSION_GROUP_MANAGE_VEHICLES    = "group.manage_vehicles"
	PERMISSION_GROUP_MANAGE_SUBMANAGERS = "group.manage_submanagers"
	PERMISSION_GROUP_DELETE             = "group.delete"
	PERMISSION_SHIFT_CREATE             = "shift.create"
	PERMISSION_SHIFT_ASSIGN_SEATS       = "shift.assign_seats"
	PERMISSION_SHIFT_CANCEL             = "shift.cancel"
	PERMISSION_SHIFT_MANAGE_TEMPLATES   = "shift.manage_templates"
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
	NOTIFICATION_KEY_RIDE_ID  = "rideId"
	NOTIFICATION_KEY_SHIFT_ID = "shiftId"
	NOTIFICATION_KEY_GROUP_ID = "groupId"
)

// deep linking url types
const (
	LIKE_TYPE_RIDE_URL         = "RIDE_URL"
	LIKE_TYPE_RIDE_REQUEST_URL = "RIDE_REQUEST_URL"
)
