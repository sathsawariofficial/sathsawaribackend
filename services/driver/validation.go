package driver

import (
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/utils"
)

func ValidateDriverRegistration(request *DriverRegistrationRequest) error {
	mobileLen := len(request.MobileNumber)
	nameLen := len(request.Name)
	passwordLen := len(request.Password)

	if !(mobileLen >= constants.MobileNumber_Min_Len && mobileLen <= constants.MobileNumber_Max_Len) {
		return fmt.Errorf("length of the mobile number should be between %v and %v characters", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}
	if !(nameLen >= constants.Name_Min_Len && nameLen <= constants.Name_Max_Len) {
		return fmt.Errorf("length of the name should be between %v and %v characters", constants.Name_Min_Len, constants.Name_Max_Len)
	}
	if !(passwordLen >= constants.Password_Min_Len && passwordLen <= constants.Password_Max_Len) {
		return fmt.Errorf("length of the password should be between %v and %v characters", constants.Password_Min_Len, constants.Password_Max_Len)
	}

	if !request.Gender.IsValid() {
		return fmt.Errorf(constants.Invalid_Data, "gender")
	}

	if !constants.Mobile_Regex.MatchString(request.MobileNumber) {
		return fmt.Errorf("invalid mobile number %v", request.MobileNumber)
	}
	if !utils.IsValidPassword(request.Password) {
		return fmt.Errorf("invalid password, should contain atleast one uppercase, lowercase letter, number and special character and length between %d and %d", constants.Password_Min_Len, constants.Password_Max_Len)
	}
	if utils.IsStringEmpty(request.DeviceId) {
		return fmt.Errorf(constants.Unable_To_Do_Job, "register at the moment, Please try again later.")
	}

	return nil
}

func ValidatePin(pin string) error {
	pinLen := len(pin)

	if pinLen != constants.Pin_Len {
		return fmt.Errorf("length of the pin should be %v", constants.Pin_Len)
	}

	return nil
}

func ValidateVehicleRegistration(request *VehicleRegistrationRequest) error {
	vehicleNumberLen := len(request.VehicleNumber)
	vehicleInfoLen := len(request.VehicleInfo)

	if !(vehicleNumberLen >= constants.VehicleNumber_Min_Len && vehicleNumberLen <= constants.VehicleNumber_Max_Len) {
		return fmt.Errorf("length of the vehicle number should be between %v and %v characters", constants.VehicleNumber_Min_Len, constants.VehicleNumber_Max_Len)
	}
	if !(vehicleInfoLen >= constants.VehicleInfo_Min_Len && vehicleInfoLen <= constants.VehicleInfo_Max_Len) {
		return fmt.Errorf("length of the vehicle information should be between %v and %v characters", constants.VehicleInfo_Min_Len, constants.VehicleInfo_Max_Len)
	}

	return ValidatePin(request.Pin)
}

func ValidateVehicleUpdate(request *VehicleUpdateRequest) error {
	vehicleNumberLen := len(request.VehicleNumber)
	vehicleInfoLen := len(request.VehicleInfo)

	if utils.IsStringEmpty(request.VehicleId) {
		return fmt.Errorf("vehicle id not found")
	}
	if !utils.IsStringEmpty(request.VehicleNumber) && !(vehicleNumberLen >= constants.VehicleNumber_Min_Len && vehicleNumberLen <= constants.VehicleNumber_Max_Len) {
		return fmt.Errorf("length of the vehicle number should be between %v and %v characters", constants.VehicleNumber_Min_Len, constants.VehicleNumber_Max_Len)
	}
	if !utils.IsStringEmpty(request.VehicleInfo) && !(vehicleInfoLen >= constants.VehicleInfo_Min_Len && vehicleInfoLen <= constants.VehicleInfo_Max_Len) {
		return fmt.Errorf("length of the vehicle information should be between %v and %v characters", constants.VehicleInfo_Min_Len, constants.VehicleInfo_Max_Len)
	}
	if !utils.IsStringEmpty(request.Status) &&
		(request.Status != constants.Status_InActive &&
			request.Status != constants.Status_Active) {
		return fmt.Errorf(constants.Invalid_Data, "status")
	}

	return ValidatePin(request.Pin)
}

func ValidateDriverLogin(request *DriverLoginRequest) error {
	var errMessage string
	if utils.IsStringEmptyWithKey(request.MobileNumber, "MobileNumber", &errMessage) ||
		utils.IsStringEmptyWithKey(request.Password, "Password", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	mobileLen := len(request.MobileNumber)
	passwordLen := len(request.Password)

	if !(mobileLen >= constants.MobileNumber_Min_Len && mobileLen <= constants.MobileNumber_Max_Len) {
		return fmt.Errorf("length of the mobile number should be between %v and %v characters", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}
	if !(passwordLen >= constants.Password_Min_Len && passwordLen <= constants.Password_Max_Len) {
		return fmt.Errorf("length of the password should be between %v and %v characters", constants.Password_Min_Len, constants.Password_Max_Len)
	}

	if !constants.Mobile_Regex.MatchString(request.MobileNumber) {
		return fmt.Errorf("invalid mobile number %v", request.MobileNumber)
	}
	if !utils.IsValidPassword(request.Password) {
		return fmt.Errorf("invalid password, should contain atleast one uppercase, lowercase letter, number and special character and length between %d and %d", constants.Password_Min_Len, constants.Password_Max_Len)
	}
	if utils.IsStringEmpty(request.DeviceId) {
		return fmt.Errorf(constants.Unable_To_Do_Job, "login at the moment, Please try again later.")
	}

	return nil
}

func ValidateRateDriver(request *RateDriverRequest) error {
	var errMessage string
	if utils.IsStringEmptyWithKey(request.MobileNumber, "MobileNumber", &errMessage) ||
		utils.IsStringEmptyWithKey(request.DriverId, "DriverId", &errMessage) {
		return fmt.Errorf(constants.Missing_Data, errMessage)
	}

	mobileLen := len(request.MobileNumber)

	if !(mobileLen > constants.MobileNumber_Min_Len && mobileLen <= constants.MobileNumber_Max_Len) {
		return fmt.Errorf("length of the mobile number should be between %v and %v characters", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}
	if !(request.Rating > constants.Rating_Min_Value && request.Rating <= constants.Rating_Max_Value) {
		return fmt.Errorf("rating should be between %v and %v characters", constants.Rating_Min_Value, constants.Rating_Max_Value)
	}

	if !constants.Mobile_Regex.MatchString(request.MobileNumber) {
		return fmt.Errorf("invalid mobile number %v", request.MobileNumber)
	}

	return nil
}

func ValidateDriverProfileInfo(driverId string) error {
	if !utils.PKValidation(driverId) {
		return fmt.Errorf(constants.Invalid_Data, "driver id")
	}

	return nil
}

func ValidateUpdateProfileStatus(status string) error {
	if utils.IsStringEmpty(status) ||
		(status != constants.Status_InActive &&
			status != constants.Status_Active) {
		return fmt.Errorf(constants.Invalid_Data, "status")
	}

	return nil
}

func ValidateChangePassword(driverId string, request *ChangePasswordRequest) error {
	if !utils.PKValidation(driverId) {
		return fmt.Errorf(constants.Invalid_Data, "driver id")
	}

	if !utils.IsValidPassword(request.OldPassword) {
		return fmt.Errorf("invalid password, should contain atleast one uppercase, lowercase letter, number and special character and length between %d and %d", constants.Password_Min_Len, constants.Password_Max_Len)
	}

	if !utils.IsValidPassword(request.NewPassword) {
		return fmt.Errorf("invalid password, should contain atleast one uppercase, lowercase letter, number and special character and length between %d and %d", constants.Password_Min_Len, constants.Password_Max_Len)
	}

	return nil
}

func ValidateForgotPassword(mobileNumber string) error {
	mobileLen := len(mobileNumber)

	if !(mobileLen > constants.MobileNumber_Min_Len && mobileLen <= constants.MobileNumber_Max_Len) {
		return fmt.Errorf("length of the mobile number should be between %v and %v characters", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}

	return nil
}

func ValidateChangePin(driverId string, request *ChangePinRequest) error {
	if !utils.PKValidation(driverId) {
		return fmt.Errorf(constants.Invalid_Data, "driver id")
	}

	if err := ValidatePin(request.OldPin); err != nil {
		return err
	}

	if err := ValidatePin(request.NewPin); err != nil {
		return err
	}

	return nil
}

func ValidateForgotPin(mobileNumber string) error {
	mobileLen := len(mobileNumber)

	if !(mobileLen > constants.MobileNumber_Min_Len && mobileLen <= constants.MobileNumber_Max_Len) {
		return fmt.Errorf("length of the mobile number should be between %v and %v characters", constants.MobileNumber_Min_Len, constants.MobileNumber_Max_Len)
	}

	return nil
}

func ValidateBookSeat(sessionId string, request BookSeatRequest) error {
	if !utils.PKValidation(request.RideId) {
		return fmt.Errorf(constants.Invalid_Data, "ride id")
	}

	return nil
}
