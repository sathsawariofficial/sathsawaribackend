package constants

// response message
const (
	// SUCCESS
	Registered_Successfully       = "%v registered successfully"
	Updated_Successfully          = "%v updated successfully"
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
)

// keys
const (
	Sessoin_KEY            = "session"
	User_KEY               = "user_id"
	Vehicle_KEY            = "vehicle_id"
	Pin_Key                = "pin"
	Encrypted_User_KEY     = "encrypted_user_id"
	Page_Key               = "page"
	Start_Time_Key         = "start_time"
	Extimated_End_Time_Key = "estimated_end_time"
	Start_Loc_Key          = "start_location"
	End_Loc_Key            = "end_location"
	Status_Key             = "status"
	Ride_Key               = "ride_id"
	Seat_Key               = "seat_id"
	Ride_Template_Key      = "ride_template_id"
	Language_Key           = "lang"
)

// User Types
const (
	User_Driver    = 1
	User_Passenger = 2
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
	FIREBASE_CREDENTIALS_JSON = "FIREBASE_CREDENTIALS_JSON"
)

// limits
const (
	Name_Min_Len              = 3
	Name_Max_Len              = 128
	Username_Min_Len          = 5
	Username_Max_Len          = 50
	Password_Min_Len          = 8
	Password_Max_Len          = 12
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
	RouteDetails_Min_Len      = 1
	RouteDetails_Max_Len      = 200
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
)

const (
	DateTimeLayout = "2006-01-02 15:04:05"
)

const (
	DRIVER_TOKEN  = "driver_token"
	ADMIN_TOKEN   = "admin_token"
	OPEN_TOKEN    = "open_token"
	NO_TOKEN_TYPE = "no_token_type" // such token whose type dont matter
)
