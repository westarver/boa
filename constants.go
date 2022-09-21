package boa

type ParameterType int

const (
	TypeBool ParameterType = iota
	TypeString
	TypeStringSlice
	TypeInt
	TypeIntSlice
	TypeFloat
	TypeFloatSlice
	TypeTime
	TypeTimeSlice
	TypeTimeDuration
	TypeTimeDurationSlice
	TypeDate
	TypeDateSlice
	TypePath
	TypePathSlice
	TypeURL
	TypeURLSlice
	TypeIPv4
	TypeIPv4Slice
	TypeEmail
	TypeEmailSlice
	TypePhone
	TypePhoneSlice
)
