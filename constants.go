package boa

type ParamType int

const (
	TypeBool ParamType = iota
	TypeString
	TypeStringSlice
	TypeInt
	TypeIntSlice
	TypeFloat
	TypeFloatSlice
	TypeTime
	TypeTimeDuration
	TypeDate
	TypeDateRange
)
